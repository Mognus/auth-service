# auth-service

Standalone gRPC microservice handling authentication, users, and roles for the [NextJS-GO Template](https://github.com/Mognus/nextjs-go-template).

---

## Overview

- Owns the `users`, `roles`, and `refresh_tokens` database tables
- Exposes a gRPC interface on `:50051`
- Runs DB migrations on startup via `cmd/migrate`
- JWT issued here, validated locally in the backend (shared secret — no RPC per request)
- Cookies are set by the backend HTTP layer, not here

## gRPC Interface

```
AuthService
  Login / Register / RefreshToken / Logout

  GetUser / ListUsers / CreateUser / UpdateUser / DeleteUser
  GetRole / ListRoles / CreateRole / UpdateRole / DeleteRole
```

Proto definition: `proto/auth/v1/auth.proto`
Generated code: `gen/auth/v1/` (do not edit)

## Environment Variables

| Variable          | Default     | Description                    |
|-------------------|-------------|--------------------------------|
| `DB_HOST`         | `localhost` | Postgres host                  |
| `DB_PORT`         | `5432`      | Postgres port                  |
| `DB_USER`         | `postgres`  | Postgres user                  |
| `DB_PASSWORD`     | `postgres`  | Postgres password              |
| `DB_NAME`         | `app_db`    | Database name                  |
| `DB_SSLMODE`      | `disable`   | SSL mode                       |
| `JWT_SECRET`      | —           | Required. Shared with backend. |
| `JWT_ACCESS_TTL`  | `15m`       | Access token lifetime          |
| `JWT_REFRESH_TTL` | `168h`      | Refresh token lifetime         |

## Migrations

Migrations live in `migrations/` and run automatically on container startup before Air.

```bash
make migrate-up       # apply pending migrations
make migrate-down     # roll back all migrations
make migrate-version  # show current version
```

Tool: [golang-migrate](https://github.com/golang-migrate/migrate)

## Development

```bash
make dev        # start full stack (runs migrations automatically)
make dev-build  # rebuild images + start
```

## Seeding (Dev only)

```bash
make seed n=10  # create 10 test users  → user1@dev.local … / password123
make seed-root  # create root admin     → root@dev.local / root1234
make seed-all   # one user per role     → root + user + guest
```

## Regenerating Proto Code

```bash
buf generate
```

Requires [buf](https://buf.build/docs/installation).
