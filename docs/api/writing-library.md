# Writing library API

The writing library stores reusable source material and reviewable generated drafts. It is deliberately generic: `documentType` distinguishes personal statements, motivation letters, recommendation letters, statements of purpose, essays, cover letters, emails, and other writing.

Canonical writing remains relational. Tags are normalized per user and connected through `writing_sample_tags`. When an agent derives a new draft from existing samples, `writing_sample_relations` records every source sample. Generated work is always saved with `origin: agent` and `status: suggested`; it is not automatically approved.

## User endpoints

- `GET /api/v1/writing-samples`
- `POST /api/v1/writing-samples`
- `GET /api/v1/writing-samples/{sampleID}`
- `PATCH /api/v1/writing-samples/{sampleID}`
- `DELETE /api/v1/writing-samples/{sampleID}`
- `GET /api/v1/writing-tags`
- `POST /api/v1/writing-tags`
- `DELETE /api/v1/writing-tags/{tagID}`

Sample lists support `documentType` and `applicationId` filters. SYSTEM requests must supply `userId`; USER requests always resolve the authenticated owner.

```json
{
  "title": "Healthcare technology motivation",
  "documentType": "motivation_letter",
  "content": "...",
  "description": "Reusable source statement",
  "tagNames": ["biomedical engineering", "medical imaging", "Nigeria"]
}
```

## Agent retrieval workflow

Agent credentials require the `writing` scope.

1. `GET /api/v1/agent/writing-tags` returns tag names and sample counts without document content.
2. `POST /api/v1/agent/writing-samples/match` accepts proposal/application tags and returns at most four source or approved samples ranked by normalized tag overlap.
3. `POST /api/v1/agent/writing-samples` saves a generated draft. `sourceSampleIds` preserves immutable `derived_from` provenance. Profile-backed drafts can instead provide an owned `sourceProfileId`; recommendation letters still require a base writing sample. See the [preparation access guide](../agents/application-preparation-access.md) for snapshot provenance and draft revision routes.

```json
{
  "tags": ["biomedical engineering", "medical imaging"],
  "documentType": "motivation_letter",
  "applicationId": "00000000-0000-0000-0000-000000000000",
  "limit": 4
}
```

Match confidence is retrieval confidence only: matched requested tags divided by requested tags. It does not measure writing quality or application success.

```json
{
  "applicationId": "00000000-0000-0000-0000-000000000000",
  "title": "Biomedical Engineering motivation draft",
  "documentType": "motivation_letter",
  "content": "...",
  "tagNames": ["biomedical engineering", "medical imaging"],
  "sourceSampleIds": ["00000000-0000-0000-0000-000000000000"]
}
```

The API overrides attempted agent values for `origin` and `status`, ensuring generated content remains a reviewable suggestion.

Agent metadata search and draft revision: see the [application preparation access guide](../agents/application-preparation-access.md).
