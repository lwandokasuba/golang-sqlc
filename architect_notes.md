# Architect's Notes

## Trade-off Analysis

### 1. SQLC vs ORM
**Decision:** We chose `sqlc` over GORM or Ent.
**Reason:**
- **Type Safety:** `sqlc` generates type-safe code derived directly from SQL, catching errors at compile time (or generate time).
- **Performance:** No reflection overhead; it's just raw SQL execution.
- **DBA-Friendly:** DBAs can review raw SQL files without understanding Go structs.
- **Trade-off:** slightly more friction (must regenerate code) vs "magic" ORMs. But simpler debugging.

### 2. Transaction Management "Store" Pattern
**Decision:** We implemented a `Store` struct wrapping `pgxpool`.
**Reason:**
- `sqlc` generates `WithTx` but doesn't prescribe how to pass the transaction context around.
- The `Store` pattern (wrapping `Queries` + `execTx`) allows us to define business transactions (`TransferTx`) as methods on the store, keeping the Service layer clean of `Begin/Commit` logic.
- **Alternative:** Passing `*sql.Tx` through every function. This is error-prone and leaks DB details.

### 3. "No SELECT *" Strategy
**Decision:** Enforce explicit column selection.
**Reason:**
- Prevents breaking code when columns are added/removed.
- Prevents leaking PII (passwords, internal flags) via API if generic structs are marshaled.
- We use DTOs to strictly define the API contract, decoupling it from the DB schema.

### 4. Microservices Shared DB Strategy
**Decision:** Monorepo/Module approach simulated.
**Reason:**
- Sharing a live DB across microservices is generally an anti-pattern (tight coupling).
- If absolutely necessary, a shared Go module (`github.com/org/db-contract`) containing the migrations and generated `sqlc` code is better than ad-hoc queries.
- **Recommendation:** Each service should own its own schema (private tables). If they must share data, use Event Sourcing or API aggregation, not shared tables.

### 5. Config Management
**Decision:** `viper` for environment variables.
**Reason:** Standard, robust, supports 12-factor apps.

### 6. Business Logic Placement
**Decision:** Currency validation is done in the Service layer, *before* the transaction.
**Reason:** 
- While DB constraints are final safeguards, checking business rules (like Currency matching) in the application layer provides better error messages to the user and reduces DB load for invalid requests.
- "Insufficient Funds" is also checked in the application logic (after reading balance) for clarity, though DB constraints (`CHECK balance >= 0`) would also catch race conditions (which we mitigate via `FOR UPDATE` or atomic updates in the Store).

### 7. Concurrency Testing
**Decision:** We write heavy integration tests (e.g., `TestTransferTx` with Go routines) for critical paths.
**Reason:** 
- Distributed systems and DBs often have subtle race conditions (deadlocks, phantom reads).
- Unit tests mocking the DB driver cannot catch these.
- We deliberately run concurrent transactions in tests to ensure our `sqlc` locking strategy (`UPDATE ...`) works as expected.

### 8. Optimization: Single Query Join
**Decision**: Use `LEFT JOIN` for `GetUser` with accounts instead of 2 separate queries.
**Reason**:
- While 2 queries is simpler to write, `LEFT JOIN` is more performant (1 RTT).
- We implemented manual row folding in Go because `sqlc` (and Go's `database/sql`) returns flat rows for joins. This adds some code complexity (row iteration) but keeps the SQL efficient.
