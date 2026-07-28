# Template Monorepo

A full-stack monorepo with a React frontend and a Go API backed by PostgreSQL.

## Stack

- **Frontend Runtime & Package Manager**: [Bun](https://bun.sh/)
- **Backend Runtime**: [Go](https://go.dev/)
- **Build System**: [Nx](https://nx.dev/)
- **Database**: PostgreSQL via [pgx](https://github.com/jackc/pgx)
- **UI**: [shadcn/ui](https://ui.shadcn.com/) with Tailwind CSS v4

## Apps

- `apps/web` - [React 19](https://react.dev/) frontend with [Vite](https://vitejs.dev/) and the [React Compiler](https://react.dev/learn/react-compiler)
- `apps/apis` - Go HTTP API with authentication

## Libraries

- `libs/ui` - Shared shadcn components, utilities, and Tailwind theme

## Getting Started

### Prerequisites

- Bun >= 1.1.0
- Go >= 1.24
- PostgreSQL database

### Installation

```bash
bun install
```

### Environment Setup

Copy the root example environment file and fill in your values:

```bash
cp .env.example .env
```

Update `.env` with your database connection string and other settings. This single file is used by all apps and libraries.

The database library applies its idempotent schema from `libs/database/schema.sql` when the API starts.

### Development

```bash
# PostgreSQL + API + web with live development processes
just dev

# PostgreSQL + production containers
just docker
```

Both commands run in the foreground and expose the web app at `http://localhost:3000` and API at `http://localhost:3001`. `just docker` reads `.env` when present and otherwise uses `postgres` as its local database password.

To stop detached services, use `devenv processes down` or `docker compose -f docker/docker-compose.yml down` respectively.

Individual processes remain available when needed:

```bash
# Run all development scripts without PostgreSQL
bun dev

# Run specific app
bun --filter @template/web dev
bun --filter @template/apis dev

# Or run the API directly
cd apps/apis && go run .
```

### Devenv

With [devenv](https://devenv.sh/) installed, one command provisions Go 1.26 and Bun, initializes a persistent PostgreSQL 18 database, and starts PostgreSQL, the API, and the web app in readiness order:

```bash
devenv up
```

PostgreSQL data is kept in devenv's project state. The API and web app remain available at `http://localhost:3001` and `http://localhost:3000`; the API automatically reloads when Go or schema files change.

### Build

```bash
bun run build
```

TypeScript projects are checked with the pinned TypeScript 7 native preview:

```bash
bun run typecheck
```

### Lint & Format

```bash
# Check all files
bun check

# Format all files
bun format
```

## UI Components

Both `apps/web/components.json` and `libs/ui/components.json` use shadcn's monorepo routing. Run component commands against the app; reusable UI is written to `libs/ui` automatically:

```bash
bunx shadcn@latest add dialog -c apps/web
```

Import shared components through package exports:

```tsx
import { Button } from "@template/ui/components/button";
```

The shared Tailwind v4 theme lives at `libs/ui/src/styles/globals.css` and is imported by the web entrypoint.

## Auth

Authentication is implemented by the Go API with the following features:

- Bcrypt password hashing
- Opaque, hashed session tokens stored in PostgreSQL
- HttpOnly, SameSite session cookies
- Trusted-origin CORS enforcement

### Auth Routes (API)

- `POST /api/auth/sign-in/email` - Sign in with email/password
- `POST /api/auth/sign-up/email` - Register with email/password
- `POST /api/auth/sign-out` - Sign out
- `GET /api/auth/session` - Get the current session

### Protected Routes

- `GET /api/me` - Get current user info (requires authentication)
