# auth-service

Standalone gRPC microservice handling authentication, users, and roles for the [NextJS-GO Template](https://github.com/Mognus/nextjs-go-template).

**Design docs:** [[auth-microservice-grpc]] | [[auth-grpc-backend-integration]]

---

## Overview

- Owns the `users` and `roles` database tables
- Exposes a gRPC interface on `:50051`
- Manages its own migrations on startup
- JWT is issued here, validated locally in the backend (shared secret, no RPC per request)
- Cookies are set by the backend HTTP layer, not here

## gRPC Interface

```
AuthService
  Login / Register

  GetUser / ListUsers / CreateUser / UpdateUser / DeleteUser
  GetRole / ListRoles / CreateRole / UpdateRole / DeleteRole
```

Proto definition: `proto/auth/v1/auth.proto`
Generated code: `gen/auth/v1/` (do not edit)

## Environment Variables

| Variable      | Default     | Description             |
|---------------|-------------|-------------------------|
| `DB_HOST`     | `localhost` | Postgres host           |
| `DB_PORT`     | `5432`      | Postgres port           |
| `DB_USER`     | `postgres`  | Postgres user           |
| `DB_PASSWORD` | `postgres`  | Postgres password       |
| `DB_NAME`     | `app_db`    | Database name           |
| `DB_SSLMODE`  | `disable`   | SSL mode                |
| `JWT_SECRET`  | —           | Required. Shared with backend. |

## Development

Runs via Docker Compose as part of the main stack:

```bash
make dev
```

Or standalone with Air:

```bash
air
```

## Seeding

Users are seeded via the script in `scripts/seed_users.go` in the parent repo — it writes directly to the database:

```bash
cd scripts && go run seed_users.go -fixed-roles -password=test123
# creates: seed_admin@example.com, seed_user@example.com, seed_guest@example.com
```

## Regenerating Proto Code

```bash
buf generate
```

Requires [buf](https://buf.build/docs/installation) and a `buf.build` account for remote plugins.
