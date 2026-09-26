# Research pipeline

Research may now start from a user-owned `ResearchTask` before any Application exists. Task links and instructions are passed to a typed `ResearchExecutor`; its sources and findings are persisted with a task-owned Research Run. Opportunity discovery produces a pending Application Proposal and stops at the human approval boundary. Only approval creates the Application and links the retained research provenance.

```text
Application
  → ResearchRun
  → ResearchSource
  → typed ResearchFinding
  → human review
  → transactional apply
  → requirements / deadlines / funding / contacts / supervisors / URLs
```

`ApplicationResearcher` isolates provider behavior from orchestration. The built-in mock provider is deterministic, makes no network calls, and requires no credentials. Providers return `ResearchResult`, not prose or persistence models.

Each source records URL and retrieval time. Findings default to `pending`; conflicting values remain separate and are marked `conflicting`. Apply refuses runs with pending findings, ignores rejected findings, inserts accepted structured records without overwriting existing records, retains `SourceID`, updates research status, and writes an audit entry.

Future web-search and AI providers may replace the mock behind the same interface. They must not receive SQL access.
