# Profile Agent Rules

The intended profile hierarchy is:

```text
Master/Base Profile
    ↓
Domain Profile
    ↓
Application Profile
```

Only the Master/Base Profile exists today. Future domain and application profiles will inherit factual base data and may add scoped overrides.

The canonical profile must remain factual. AI customization must never fabricate employment, education, publications, grades, skills, awards, certifications, or research experience. Agents should propose edits with provenance and require an explicit controlled acceptance workflow before canonical facts change.
