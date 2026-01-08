# Architecture Guide

This document explains the project architecture for developers coming from any programming background.

## Quick Reference for Developers from Other Languages

If you're coming from Node.js, Python, Java, or other languages, this table maps Go concepts to familiar patterns:

| Go Concept | Node.js Equivalent | Python Equivalent | Java Equivalent |
|------------|-------------------|-------------------|-----------------|
| `cmd/server/main.go` | `app.js` / `index.js` | `app.py` / `main.py` | `Application.java` |
| `internal/handler/` | Express routes/controllers | Flask/FastAPI routes | `@Controller` classes |
| `internal/service/` | Service classes | Service layer | `@Service` classes |
| `internal/repository/` | Database models/DAO | SQLAlchemy models | `@Repository` classes |
| `internal/domain/` | TypeScript interfaces | Pydantic models | Entity classes |
| `internal/middleware/` | Express middleware | Flask middleware | Filters/Interceptors |
| `web/templ/` | React components | Jinja2 templates | Thymeleaf templates |
| `go.mod` | `package.json` | `requirements.txt` | `pom.xml` / `build.gradle` |
| `internal/` | N/A (convention) | N/A | `impl` packages |

## Project Structure

```
go-starter/
├── cmd/                          # Application entry points
│   └── server/
│       └── main.go              # 🚀 START HERE - Bootstraps the app
│
├── internal/                     # Private application code (Go enforced)
│   ├── config/                  # Configuration management
│   │   └── config.go            # Loads env vars into typed config
│   │
│   ├── domain/                  # 📦 Core business entities
│   │   ├── user.go              # User entity & validation rules
│   │   ├── session.go           # Session entity
│   │   ├── role.go              # Role definitions (User/Admin/SuperAdmin)
│   │   ├── activity_log.go      # Activity tracking entity
│   │   ├── password_reset.go    # Password reset token entity
│   │   └── errors.go            # Domain-specific errors
│   │
│   ├── handler/                 # 🌐 HTTP handlers (Controllers)
│   │   ├── handler.go           # Base handler with shared utilities
│   │   ├── auth_handler.go      # Sign in, sign up, logout
│   │   ├── user_handler.go      # User CRUD operations
│   │   ├── home_handler.go      # Home & dashboard pages
│   │   ├── profile_handler.go   # User profile management
│   │   ├── settings_handler.go  # User settings
│   │   ├── activity_handler.go  # Activity log display
│   │   ├── analytics_handler.go # Admin analytics
│   │   └── audit_handler.go     # Audit log & system health
│   │
│   ├── service/                 # 💼 Business logic layer
│   │   ├── interfaces.go        # Service interfaces (contracts)
│   │   ├── auth_service.go      # Authentication logic
│   │   ├── user_service.go      # User management logic
│   │   ├── activity_service.go  # Activity logging logic
│   │   └── email_service.go     # Email sending (Resend)
│   │
│   ├── repository/              # 💾 Data access layer
│   │   ├── interfaces.go        # Repository interfaces
│   │   └── postgres/            # PostgreSQL implementations
│   │       ├── migrations/      # 📄 Database schema SQL files
│   │       ├── db.go            # Database connection & migrations
│   │       ├── queries.go       # SQL file loader (uses Go embed)
│   │       ├── user_repo.go     # User CRUD operations
│   │       ├── session_repo.go  # Session management
│   │       └── activity_repo.go # Activity log storage
│   │
│   ├── middleware/              # 🔒 HTTP middleware
│   │   ├── auth.go              # Authentication & authorization
│   │   ├── cors.go              # CORS configuration
│   │   ├── logging.go           # Request logging
│   │   ├── rate_limit.go        # Rate limiting
│   │   └── recovery.go          # Panic recovery
│   │
│   ├── storage/                 # 🖼️ Profile image storage service
│   │   ├── service.go           # Storage interface (database or S3)
│   │   ├── database.go          # PostgreSQL storage implementation
│   │   └── factory.go           # Creates storage based on config
│   │
│   └── templates/               # Template utilities
│
├── web/                         # Frontend assets
│   ├── assets/
│   │   ├── css/                 # Tailwind CSS source
│   │   └── vendor/              # Third-party JS (htmx, alpine)
│   └── templ/                   # 🎨 Templ templates
│       ├── components/          # Reusable UI components
│       │   ├── navbar.templ     # Navigation bar
│       │   ├── sidebar.templ    # Admin sidebar
│       │   └── footer.templ     # Page footer
│       ├── layouts/             # Page layouts
│       │   ├── base.templ       # Main layout (logged in users)
│       │   └── auth.templ       # Auth pages layout
│       └── pages/               # Page templates
│           ├── home.templ       # Landing page
│           ├── auth/            # Sign in, sign up, etc.
│           ├── dashboards/      # Role-based dashboards
│           ├── users/           # User management pages
│           ├── profile/         # User profile pages
│           └── admin/           # Admin-only pages
│
└── docs/                        # Documentation
```

## Request Flow

Here's how a typical HTTP request flows through the application:

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Request                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Middleware Stack                             │
│  CORS → Recovery → Logging → Auth                               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Handler                                   │
│  • Parses request (form data, JSON, URL params)                 │
│  • Validates input                                               │
│  • Calls service layer                                           │
│  • Renders response (HTML template or JSON)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service                                   │
│  • Business logic                                                │
│  • Orchestrates repository calls                                 │
│  • Handles transactions                                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Repository                                 │
│  • Database operations (CRUD)                                    │
│  • SQL queries                                                   │
│  • Data mapping                                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       PostgreSQL                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Key Patterns Explained

### 1. Dependency Injection

Dependencies are created in `main.go` and passed down through constructors:

```go
// main.go - Dependencies flow downward
userRepo := postgres.NewUserRepository(db)       // Create repo
userService := service.NewUserService(userRepo)  // Inject into service
userHandler := handler.NewUserHandler(userService) // Inject into handler
```

**Why?** Makes testing easy—swap real implementations for mocks.

### 2. Interface-Based Design

Interfaces are defined where they're used, not where implemented:

```go
// internal/service/interfaces.go
type UserService interface {
    GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
    // ...
}
```

**Why?** Loose coupling—handlers depend on interfaces, not concrete types.

### 3. Context for Request Scope

Every function takes `context.Context` as first parameter:

```go
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
```

**Why?** Carries request-scoped data (user, timeout, trace ID) through the stack.

### 4. Domain Entities

Pure data structures without database or HTTP coupling:

```go
// internal/domain/user.go
type User struct {
    ID        uuid.UUID
    Email     string
    Name      string
    Role      Role
    CreatedAt time.Time
}
```

**Why?** Business rules stay independent of infrastructure choices.

## Adding a New Feature

Here's how to add a new resource (e.g., "Posts"):

1. **Create domain entity** → `internal/domain/post.go`
2. **Add repository interface** → `internal/repository/interfaces.go`
3. **Implement repository** → `internal/repository/postgres/post_repo.go`
4. **Add service interface** → `internal/service/interfaces.go`
5. **Implement service** → `internal/service/post_service.go`
6. **Create handler** → `internal/handler/post_handler.go`
7. **Add templates** → `web/templ/pages/posts/`
8. **Wire everything** → `cmd/server/main.go`
9. **Add routes** → `cmd/server/main.go`

## Technologies at a Glance

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.23 | Fast, simple, strongly typed |
| **Database** | PostgreSQL | Reliable relational database |
| **Templates** | Templ | Type-safe HTML templating |
| **CSS** | Tailwind CSS v4 | Utility-first styling |
| **Components** | DaisyUI v5 | Pre-built UI components |
| **Interactivity** | HTMX + Alpine.js | Dynamic UX without heavy JS |
| **Icons** | Lucide | Beautiful icon set |

## File Naming Conventions

| Pattern | Example | Purpose |
|---------|---------|---------|
| `*_handler.go` | `auth_handler.go` | HTTP request handlers |
| `*_service.go` | `auth_service.go` | Business logic |
| `*_repo.go` | `user_repo.go` | Repository implementations |
| `*.templ` | `home.templ` | Templ template sources |
| `*_templ.go` | `home_templ.go` | Generated Go code (don't edit) |
| `interfaces.go` | `service/interfaces.go` | Interface definitions |

## Further Reading

- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Architecture principles
- [HTMX Documentation](https://htmx.org/docs/) - Frontend interactivity
- [Templ Documentation](https://templ.guide/) - Template syntax
