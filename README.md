# Template Monorepo

A full-stack monorepo using Turborepo with SolidJS, Hono, Drizzle ORM, and Better Auth.

## Stack

- **Runtime & Package Manager**: [Bun](https://bun.sh/)
- **Build System**: [Turborepo](https://turbo.build/)
- **Linter/Formatter**: [Biome](https://biomejs.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)

## Apps

- `apps/web` - [SolidJS](https://www.solidjs.com/) frontend with [Vite](https://vitejs.dev/)
- `apps/apis` - [Hono](https://hono.dev/) backend API (runs natively on Bun)

## Packages

- `packages/db` - Database client using [Drizzle ORM](https://orm.drizzle.team/) with PostgreSQL via Bun's native SQL driver
- `packages/types` - Shared TypeScript types

## Getting Started

### Prerequisites

- Bun >= 1.1.0
- PostgreSQL database

### Installation

```bash
bun install
```

### Environment Setup

Copy the example environment file for the API:

```bash
cp apps/apis/.env.example apps/apis/.env
```

Update `apps/apis/.env` with your database connection string and other settings.

### Database Setup

```bash
# Push schema to database
bun db:push

# Or generate and run migrations
bun db:generate
bun db:migrate
```

### Development

```bash
# Run all apps in development mode
bun dev

# Run specific app
bun --filter @template/web dev
bun --filter @template/apis dev
```

### Build

```bash
bun run build
```

### Lint & Format

```bash
# Check all files
bun check

# Format all files
bun format
```

## Auth

Authentication is handled by [Better Auth](https://www.better-auth.com/). The following features are configured:

- Email & Password authentication
- Session management via PostgreSQL

### Auth Routes (API)

- `POST /api/auth/sign-in/email` - Sign in with email/password
- `POST /api/auth/sign-up/email` - Register with email/password
- `POST /api/auth/sign-out` - Sign out
- `GET /api/auth/session` - Get current session

### Protected Routes

- `GET /api/me` - Get current user info (requires authentication)
