package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/pkg/encryption"
)

type OAuthRepository struct {
	db         *DB
	authSecret string
}

func NewOAuthRepository(db *DB, authSecret string) *OAuthRepository {
	return &OAuthRepository{
		db:         db,
		authSecret: authSecret,
	}
}

func (r *OAuthRepository) GetProvider(ctx context.Context, name domain.OAuthProviderType) (*domain.OAuthProvider, error) {
	query := `
		SELECT provider, client_id, client_secret, enabled, scopes, auth_url, token_url, user_info_url, created_at, updated_at
		FROM oauth_providers
		WHERE provider = $1
	`

	var p domain.OAuthProvider
	var scopes []string
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&p.Provider,
		&p.ClientID,
		&p.ClientSecret,
		&p.Enabled,
		&scopes,
		&p.AuthURL,
		&p.TokenURL,
		&p.UserInfoURL,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get oauth provider: %w", err)
	}

	// Decrypt secrets
	if p.ClientSecret != "" {
		decrypted, err := encryption.Decrypt(p.ClientSecret, r.authSecret)
		if err == nil {
			p.ClientSecret = decrypted
		}
		// If error, keep original (maybe not encrypted yet or legacy)
	}
	if p.ClientID != "" {
		decrypted, err := encryption.Decrypt(p.ClientID, r.authSecret)
		if err == nil {
			p.ClientID = decrypted
		}
	}

	p.Scopes = scopes
	return &p, nil
}

func (r *OAuthRepository) ListProviders(ctx context.Context) ([]*domain.OAuthProvider, error) {
	query := `
		SELECT provider, client_id, client_secret, enabled, scopes, auth_url, token_url, user_info_url, created_at, updated_at
		FROM oauth_providers
		ORDER BY provider
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth providers: %w", err)
	}
	defer rows.Close()

	var providers []*domain.OAuthProvider
	for rows.Next() {
		var p domain.OAuthProvider
		var scopes []string
		if err := rows.Scan(
			&p.Provider,
			&p.ClientID,
			&p.ClientSecret,
			&p.Enabled,
			&scopes,
			&p.AuthURL,
			&p.TokenURL,
			&p.UserInfoURL,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan oauth provider: %w", err)
		}

		// Decrypt secrets
		if p.ClientSecret != "" {
			decrypted, err := encryption.Decrypt(p.ClientSecret, r.authSecret)
			if err == nil {
				p.ClientSecret = decrypted
			}
		}
		if p.ClientID != "" {
			decrypted, err := encryption.Decrypt(p.ClientID, r.authSecret)
			if err == nil {
				p.ClientID = decrypted
			}
		}

		p.Scopes = scopes
		providers = append(providers, &p)
	}

	return providers, nil
}

func (r *OAuthRepository) UpdateProvider(ctx context.Context, provider *domain.OAuthProvider) error {
	query := `
		INSERT INTO oauth_providers (provider, client_id, client_secret, enabled, scopes, auth_url, token_url, user_info_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (provider) DO UPDATE SET
			client_id = EXCLUDED.client_id,
			client_secret = EXCLUDED.client_secret,
			enabled = EXCLUDED.enabled,
			scopes = EXCLUDED.scopes,
			auth_url = EXCLUDED.auth_url,
			token_url = EXCLUDED.token_url,
			user_info_url = EXCLUDED.user_info_url,
			updated_at = NOW()
	`

	// Encrypt sensitive fields
	// We operate on a copy or modify input? Modifying input is risky if caller reuses.
	// But usually it's fine. Let's create vars.
	encClientID, err := encryption.Encrypt(provider.ClientID, r.authSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt client id: %w", err)
	}

	encClientSecret, err := encryption.Encrypt(provider.ClientSecret, r.authSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt client secret: %w", err)
	}

	_, err = r.db.Pool.Exec(ctx, query,
		provider.Provider,
		encClientID,
		encClientSecret,
		provider.Enabled,
		provider.Scopes,
		provider.AuthURL,
		provider.TokenURL,
		provider.UserInfoURL,
	)

	if err != nil {
		return fmt.Errorf("failed to update oauth provider: %w", err)
	}

	return nil
}

func (r *OAuthRepository) CreateUserOAuth(ctx context.Context, userOAuth *domain.UserOAuth) error {
	query := `
		INSERT INTO user_oauths (user_id, provider, provider_user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	// Encrypt tokens
	encAccessToken, err := encryption.Encrypt(userOAuth.AccessToken, r.authSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encRefreshToken := ""
	if userOAuth.RefreshToken != "" {
		enc, err := encryption.Encrypt(userOAuth.RefreshToken, r.authSecret)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		encRefreshToken = enc
	}

	err = r.db.Pool.QueryRow(ctx, query,
		userOAuth.UserID,
		userOAuth.Provider,
		userOAuth.ProviderUserID,
		encAccessToken,
		encRefreshToken,
		userOAuth.ExpiresAt,
	).Scan(&userOAuth.ID, &userOAuth.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("failed to create user oauth: %w", err)
	}

	return nil
}

func (r *OAuthRepository) GetUserOAuth(ctx context.Context, provider domain.OAuthProviderType, providerUserID string) (*domain.UserOAuth, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at
		FROM user_oauths
		WHERE provider = $1 AND provider_user_id = $2
	`

	var u domain.UserOAuth
	err := r.db.Pool.QueryRow(ctx, query, provider, providerUserID).Scan(
		&u.ID,
		&u.UserID,
		&u.Provider,
		&u.ProviderUserID,
		&u.AccessToken,
		&u.RefreshToken,
		&u.ExpiresAt,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user oauth: %w", err)
	}

	// Decrypt tokens (silently ignore errors for robust reading? or error out?)
	// Let's decrypt if present.
	if u.AccessToken != "" {
		dec, err := encryption.Decrypt(u.AccessToken, r.authSecret)
		if err == nil {
			u.AccessToken = dec
		}
	}
	if u.RefreshToken != "" {
		dec, err := encryption.Decrypt(u.RefreshToken, r.authSecret)
		if err == nil {
			u.RefreshToken = dec
		}
	}

	return &u, nil
}

func (r *OAuthRepository) GetUserOAuthByUserID(ctx context.Context, userID uuid.UUID, provider domain.OAuthProviderType) (*domain.UserOAuth, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at
		FROM user_oauths
		WHERE user_id = $1 AND provider = $2
	`

	var u domain.UserOAuth
	err := r.db.Pool.QueryRow(ctx, query, userID, provider).Scan(
		&u.ID,
		&u.UserID,
		&u.Provider,
		&u.ProviderUserID,
		&u.AccessToken,
		&u.RefreshToken,
		&u.ExpiresAt,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user oauth by user id: %w", err)
	}

	if u.AccessToken != "" {
		dec, err := encryption.Decrypt(u.AccessToken, r.authSecret)
		if err == nil {
			u.AccessToken = dec
		}
	}
	if u.RefreshToken != "" {
		dec, err := encryption.Decrypt(u.RefreshToken, r.authSecret)
		if err == nil {
			u.RefreshToken = dec
		}
	}

	return &u, nil
}
