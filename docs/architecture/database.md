# Database architecture

`users` is the account table. `applicant_profiles` belongs to a user and is the aggregate root for reusable applicant facts. A partial unique index ensures at most one active default profile per user, while the service switches defaults transactionally.

Each profile has zero or one `profile_personal_info` row and zero or more rows in education history, employment history, projects, publications, articles, skills, research interests, career goals, certifications, awards, and volunteer experiences. All primary and foreign keys are PostgreSQL UUIDs. Child rows cascade when their parent profile is permanently removed; ordinary profile and list-section removal uses GORM soft deletion.

Date-oriented fields are PostgreSQL `DATE`. Go currently represents them as nullable `time.Time`; only the calendar date is semantically meaningful. All audit timestamps are stored as `TIMESTAMPTZ` and generated in UTC. Application-generated UUIDs are authoritative, so no UUID extension or database default is required.
