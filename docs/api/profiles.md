# Profile API

The API base path is `/api/v1`. Single resources use `{ "data": ... }`, collections use `{ "data": [...], "meta": { "count": n } }`, and errors use `{ "error": { "code": "...", "message": "..." } }`.

## Users and profiles

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | API and database health |
| POST | `/api/v1/users` | Create a user |
| GET | `/api/v1/users/:userID` | Get a user |
| GET | `/api/v1/users/:userID/profiles` | List profiles owned by a user |
| POST | `/api/v1/users/:userID/profiles` | Create a profile; the first becomes default |
| GET | `/api/v1/profiles/:profileID` | Get profile metadata |
| PATCH | `/api/v1/profiles/:profileID` | Update metadata or switch the default |
| GET | `/api/v1/profiles/:profileID/full` | Get the aggregate and all sections |

## Personal information

`GET`, `PUT`, and `PATCH /api/v1/profiles/:profileID/personal-info` retrieve, replace/upsert, and partially update personal information. PUT requires `firstName` and `lastName`.

## Repeating sections

The section names are `education`, `employment`, `projects`, `publications`, `articles`, `skills`, `research-interests`, `career-goals`, `certifications`, `awards`, and `volunteering`.

Each supports:

| Method | Path |
|---|---|
| GET | `/api/v1/profiles/:profileID/{section}` |
| POST | `/api/v1/profiles/:profileID/{section}` |
| GET | `/api/v1/profiles/:profileID/{section}/:sectionID` |
| PATCH | `/api/v1/profiles/:profileID/{section}/:sectionID` |
| DELETE | `/api/v1/profiles/:profileID/{section}/:sectionID` |

DELETE returns `204 No Content`. Date values use RFC 3339 JSON strings because the Go transport currently uses nullable `time.Time`; only their date component is persisted and semantically relevant. Unknown JSON fields are rejected. Required fields include email, profile name, institution/degree/field of study, job organization/title, project/publication/article/award title, skill/research-interest/certification name, career-goal title/description, and volunteering organization/role.
