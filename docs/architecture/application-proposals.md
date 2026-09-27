# Application Proposals

Research discovery creates `ApplicationProposal`, never `Application`. A proposal includes structured institution/programme/scholarship candidates, intake, confidence, a concise reasoning summary, and links to persisted Research Sources.

Approval and materialization are separate transactions. Human approval locks the pending proposal, validates and records its Master or Domain parent, and marks it approved. An application-scoped agent later locks the approved proposal, conservatively matches or creates catalog records, creates an isolated Application Profile, creates the Application and checklist, attaches research provenance, applies supported structured findings, and records task outputs and audit activity. Rejection retains all research. A rejected proposal may be explicitly reopened.
