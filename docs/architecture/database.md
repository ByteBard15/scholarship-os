# Database architecture

`users` is the account table. `applicant_profiles` belongs to a user and is the aggregate root for reusable applicant facts. A partial unique index ensures at most one active default profile per user, while the service switches defaults transactionally.

Profiles have a `profile_type` of `master`, `domain`, or `application` and an optional self-referencing parent. Master profiles have no parent, domain profiles require a master parent, and application profiles require either a master or domain parent. Domain parenting is preferred when tailoring by discipline. The service verifies common ownership and rejects cycles; database checks provide a second line of defense for type and self-parent rules.

Each profile has zero or one `profile_personal_info` row and zero or more rows in education history, employment history, projects, publications, articles, skills, research interests, career goals, certifications, awards, and volunteer experiences. All primary and foreign keys are PostgreSQL UUIDs. Child rows cascade when their parent profile is permanently removed; ordinary profile and list-section removal uses GORM soft deletion.

Date-oriented fields are PostgreSQL `DATE`. Go currently represents them as nullable `time.Time`; only the calendar date is semantically meaningful. All audit timestamps are stored as `TIMESTAMPTZ` and generated in UTC. Application-generated UUIDs are authoritative, so no UUID extension or database default is required.

Phase 2 adds `profile_overrides`, immutable `profile_snapshots`, uploaded `profile_documents`, reviewed `profile_imports` and `profile_import_candidates`, `profile_evidence`, and lightweight `audit_logs`. JSONB is limited to polymorphic override values, immutable effective-profile snapshots, typed extraction payloads, candidate payloads, and small audit metadata. Canonical profile facts remain relational.

Phase 4 adds `research_tasks`, task links/outputs, human-reviewed `application_proposals`, application fields, questionnaires/questions/answers, information requests with immutable response history, prefill runs/actions, and concise agent activities. `research_runs.application_id` is nullable and task-owned runs use `research_task_id`; at least one owner is required. JSONB remains limited to typed polymorphic values and provider candidates rather than replacing relational application data.

## Application domain

`Institution`, `Programme`, and `Scholarship` are reusable catalog records. An `Application` belongs to a user and one derived Application Profile, and optionally references each catalog entity. Requirements, deadlines, funding, contacts, supervisors, URLs, and tasks are separate application-owned rows. Task parents are self-references validated by the service. Requirement evidence links requirements to effective-profile entities and optional profile evidence.

`ResearchRun` owns source observations and structured findings. Application-owned records may reference a `ResearchSource` so researched facts retain provenance. Findings use JSONB only for typed variable values and include independent verification and human-review states. Applying findings creates relational application records in one transaction.

`ExternalTaskReference` maps an application task to a future provider task without implementing synchronization.
