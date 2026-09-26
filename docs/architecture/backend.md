# Backend architecture

The API follows one dependency direction:

```text
HTTP handler → service → repository interface → GORM implementation → PostgreSQL
```

Authentication runs before protected handlers. Public routes are limited to health and login. The router separates USER/SYSTEM workspace routes, SYSTEM-only administration, and scoped AGENT routes. Middleware installs a typed request principal and validates route-resource ownership; user-owned service entry points reinforce owner filters before repositories run. SYSTEM is the single centralized ownership bypass.

Handlers parse UUIDs and JSON, run transport validation, invoke a service, translate domain errors, and serialize DTOs. Services enforce ownership-ready lookup boundaries, profile rules, application state transitions, task lineage, deadline precision, research review, and transactional application of findings. Repository interfaces isolate persistence and enable service tests without PostgreSQL. GORM implementations use request contexts and map missing records or conflicts to safe domain errors.

Domain packages live under `internal/features/<feature>` and keep their handlers, DTOs, services, repository interfaces, and GORM implementations together. Reusable infrastructure such as response helpers and file storage lives under `pkg`. Foundational process concerns remain under `internal/config`, `internal/database`, `internal/middleware`, and `internal/server`.

Public JSON, query-string, and multipart field names use `camelCase`. PostgreSQL tables, columns, indexes, constraints, and GORM column mappings use `snake_case`.

Research providers return typed results to `ResearchService`. The controlled `internal/agenttools` package wraps services and deliberately exposes no database handle or generic mutation function.

`internal/server` owns routing and cross-cutting middleware. Request IDs, recovery, CORS, and structured request logs are configured centrally. `cmd/server` loads typed configuration, opens the database pool, wires dependencies, and performs signal-driven graceful shutdown.

The request logging middleware derives a request-scoped logger through `pkg/logger` and stores it in the request context. Handlers, services, and repositories should log through that package with their active context; the wrapper adds `request_id` automatically and falls back safely to the process logger outside HTTP requests.
