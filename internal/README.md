# Internal Packages

This directory contains private application code. Go prevents external packages from importing anything under `internal/`.

## Package Overview

```
internal/
├── config/       # Environment configuration loading
├── domain/       # Core business entities (pure data, no dependencies)
├── handler/      # HTTP handlers (parse requests, render responses)
├── service/      # Business logic (orchestrates repositories)
├── repository/   # Data access (database operations)
├── middleware/   # HTTP middleware (auth, logging, rate limiting)
├── storage/      # File storage abstraction (profile images)
└── templates/    # Template utilities
```

## Dependency Rules

Dependencies flow **inward** - outer layers depend on inner layers:

```
handler → service → repository → domain
                             ↘
                              → domain
```

- **domain**: No external dependencies
- **repository**: Depends on domain only
- **service**: Depends on repository interfaces and domain
- **handler**: Depends on service interfaces

## Adding New Code

1. **New entity?** → Add to `domain/`
2. **New database operation?** → Add to `repository/`
3. **New business logic?** → Add to `service/`
4. **New HTTP endpoint?** → Add to `handler/`
