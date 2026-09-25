# Architecture overview

Scholarship OS is a monorepo with independently runnable API and web applications. The API is the only component that owns domain rules or database access. The browser consumes versioned JSON endpoints.

```text
React/Vite → JSON API → chi handler → service → repository → GORM → PostgreSQL
```

The present bounded contexts are users and applicant profiles. An account-level `User` can own multiple `ApplicantProfile` aggregates. Profile sections are relational child entities loaded only for their specific endpoints or the explicit `/full` endpoint.

External authentication can later map provider identities to `User.ExternalAuthID`. Authentication middleware will eventually be mounted centrally in the server package; no fake identity behavior exists today. Future scholarship, application, document, integration, and agent domains must be introduced as separate phases.
