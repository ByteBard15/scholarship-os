# Scholarship OS

## Purpose

Scholarship OS is an AI-assisted scholarship application management platform.

The current implementation phase supports canonical and derived profiles, scholarship catalog records, concrete applications, Research Tasks, human-reviewed Application Proposals, questionnaires, controlled application prefill, Information Requests, and USER/AGENT/SYSTEM authentication. A deterministic provider supports the complete local workflow; vendor-backed live research, Google Tasks synchronization, notifications, document generation, and autonomous submission remain future work.

## Architecture

- Frontend: React, TypeScript, and Vite
- Backend: Go, `net/http`, and chi
- Database: PostgreSQL
- ORM: GORM

The backend flow is handler → service → repository → GORM → PostgreSQL.

## Backend Rules

- Handlers must not access GORM directly. Handlers call services.
- Services own business rules. Repositories own persistence.
- Keep domain logic independent of HTTP.
- Use `context.Context` throughout repository and service boundaries.
- Do not expose GORM entities directly from HTTP handlers; use DTOs.
- Keep domain packages under `internal/features/<feature>`, with each feature owning its DTOs, handlers, services, repositories, and domain errors. Put reusable utilities under `pkg`; keep server, middleware, configuration, and database bootstrap under `internal`.
- Use `camelCase` for public JSON, query-string, and multipart field names. Use `snake_case` for PostgreSQL/GORM schema identifiers.
- Use `pkg/logger` for request-scoped handler, service, and repository logs. Pass the active `context.Context` so the request ID installed by middleware is included automatically; do not add `request_id` manually at each call site.
- Use UUIDs for every persistent domain ID.
- Avoid global mutable state.
- Prefer explicit code over reflection-heavy abstractions.
- Avoid generic repository frameworks.
- Schema changes require paired up/down migrations.
- Do not rely on AutoMigrate as the production migration strategy.
- Keep `cmd/server/main.go` focused on dependency wiring and process lifecycle.
- Effective profiles must be resolved through the profile service. Application code must not manually reconstruct inheritance.
- Master profiles contain canonical facts. Derived profiles customize inherited content with overrides or add their own relational section rows.
- Imported CV facts remain candidates until explicitly reviewed and applied.
- Applications must link to an owned Application Profile created from a Master or Domain parent. Never mutate the parent while tailoring an application.
- Research providers return typed `ResearchResult` values. They do not manipulate GORM models.
- Research findings are proposals until reviewed. Applying accepted findings is transactional and preserves source provenance.
- Do not silently overwrite conflicting application facts. Retain both findings and require review.
- Deadline precision matters: vague source text must not be converted into an invented exact timestamp.
- Agents and integrations use `internal/agenttools`; never expose a generic database tool.
- No discovered opportunity becomes an `Application` until a human explicitly approves its `ApplicationProposal`.
- `ResearchTask` input, links, runs, sources, findings, proposals, and outputs retain a traceable lineage.
- Prefill agents may write supported factual fields and suggested answers, but subjective answers remain reviewable and unsupported facts become deduplicated `InformationRequest` records.
- An `InformationRequest` becomes completed only after its response has been transactionally applied to the controlled target. Completed requests are not reopened implicitly.

## Database Rules

- PostgreSQL is the source of truth.
- All domain primary keys and foreign keys use UUID.
- Foreign keys must be explicit and have intentional delete behavior.
- Create useful indexes for ownership and lookup paths.
- Avoid storing structured profile information as giant JSON blobs.
- Core profile sections must remain relational.
- Timestamps are UTC; date-only profile fields use PostgreSQL `DATE`.

## Authentication

Three request principals exist: `USER`, `AGENT`, and `SYSTEM`.

- USER opaque session tokens access only resources owned by that user.
- AGENT API keys access only dedicated agent routes allowed by their scopes and associated user workspace.
- SYSTEM authentication may access every endpoint and bypasses workspace ownership checks.
- Only SYSTEM may register users or issue and revoke agent credentials.
- Agents may research, propose, prefill, and process controlled Information Requests, but may not approve Application Proposals.
- Agents may not change passwords, manage users, or modify canonical Master Profile facts through agent routes.
- Authentication and ownership checks are centralized in middleware and reinforced at domain service entry points.
- Never expose passwords, password hashes, session tokens, agent keys, the system key, or Authorization headers in logs.
- `SYSTEM_API_KEY` is root-equivalent and must never be compiled into the frontend or given to agents.

## AI Safety Rules

AI agents must never receive unrestricted SQL access. Agents operate through typed domain tools and must never silently overwrite canonical applicant facts. AI-generated modifications are proposals, reviewed research findings, or derived-profile overrides. New extracted facts are candidates until approved. Imported CV data must go through review. Tools consume `EffectiveProfile`, never raw GORM models or hand-built inheritance logic. Agents must not assess admission probability or submit applications.

## Credentials

Never commit `.env`, credentials, API keys, OAuth secrets, service-account files, or tokens. Use environment variables and keep `.env.example` free of secrets.

## Development Philosophy

Prefer simple, explicit, testable, and composable solutions over clever, abstract, prematurely distributed, or framework-heavy ones. Keep this foundation narrow until a later phase is explicitly requested.
