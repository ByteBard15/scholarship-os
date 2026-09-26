# Research Tasks

`ResearchTask` is the user-owned inbox item for manual or agent research. It stores the objective, custom instructions, selected Applicant Profile, links, optional target Application, lifecycle timestamps, and parent/child lineage. Parent cycles, cross-user parenting, and cross-user profile selection are rejected. Draft input can be edited, but it is frozen when the task is queued.

`ResearchContext` is an independent, reusable user-owned question and answer, such as course interests or existing personal-statement material. The `research_task_contexts` join table lets multiple tasks consume the same context without duplicating it. This context informs research only: it is not an `InformationRequest`, does not represent missing application data, and never modifies the canonical Master Profile.

Queued tasks are consumed by an out-of-process authenticated research agent. The agent polls `GET /api/v1/agent/research-tasks?limit=N`, atomically claims one queued task, reads its links, attached Research Context, and optional target Application, then writes sources, findings, and proposals through typed agent APIs. Claiming creates a `ResearchRun`; it does not execute research inside the HTTP request.

Multiple pollers may observe the same queued item, but only one can claim it. After submitting results, the agent explicitly completes the run, which moves it to `review_required`. Proposal approval or rejection remains a human/system operation. Agent credentials and tokens must never be logged.
