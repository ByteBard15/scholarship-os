# Backend architecture

The API follows one dependency direction:

```text
HTTP handler → service → repository interface → GORM implementation → PostgreSQL
```

Handlers parse UUIDs and JSON, run transport validation, invoke a service, translate domain errors, and serialize DTOs. Services enforce ownership-ready lookup boundaries, profile default rules, required section fields, and date ordering. Repository interfaces isolate persistence and enable service tests without PostgreSQL. GORM implementations use request contexts and map missing records or conflicts to safe domain errors.

`internal/server` owns routing and cross-cutting middleware. Request IDs, recovery, CORS, and structured request logs are configured centrally. `cmd/server` loads typed configuration, opens the database pool, wires dependencies, and performs signal-driven graceful shutdown.
