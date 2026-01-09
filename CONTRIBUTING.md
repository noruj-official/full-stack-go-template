# Contributing to Full Stack Go Template

Thank you for your interest in contributing! This guide will help you get started with contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Issue Guidelines](#issue-guidelines)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

We are committed to fostering a welcoming community. Please read and follow our [Code of Conduct](./CODE_OF_CONDUCT.md) in all interactions.

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.23+** - [Install Go](https://go.dev/doc/install)
- **Node.js 18+** - [Install Node.js](https://nodejs.org/)
- **PostgreSQL 16** - Either locally or via Docker
- **Docker** (optional but recommended) - [Install Docker](https://www.docker.com/get-started)
- **Git** - [Install Git](https://git-scm.com/)

### Understanding the Project

Before contributing, familiarize yourself with:

1. **[README.md](./README.md)** - Project overview and features
2. **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - Detailed architecture documentation
3. **[GETTING_STARTED.md](./docs/GETTING_STARTED.md)** - Quick start guide

## Development Environment Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/full-stack-go-template.git
cd full-stack-go-template

# Add upstream remote
git remote add upstream https://github.com/noruj-official/full-stack-go-template.git
```

### 2. Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
# At minimum, set DATABASE_URL and AUTH_SECRET
```

### 3. Start the Database

```bash
# Using Docker (recommended)
docker-compose up -d

# Verify database is running
docker-compose ps
```

### 4. Install Dependencies

```bash
# Install Node.js dependencies
npm install

# Download Go modules
go mod download

# Install development tools
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air/v2@latest
```

### 5. Run the Application

```bash
# Start development server with hot-reload
npm run dev

# The app will be available at http://localhost:3000
```

## Project Structure

```
├── cmd/server/           # Application entry point
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── domain/          # Business entities
│   ├── handler/         # HTTP handlers (controllers)
│   ├── middleware/      # HTTP middleware
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   ├── storage/         # File storage service
│   └── templates/       # Template utilities
├── web/
│   ├── assets/          # CSS and vendor files
│   └── templ/           # Templ templates
│       ├── components/  # Reusable UI components
│       ├── layouts/     # Page layouts
│       └── pages/       # Page templates
├── docs/                # Documentation
└── scripts/             # Build and development scripts
```

## Development Workflow

### Creating a New Feature

1. **Create a feature branch**

```bash
git checkout -b feature/your-feature-name
```

2. **Follow the layered architecture**

When adding a new feature, follow this order:

- **Domain Entity** (`internal/domain/`) - Define your data models
- **Repository Interface** (`internal/repository/interfaces.go`) - Define data operations
- **Repository Implementation** (`internal/repository/postgres/`) - Implement data access
- **Service Interface** (`internal/service/interfaces.go`) - Define business logic contract
- **Service Implementation** (`internal/service/`) - Implement business logic
- **Handler** (`internal/handler/`) - Implement HTTP endpoints
- **Templates** (`web/templ/`) - Create UI templates
- **Routes** (`cmd/server/main.go`) - Wire everything together

3. **Example: Adding a Blog Feature**

```go
// 1. Domain entity - internal/domain/post.go
type Post struct {
    ID        uuid.UUID
    Title     string
    Content   string
    AuthorID  uuid.UUID
    CreatedAt time.Time
}

// 2. Repository interface - internal/repository/interfaces.go
type PostRepository interface {
    Create(ctx context.Context, post *domain.Post) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
}

// 3. Service interface - internal/service/interfaces.go
type PostService interface {
    CreatePost(ctx context.Context, post *domain.Post) error
    GetPost(ctx context.Context, id uuid.UUID) (*domain.Post, error)
}

// 4. Wire in main.go
postRepo := postgres.NewPostRepository(db)
postService := service.NewPostService(postRepo)
postHandler := handler.NewPostHandler(postService)
```

### Database Migrations

All database migrations are stored in `internal/repository/postgres/migrations/schema.sql`.

To add a migration:

1. Edit `schema.sql` to add your new tables/columns
2. Delete the database and restart to apply migrations (development only)

```bash
# Reset database (development only)
docker-compose down -v
docker-compose up -d

# Restart app to run migrations
npm run dev
```

### Frontend Development

#### Templ Templates

Templ is a type-safe templating engine for Go. Learn more at [templ.guide](https://templ.guide/).

```go
// web/templ/components/card.templ
package components

templ Card(title string, content string) {
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <h2 class="card-title">{ title }</h2>
            <p>{ content }</p>
        </div>
    </div>
}
```

After creating or editing `.templ` files, they are automatically compiled to Go code by the dev server.

#### Styling with Tailwind CSS

- CSS source: `web/assets/css/app.css`
- Built CSS: `web/assets/css/output.css` (auto-generated)
- Theme: DaisyUI v5 with `corporate` (light) and `night` (dark) themes

```css
/* Add custom styles in app.css */
@import "tailwindcss";

.custom-class {
    @apply text-primary hover:text-primary-focus;
}
```

#### Adding JavaScript Interactivity

We use **HTMX** for dynamic HTML updates and **Alpine.js** for reactivity.

```html
<!-- HTMX example -->
<button hx-post="/api/like" hx-swap="outerHTML">
    Like
</button>

<!-- Alpine.js example -->
<div x-data="{ open: false }">
    <button @click="open = !open">Toggle</button>
    <div x-show="open">Content</div>
</div>
```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` to format code (automatic with most editors)
- Write clear, descriptive variable and function names
- Add comments for exported functions and complex logic

```go
// Good: Descriptive function name with clear purpose
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
    if email == "" {
        return nil, domain.ErrInvalidEmail
    }
    return s.userRepo.GetByEmail(ctx, email)
}

// Bad: Unclear naming
func (s *userService) Get(ctx context.Context, e string) (*domain.User, error) {
    return s.userRepo.GetByEmail(ctx, e)
}
```

### Error Handling

- Always check and handle errors
- Use domain-specific errors defined in `internal/domain/errors.go`
- Provide context when wrapping errors

```go
// Good: Proper error handling with context
user, err := s.userRepo.GetByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get user %s: %w", id, err)
}

// Bad: Ignoring errors
user, _ := s.userRepo.GetByID(ctx, id)
```

### Context Usage

- Always pass `context.Context` as the first parameter
- Use context for cancellation, timeouts, and request-scoped values

```go
func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, error) {
    // Use ctx in all downstream calls
    user, err := s.userRepo.GetByEmail(ctx, email)
    // ...
}
```

### File Naming Conventions

| Pattern | Example | Purpose |
|---------|---------|---------|
| `*_handler.go` | `auth_handler.go` | HTTP handlers |
| `*_service.go` | `user_service.go` | Business logic |
| `*_repo.go` | `session_repo.go` | Repository implementations |
| `*.templ` | `navbar.templ` | Templ templates (source) |
| `*_templ.go` | `navbar_templ.go` | Generated Go code (don't edit) |

## Testing Guidelines

### Writing Tests

```go
// internal/service/user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &mockUserRepository{}
    service := NewUserService(mockRepo)
    
    user := &domain.User{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    // Act
    err := service.CreateUser(context.Background(), user)
    
    // Assert
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/service/...
```

## Submitting Changes

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```bash
git commit -m "feat(auth): add password reset functionality"
git commit -m "fix(user): resolve email validation bug"
git commit -m "docs: update contributing guidelines"
git commit -m "refactor(handler): simplify error handling logic"
```

### Keeping Your Fork Updated

```bash
# Fetch upstream changes
git fetch upstream

# Merge upstream changes into your local main
git checkout main
git merge upstream/main

# Update your fork on GitHub
git push origin main
```

## Issue Guidelines

### Before Creating an Issue

1. Search existing issues to avoid duplicates
2. Check if the issue exists in the latest version
3. Gather relevant information (logs, screenshots, etc.)

### Creating a Good Issue

Include:
- **Clear title** - Briefly describe the issue
- **Description** - What happened, what you expected
- **Steps to reproduce** - How to recreate the issue
- **Environment** - OS, Go version, browser (if applicable)
- **Screenshots** - If relevant

**Example:**

```markdown
## Bug Report

**Title:** Session cookie not persisting after login

**Description:**
After successful login, the session cookie is not being stored, causing 
the user to be redirected back to the login page.

**Steps to Reproduce:**
1. Navigate to /signin
2. Enter valid credentials
3. Submit the form
4. Observe redirect to /signin instead of /dashboard

**Environment:**
- OS: Windows 11
- Go: 1.23
- Browser: Chrome 120

**Screenshots:**
[Attach screenshot of browser DevTools showing missing cookie]
```

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest upstream changes
2. **Run tests** to ensure nothing is broken
3. **Format code** using `gofmt`
4. **Test manually** to verify your changes work

```bash
# Update from upstream
git fetch upstream
git rebase upstream/main

# Run tests
go test ./...

# Format code
go fmt ./...

# Build to check for errors
npm run build
```

### Creating a Pull Request

1. **Push your branch** to your fork

```bash
git push origin feature/your-feature-name
```

2. **Open a Pull Request** on GitHub from your fork to the upstream repository

3. **Fill out the PR template** with:
   - Description of changes
   - Related issue number (if applicable)
   - Testing performed
   - Screenshots (for UI changes)

### PR Title Format

Use the same format as commit messages:

```
feat(auth): add password reset functionality
fix(dashboard): resolve loading spinner issue
```

### Review Process

- Maintainers will review your PR
- Address any requested changes
- Once approved, your PR will be merged

### After Your PR is Merged

```bash
# Switch to main branch
git checkout main

# Pull the latest changes
git pull upstream main

# Delete your feature branch
git branch -d feature/your-feature-name

# Delete remote branch
git push origin --delete feature/your-feature-name
```

## Additional Resources

### Project Documentation

- [README.md](./README.md) - Project overview
- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - Architecture guide
- [GETTING_STARTED.md](./docs/GETTING_STARTED.md) - Quick start guide

### External Resources

- [Go Documentation](https://go.dev/doc/) - Official Go docs
- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
- [Templ Guide](https://templ.guide/) - Templ templating
- [HTMX Documentation](https://htmx.org/docs/) - HTMX reference
- [Tailwind CSS](https://tailwindcss.com/) - CSS framework
- [DaisyUI](https://daisyui.com/) - Component library

## Getting Help

If you need help:

1. Check the [documentation](./docs/)
2. Search [existing issues](https://github.com/noruj-official/full-stack-go-template/issues)
3. Create a new issue with the `question` label

## License

By contributing, you agree that your contributions will be licensed under the same [MIT License](./LICENSE) that covers this project.

---

Thank you for contributing to Full Stack Go Template! 🎉
