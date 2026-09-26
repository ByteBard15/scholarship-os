# Research Tasks

`ResearchTask` is the user-owned inbox item for manual or agent research. It stores the objective, custom instructions, links, optional target Application, lifecycle timestamps, and parent/child lineage. Parent cycles and cross-user parenting are rejected.

Queued tasks are consumed by an out-of-process authenticated research agent. The agent polls `GET /api/v1/agent/research-tasks?limit=N`, atomically claims one queued task, reads its links and optional target Application, then writes sources, findings, and proposals through typed agent APIs. Claiming creates a `ResearchRun`; it does not execute research inside the HTTP request.

Multiple pollers may observe the same queued item, but only one can claim it. After submitting results, the agent explicitly completes the run, which moves it to `review_required`. Proposal approval or rejection remains a human/system operation. Agent credentials and tokens must never be logged.
