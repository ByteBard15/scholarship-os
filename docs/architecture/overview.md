# Architecture overview

Scholarship OS is a monorepo with independently runnable API and web applications. The API is the only component that owns domain rules or database access. The browser consumes versioned JSON endpoints.

```text
React/Vite → JSON API → chi handler → service → repository → GORM → PostgreSQL
```

The bounded contexts are users, applicant profiles, opportunity catalog, applications, and controlled research. An account-level `User` owns canonical/derived profiles and concrete applications. Profile inheritance remains encapsulated by the profile service. Applications own relational requirements, deadlines, funding, contacts, supervisors, URLs, tasks, and evidence links. Research runs persist sources and reviewable findings before structured data is applied.

External authentication can later map provider identities to `User.ExternalAuthID`. Authentication middleware will eventually be mounted centrally in the server package; no fake identity behavior exists today. Future integrations and live agent providers remain behind typed service/tool boundaries.
