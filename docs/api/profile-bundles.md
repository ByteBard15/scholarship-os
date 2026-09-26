# Profile JSON bundles

Profile bundles provide a versioned JSON format for moving structured profile
facts into or out of Scholarship OS. The current schema is
[`profile-bundle.schema.json`](../schemas/profile-bundle.schema.json) version
`1.0`.

`GET /api/v1/profiles/{profileID}/export` exports the effective profile. Parent
facts and applied overrides are materialized in the returned sections. It does
not export documents, file storage keys, CV imports, evidence, snapshots,
override records, or audit history.

`POST /api/v1/users/{userID}/profiles/import` creates a new profile and all its
structured relational rows in one database transaction. `profileType` must be
`master`, `domain`, or `application`. A master profile must not have
`parentProfileId`; domain and application profiles require an owned, valid
parent under the normal profile hierarchy rules.

An exported domain or application bundle retains its original parent ID. When
moving it elsewhere, set `profileType` and `parentProfileId` to values valid in
the destination workspace. Imported bundle content is materialized as rows
owned by the new profile; it never mutates its parent.
