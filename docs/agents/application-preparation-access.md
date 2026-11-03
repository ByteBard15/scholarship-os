# Agent access for application preparation

This update adds the preparation APIs; deploying code does not silently expand existing agent credentials. Only SYSTEM can amend scopes. Agents cannot grant their own access, obtain SYSTEM credentials, approve proposals/writing, sign references, send external messages or submit external applications through these APIs.

## Deploy and grant access

1. Apply migration `000017_agent_writing_preparation` before deploying the updated backend. It adds the `email` document type and a nullable snapshot provenance link on writing samples. The paired down migration converts email records to `other`, preserves their content, and removes the new provenance column.
2. Deploy backend and frontend.
3. Call `GET /api/v1/agent/capabilities` with the existing agent key. It returns the agent credential ID, workspace, granted scopes, missing preparation scopes and human steps. It does not expose keys or grant permissions.
4. An administrator uses SYSTEM authentication to call `PATCH /api/v1/admin/agent-credentials/{credentialID}/scopes` with the body below. This replaces the scope set; include every scope to retain. Existing API keys remain valid, and the next authenticated request reads the new scopes from PostgreSQL.
5. Repeat the capability check. `preparationReady: true` means the workspace and required scopes are present, not that a worker is running or every external integration exists.

```json
{
  "scopes": [
    "research", "applications", "profiles", "writing",
    "questionnaires", "tasks", "information_requests"
  ]
}
```

The administrator can find credential IDs with the existing SYSTEM-only `GET /api/v1/admin/agent-credentials`. Updates are transactional with an audit record and cannot reactivate revoked credentials. Never give an agent the SYSTEM key. There is no endpoint by which an agent can add its own scopes.

## Preparation capabilities

| Scope | Supported operations |
| --- | --- |
| research | Create/import workspace research tasks, list drafts, queue, claim, add sources/findings, create proposals and complete research |
| applications | Read an application, materialize a human-approved proposal, inspect fields/progress and run controlled prefill |
| profiles | Read EffectiveProfiles, maintain owner-authorized Master data, create Domain variants and tailor derived profiles |
| writing | Discover tags/titles, retrieve relevant samples, create suggested drafts and revise suggested agent work |
| questionnaires | Inspect/create application questionnaires and unanswered prompts |
| tasks | Read/create/update application checklist items |
| information_requests | List/create/read controlled requests and apply supplied responses through existing processing rules |

New routes under `/api/v1`:

- `GET /agent/capabilities`
- `POST /agent/research-tasks?queue=true`; uses the existing creation DTO with owned `userId` and `profileId`.
- `POST /agent/research-tasks/import`; uses the existing import schema with `queue`.
- `GET /agent/applications/{applicationID}/fields`
- `GET /agent/applications/{applicationID}/preparation`
- `GET`, `POST` `/agent/applications/{applicationID}/tasks`
- `PATCH /agent/applications/{applicationID}/tasks/{taskID}`
- `GET`, `POST` `/agent/applications/{applicationID}/questionnaires`
- `GET /agent/applications/{applicationID}/questionnaires/{questionnaireID}`
- `GET`, `POST` `/agent/questionnaires/{questionnaireID}/questions`
- `GET /agent/information-requests`

Agent-created questionnaires start `not_started` and prompts start `unanswered`, regardless of requested status. Answers and factual prefill use the existing controlled prefill flow, with unsupported facts becoming Information Requests and subjective text remaining reviewable. These routes do not approve answers or proposals. Task completion describes a local checklist item, not external submission.

## Writing workflow

List tags first, then inspect only the metadata necessary to find a writing sample:

```http
GET /api/v1/agent/writing-tags
GET /api/v1/agent/writing-samples?title=Base%20Recommendation%20Letter&limit=20
GET /api/v1/agent/writing-samples?title=Base%20Recommendation%20Letter%20%232&limit=20
GET /api/v1/agent/writing-samples/{sampleID}
```

Metadata listing returns IDs, titles, type, status, origin, application association and update time, without content. It supports exact case-insensitive title, documentType, limit (1–50) and offset. Source matching still returns at most four source/approved documents. Exact-ID retrieval can read an owned draft, including the user's unapproved base letters. Reading a draft does not verify its claims.

Create `POST /agent/writing-samples` with existing `sourceSampleIds` to preserve `derived_from`. Recommendation letters require a source writing sample; do not invent referee relationships or firsthand observations. Generated work is forced to `origin: agent`, `status: suggested`.

A new email may instead use an owned `sourceProfileId` with both `writing` and `profiles` scopes:

```json
{
  "applicationId": "<owned-application-uuid>",
  "title": "Calgary supervisor enquiry",
  "documentType": "email",
  "content": "<supported, reviewable message>",
  "tagNames": ["calgary", "biomedical-engineering", "supervisor-email"],
  "sourceProfileId": "<owned-effective-profile-uuid>"
}
```

The service checks ownership through the profile service and creates an immutable EffectiveProfile snapshot. The response's `sourceProfileSnapshotId` identifies the factual source. Snapshot creation precedes draft persistence: if persistence fails, the immutable snapshot may remain without a draft. Writing-sample relations and tags remain transactional with draft creation. This permits profile-based drafts without inventing an unrelated source writing sample. Recommendation letters still need their base sample ID even if a profile is supplied.

`PATCH /agent/writing-samples/{sampleID}` revises only suggested agent-generated writing. Agents cannot edit human source writing, approved work, archived work or review status. Existing source and snapshot provenance cannot be replaced by PATCH. The repository locks the row and rechecks review status so concurrent human approval cannot be overwritten.

## Human review and current limits

Proposal approval, writing approval and referee sign-off remain human decisions. No unrestricted SQL, credential management, message sending or external submission is exposed to agents. Vendor-backed research, notification delivery and an autonomous background worker are not implemented by this API expansion. External agents can orchestrate the preparation steps using their assigned scopes; `preparationReady` reports access only.

After access is granted, the pending Calgary work is to retrieve the two base letters, tailor them without unsupported claims, save the three email drafts and recommendation drafts as suggested writing, and return their library IDs for review.
