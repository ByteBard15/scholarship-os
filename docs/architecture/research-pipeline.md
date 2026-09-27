# Research pipeline

Research starts from a user-owned `ResearchTask` before any Application exists. An authenticated external agent polls and claims queued tasks, reads their links and instructions, and persists sources and findings through typed APIs on a task-owned Research Run. Opportunity discovery produces a pending Application Proposal and stops at the human approval boundary. Human approval records intent and the selected parent profile. A separate authenticated application-agent operation creates the Application and links the retained research provenance.

```text
Application
  → ResearchRun
  → ResearchSource
  → typed ResearchFinding
  → human review
  → transactional apply
  → requirements / deadlines / funding / contacts / supervisors / URLs
```

Research execution is intentionally out of process. The API persists queued work and exposes typed, scoped agent endpoints for claiming tasks and writing sources, findings, and proposals. The agent does not receive database access or generic mutation primitives.

Each source records URL and retrieval time. Findings default to `pending`; conflicting values remain separate and are marked `conflicting`. Apply refuses runs with pending findings, ignores rejected findings, inserts accepted structured records without overwriting existing records, retains `SourceID`, updates research status, and writes an audit entry.

Web-search and AI implementations may evolve independently as long as they use the same controlled APIs. They must not receive SQL access.
