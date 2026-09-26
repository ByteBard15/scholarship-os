# Profile inheritance and overrides API

- `GET /api/v1/profiles/:profileID/lineage` returns root-to-current lineage.
- `GET /api/v1/profiles/:profileID/effective` returns resolved content without raw override internals.
- `GET /api/v1/profiles/:profileID/completeness` returns coverage only, never applicant quality or admission probability.
- `GET /api/v1/profiles/:profileID/compare/:otherProfileID` reports inherited, hidden, modified, and appended content.
- `GET|POST /api/v1/profiles/:profileID/overrides` lists or creates overrides.
- `PATCH|DELETE /api/v1/profiles/:profileID/overrides/:overrideID` updates or restores inherited content.
- `GET|POST /api/v1/profiles/:profileID/snapshots` lists or creates immutable effective-profile snapshots.
- `GET /api/v1/profiles/:profileID/snapshots/:snapshotID` retrieves one snapshot.
- `GET|POST /api/v1/profiles/:profileID/evidence` lists or creates provenance.

Profile creation accepts `profileType` and `parentProfileId`. Master profiles reject parents; domain profiles require a master parent; application profiles require a domain parent. Override values are JSON because target field types vary, but canonical facts remain relational.
