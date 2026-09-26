# Profile inheritance

Scholarship OS uses a deliberately shallow hierarchy:

```text
Master Profile → Domain Profile → Application Profile
```

The master profile is the canonical factual source. A domain profile must inherit from a master owned by the same user. An application profile must inherit from a domain owned by the same user. Self-parenting, cross-user parents, chains deeper than these three levels, and cycles are rejected by the service.

The effective profile resolver starts with master relational content, appends whole entities owned directly by each derived profile, and applies overrides from root to child. Child overrides win. A `replace` override changes one inherited field only in the effective view, `hide` removes an inherited entity from that view, and `append` appends wording to a string field. Removing an override restores inheritance. Whole new profile-specific entities are normal relational rows belonging to the derived profile, not JSONB append records.

Consumers—including future agents—must call the effective-profile service/API. They must not query raw profile rows and manually reconstruct inheritance.

No `ApplicationID` column exists yet because the scholarship application domain is intentionally absent. A later additive migration can introduce that nullable relationship when applications are implemented.
