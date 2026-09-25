# Future Agent Architecture

This directory documents future agent behavior; it contains no executable agents in the current phase.

Agents will interact with Scholarship OS only through controlled, typed backend tools or APIs. They must not receive database credentials or unrestricted SQL access. Tool authorization, ownership checks, provenance, review, and auditability belong at the backend boundary.

Possible future categories include scholarship researchers, eligibility analyzers, profile-tailoring agents, CV agents, application planners, and motivation-letter agents.

Agents must not mutate canonical profile facts without an explicit controlled workflow. Suggested changes should remain proposals until a user accepts them. Do not implement these agents or their future domains unless the task explicitly requests the next phase.
