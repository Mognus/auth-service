# Auth Service

Standalone Go authentication service for the Personal Blog project.

This service was built primarily as a learning project to get comfortable with
Go service code, PostgreSQL-backed auth flows, protobuf contracts, Buf code
generation, and gRPC. It is intentionally kept separate from the main backend
because the service boundary is still useful for experimenting with auth as an
isolated component.

## Scope

- User and role persistence.
- Login and session-related auth flows.
- Protobuf/gRPC service contract.
- Generated client/server code through Buf.
- Database migrations and seed commands.

## Context

The surrounding Personal Blog architecture is being simplified, but this service
remains a standalone module for now. It represents the original Go/gRPC learning
phase of the project and may later be replaced, rewritten, or folded into a
newer architecture depending on the direction of the Rust-based successor.

## Development

Generate protobuf code:

```bash
buf generate
```

Run from the repository root through Docker Compose:

```bash
make dev
```
