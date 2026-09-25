# Scholarship OS

## Purpose

Scholarship OS is an AI-assisted scholarship application management platform.

The current implementation phase is base account and applicant-profile infrastructure only. Future domains will include scholarships, applications, application-specific profiles, research agents, requirements, deadlines, CV generation, motivation letters, statements of purpose, and integrations such as Google Tasks. Do not prematurely implement future domains unless explicitly requested.

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
- Use UUIDs for every persistent domain ID.
- Avoid global mutable state.
- Prefer explicit code over reflection-heavy abstractions.
- Avoid generic repository frameworks.
- Schema changes require paired up/down migrations.
- Do not rely on AutoMigrate as the production migration strategy.
- Keep `cmd/server/main.go` focused on dependency wiring and process lifecycle.

## Database Rules

- PostgreSQL is the source of truth.
- All domain primary keys and foreign keys use UUID.
- Foreign keys must be explicit and have intentional delete behavior.
- Create useful indexes for ownership and lookup paths.
- Avoid storing structured profile information as giant JSON blobs.
- Core profile sections must remain relational.
- Timestamps are UTC; date-only profile fields use PostgreSQL `DATE`.

## AI Safety Rules

Future AI agents must never receive unrestricted SQL access. Agents must operate through typed domain tools and must never silently overwrite canonical applicant facts. AI-generated modifications should be proposals or application-specific overrides. Canonical user profile data represents factual source material.

## Credentials

Never commit `.env`, credentials, API keys, OAuth secrets, service-account files, or tokens. Use environment variables and keep `.env.example` free of secrets.

## Development Philosophy

Prefer simple, explicit, testable, and composable solutions over clever, abstract, prematurely distributed, or framework-heavy ones. Keep this foundation narrow until a later phase is explicitly requested.
