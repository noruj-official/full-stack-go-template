# Authentication Implementation Plan (`auth_todo.md`)

This document outlines the step-by-step plan to implement recommended authentication and authorization features. The work is divided into daily milestones to ensure steady progress and thorough testing.

> [!IMPORTANT]
> **Testing Strategy**: Each feature MUST be verified before moving to the next.
> - **Automated**: Run `go test ./...` after each major change.
> - **Manual**: Verify flows using the browser or API client (Postman/curl).

---

## Phase 1: Security Hardening (Days 1-3)

### Day 1: Session Security & Rate Limiting
**Goal**: Secure the login/signup endpoints and enhance session tracking.

- [x] **Step 1: Database Migration for Sessions**
    - [x] Add `ip_address` (VARCHAR), `user_agent` (VARCHAR), and `last_activity_at` (TIMESTAMP) to `sessions` table.
- [x] **Step 2: Update Session Creation Logic**
    - [x] Update `auth_service.go` to capture IP and User-Agent from context/request.
    - [x] Update `domain/session.go` struct.
- [x] **Step 3: Implement Rate Limiting Middleware**
    - [x] Create `internal/middleware/rate_limit.go`.
    - [x] Use `golang.org/x/time/rate`.
    - [x] Apply to `/signin`, `/signup` routes in `cmd/server/routes.go`.
    - [x] **Test**: Try to spam login and verify 429 Too Many Requests.
- [x] **Step 4: Session Security Enhancements**
    - [x] Implement Session ID regeneration on login (if not existing).
    - [x] Implement "Sign Out All Devices" (optional, but good to have prepared).

### Day 2: Email Verification
**Goal**: Ensure all registered users have valid email addresses.

- [x] **Step 1: Database Migration for Users**
    - [x] Add `email_verified` (BOOLEAN, default false) and `verification_token` (VARCHAR) to `users` table.
- [x] **Step 2: Update Registration Flow**
    - [x] Generate unique token on registration.
    - [x] Send email with link (Resend implemented).
- [x] **Step 3: Create Verification Endpoint**
    - [x] Add `GET /verify-email?token=...` handler.
    - [x] Verify token matches and expire it.
    - [x] Update user status to `email_verified = true`.
- [x] **Step 4: Middleware Check**
    - [x] Create `RequireVerifiedEmail` middleware.
    - [x] Protect sensitive routes (e.g., creating posts) with this middleware.
    - [x] **Test**: Register new user, try to access protected route (fail), click link, try again (pass).

### Day 3: Password Reset Flow
**Goal**: Allow users to recover lost passwords securely.

- [x] **Step 1: Create Password Reset Table**
    - [x] `password_reset_tokens` (id, user_id, token_hash, expires_at).
- [x] **Step 2: Request Password Reset Endpoint**
    - [x] `POST /forgot-password`.
    - [x] Generate token, hash it, store in DB.
    - [x] Send email with link.
- [x] **Step 3: Reset Password Endpoint**
    - [x] `POST /reset-password`.
    - [x] Verify token hash and expiry.
    - [x] Update user password with new hash.
    - [x] Invalidate token.
    - [x] **Test**: Full flow from "forgot password" to logging in with new password.

---

## Phase 2: Administrative & Feature Controls (Days 4-5)

### Day 4: Enhanced User Status (Ban/Suspend)
**Goal**: Give admins control over user access without deleting data.

- [ ] **Step 1: Database Migration**
    - [ ] Add `status` column to `users` (active, suspended, banned).
- [ ] **Step 2: Update Auth Middleware**
    - [ ] Check `user.Status` in `RequireAuth` middleware.
    - [ ] Deny access if not "active".
- [ ] **Step 3: Admin Actions**
    - [ ] Add endpoints for Admins to change user status.
    - [ ] **Test**: Admin suspends User A. User A is immediately logged out/blocked on next request.

### Day 5: Feature Flag System
**Goal**: Enable safe rollout of new features (like those in Phase 3).

- [ ] **Step 1: Create Feature Flags Table**
    - [ ] `feature_flags` (key, enabled, description).
- [ ] **Step 2: Feature Service**
    - [ ] Create `internal/service/feature_service.go` (IsEnabled check).
- [ ] **Step 3: Admin UI for Flags**
    - [ ] Simple page to toggle flags on/off (Admin only).
- [ ] **Step 4: Integration**
    - [ ] Add `Features` to template context or API responses.
    - [ ] **Test**: Add dummy flag, toggle it, verify UI changes.

---

## Phase 3: Future Growth (Optional - Day 6+)

### Day 6: OAuth2 (Social Login)
- [ ] **Step 1**: Set up `goth` or similar package.
- [ ] **Step 2**: Configure Google/GitHub providers.
- [ ] **Step 3**: Handle callbacks and account linking.
