# Getting Started

A quick guide to help developers from any background start contributing to this project.

## First Steps

1. **Read** [`ARCHITECTURE.md`](./ARCHITECTURE.md) for the overall structure
2. **Start** with `cmd/server/main.go` - see how everything connects
3. **Explore** the domain models in `internal/domain/`
4. **Trace** a feature by following a handler → service → repository

## Key Files to Understand

| File | What to Learn |
|------|---------------|
| `cmd/server/main.go` | Dependency injection, route registration, startup flow |
| `internal/handler/auth_handler.go` | How HTTP handlers work |
| `internal/service/auth_service.go` | Business logic patterns |
| `internal/repository/postgres/user_repo.go` | Database access patterns |
| `web/templ/layouts/base.templ` | Template structure and rendering |

## Common Development Tasks

### Running the App

```bash
# Start PostgreSQL (via Docker)
docker-compose up -d

# Run in development mode (hot reload)
npm run dev
```

### Adding a New Page

1. Create template: `web/templ/pages/mypage.templ`
2. Run `templ generate` (or let hot-reload handle it)
3. Add handler method in appropriate `*_handler.go`
4. Register route in `cmd/server/main.go`

### Adding a New API Endpoint

1. Add method to service interface in `internal/service/interfaces.go`
2. Implement in the service (`internal/service/*_service.go`)
3. Add handler method (`internal/handler/*_handler.go`)
4. Register route in `cmd/server/main.go`

### Database Changes

1. Add migration in `internal/repository/postgres/migrations/schema.sql`
2. Update domain entity in `internal/domain/`
3. Update repository methods

## Environment Setup

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Required variables:
- `DATABASE_URL` - PostgreSQL connection string
- `AUTH_SECRET` - Session encryption key (any random string)

Optional:
- `RESEND_API_KEY` - For email functionality
- `PROFILE_IMAGE_STORAGE` - `database` or `s3`

## Understanding the Tech Stack

| If you know... | Then this will feel familiar |
|----------------|------------------------------|
| **Express.js** | Handlers are like route handlers, middleware is the same concept |
| **Django/Flask** | Services are like business logic, repos are like ORM models |
| **Spring Boot** | Very similar layered architecture (Controller → Service → Repository) |
| **Ruby on Rails** | Less magic, more explicit wiring, but same MVC concept |
| **React/Vue** | Templ templates are like components but server-rendered |

## Debugging Tips

1. **Check logs** - Everything is logged to console
2. **Database issues** - Verify `DATABASE_URL` and run `docker-compose logs db`
3. **Template errors** - Run `templ generate` to see compilation errors
4. **Authentication** - Check if session cookie is being set/sent
