# Future Agent Architecture

This directory documents controlled agent behavior. Phase 3 provides typed backend tool and research-provider abstractions but no autonomous or live web agent runtime.

Agents will interact with Scholarship OS only through controlled, typed backend tools or APIs. They must not receive database credentials or unrestricted SQL access. Tool authorization, ownership checks, provenance, review, and auditability belong at the backend boundary.

Agent categories may include scholarship researchers, eligibility analyzers, profile-tailoring agents, CV agents, application planners, and motivation-letter agents. Research output must be stored as sourced findings and reviewed before application.

Agents must not mutate canonical profile facts without an explicit controlled workflow. Suggested changes remain proposals until a user accepts them. Agents never execute SQL, invent deadlines, assess admission probability, or submit applications. Do not implement live providers or external automation unless explicitly requested in a later phase.

Agents authenticate only with an assigned `Authorization: Bearer agt_...` credential and use dedicated `/api/v1/agent` routes. They must never use human credentials, request or use `SYSTEM_API_KEY`, call user-management endpoints, approve an Application Proposal, or attempt to bypass their scopes or associated user workspace.
