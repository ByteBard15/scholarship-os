# Research Tasks

`ResearchTask` is the user-owned inbox item for manual or agent research. It stores the objective, custom instructions, links, optional target Application, lifecycle timestamps, and parent/child lineage. Parent cycles and cross-user parenting are rejected.

Starting a task invokes the configured typed `ResearchExecutor`. A run can belong to a task without an Application. Sources and findings are persisted before proposals, and task outputs record every produced entity. The deterministic local provider allows this flow without credentials. Successful discovery ends in `review_required`; approval or rejection completes the task.

No external agent endpoint is enabled in Phase 4, so `AGENT_API_TOKEN` is not required. A future out-of-process executor may add token authentication at the controlled tool boundary; tokens must never be logged.
