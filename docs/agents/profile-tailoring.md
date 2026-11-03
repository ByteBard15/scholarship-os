# Agent profile tailoring

Deploy the updated API; no database migration is needed. An administrator must issue or update a workspace-associated agent credential with `profiles`. Private writing retrieval and draft storage separately require `writing`. Never give an agent SYSTEM credentials.

All routes below use `Authorization: Bearer <agent-key>` and JSON. Ownership and scope checks apply in middleware and services.

| Method | Path under `/api/v1` | Purpose |
| --- | --- | --- |
| GET | `/agent/profiles` | List workspace profiles (metadata) |
| GET | `/agent/profiles/{profileID}/effective` | Resolve inherited facts through the profile service |
| POST | `/agent/profiles` | Create a Domain profile from an owned Master |
| PATCH | `/agent/profiles/{profileID}` | Update an owned Master or derived profile name, headline or summary |
| POST | `/agent/profiles/{profileID}/overrides` | Replace supported section description wording |

Create a scholarship scope:

```json
{"parentProfileId":"<owned-master-uuid>","name":"Biomedical engineering","headline":"Electronic and computer engineering graduate pursuing biomedical signal processing","summary":"<supported, user-reviewable wording>"}
```

Patch accepts only `name`, `headline`, and `summary`. Creation fixes `profileType` to Domain and does not make it the default. PATCH may update an owned Master Profile as authorized by the workspace owner. These operations do not bootstrap an Application. Application profiles are created through the approved-proposal workflow.

Description overrides require an existing effective section entity, `fieldName: "description"`, `overrideType: "replace"`, a string `value`, and a nonblank `reason` identifying supporting facts. `source` is forced to `agent`.

```json
{"entityType":"education","entityId":"<effective-education-uuid>","fieldName":"description","overrideType":"replace","value":"<supported wording>","reason":"Emphasise DSP coursework already recorded in the Master Profile."}
```

Supported entity types: education, employment, projects, publications, articles, certifications, awards, volunteering, research_interests, career_goals. The override endpoint remains limited to derived descriptions. Direct personal/section endpoints below allow supported canonical fact maintenance on an owned Master, including dates, grades and authorship. Do not invent facts or turn an ongoing manuscript into a published paper. Text remains subject to factual review; this API restricts fields, not the truth of generated prose.

Before creating a Domain profile, list existing profiles to avoid duplicates. After any change, resolve its effective profile and verify that the intended text is visible and parent facts remain intact when tailoring a derived profile. Master edits intentionally affect inherited facts in descendant profiles unless overridden.

For private writing, list `/agent/writing-tags` before calling `/agent/writing-samples/match`; retrieve no more than four relevant samples. Existing writing creation requires source sample IDs, preserves `derived_from`, and saves generated work as `suggested`. Profile-only local drafts are not a substitute for a source writing sample and must not be uploaded with invented provenance.

## Owner-authorized Master maintenance

The workspace owner has explicitly authorized agents with `profiles` to maintain the Master Profile. Existing credentials with that scope gain these capabilities after deployment; no new scope or migration is required. Agents still cannot manage accounts, approve proposals, submit applications, change profile ownership/type/default status, or delete profiles/section records through these routes.

- GET, PUT, PATCH `/agent/profiles/{profileID}/personal-info`: read, create/replace or patch personal information. PUT requires first and last names; PATCH requires an existing record.
- GET, POST `/agent/profiles/{profileID}/{section}`: read/add relational entries.
- GET, PATCH `/agent/profiles/{profileID}/{section}/{sectionID}`: read/update entries belonging to that profile.

Sections: `education`, `employment`, `projects`, `publications`, `articles`, `skills`, `research-interests`, `career-goals`, `certifications`, `awards`, `volunteering`. Request fields and validation match the corresponding user-facing profile endpoints. Use the resolved EffectiveProfile as research context, and edit actual section records on their owning profile; inherited entries must be tailored with derived overrides or deliberately maintained on their Master source.

Example Master summary update:

```http
PATCH /api/v1/agent/profiles/<master-profile-uuid>
Content-Type: application/json
Authorization: Bearer <agent-key>

{"summary":"<supported updated biography>"}
```

Always report canonical changes to the owner and retain uncertain claims as pending information. Imported CV facts still use the import review flow. Agent profile writes produce request-scoped metadata logs and best-effort audit entries, without logging profile contents or credentials.
