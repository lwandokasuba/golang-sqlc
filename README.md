# Golang SQLC Blueprint

This project demonstrates a robust, type-safe Golang backend architecture using `sqlc`, separating concerns between Database Administrators (DBA) and Backend Developers.

## Project Structure

- `cmd/api`: Application entrypoint.
- `internal/db`: Database access layer (generated code + Store implementation).
- `internal/db/migrations`: Database schema migrations (DBA owned).
- `internal/db/queries`: SQL queries (Developer owned).
- `internal/service`: Business logic layer.
- `internal/transport/http`: REST API handlers.
- `internal/dto`: Data Transfer Objects for API contracts.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [goose](https://github.com/pressly/goose) (for migrations)

### Setup

1.  **Start Database:**
    ```bash
    make network
    make postgres
    make createdb
    ```

2.  **Run Migrations:**
    ```bash
    make migrateup
    ```

3.  **Generate Go Code:**
    ```bash
    make sqlc
    ```

4.  **Run Server:**
    ```bash
    make server
    ```

## DBA vs Developer Workflow

### DBA Responsibilities
- Manage strict schema definitions in `internal/db/migrations`.
- Review performance of queries.
- Owns `create index` and constraints.
- Workflow:
    1.  Create migration: `goose -dir internal/db/migrations create add_users_table sql`
    2.  Review and apply: `make migrateup`

### Developer Responsibilities
- Write SQL queries in `internal/db/queries/*.sql`.
- **Rule:** NEVER use `SELECT *`. Always specify columns to avoid leaking data and breaking on schema changes.
- Workflow:
    1.  Add query to `.sql` file.
    2.  Run `make sqlc` to generate interface.
    3.  Implement business logic using the generated `Querier` interface.

## Architecture Decisions

- **Transactions:** We use a `Store` interface that embeds `Queries` and adds a `TransferTx` method. This allows business logic to run strictly within a transaction boundary provided by the `Store`, maintaining testability and atomicity.
- **DTOs:** We avoid returning database models directly in the API. `internal/dto` defines the API contract, ensuring internal DB changes don't accidentally rename JSON fields or leak sensitive columns like passwords.
- **Microservices Ready:** The `internal/db` package is consistent and self-contained. In a microservices environment, this folder structure can be moved to a shared library `db-contract` repo if strict sharing is needed, or simply replicated per service (DDD style) where each service owns its schema.

## API Endpoints

- `POST /users`: Create user.
- `GET /users/:id`: Get user.
- `POST /accounts`: Create account.
- `POST /transfers`: Execute money transfer (Atomic).

## Testing

Run integration tests (requires Docker DB running):
```bash
make test
```
This includes a concurrency test (`TestTransferTx`) that simulates 5 concurrent transfers to verify deadlock protection and atomicity.

## Schema Evolution Example

We demonstrated schema evolution by adding a `hashed_password` column to the `users` table.
1.  Created migration `002_add_users_password.sql`.
2.  Updated `internal/db/queries/user.sql` to include the new field.
3.  Regenerated code with `make sqlc`.
4.  Updated `Service` layer to match the new type signature.
