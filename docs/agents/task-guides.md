# Actionable application task guides

Application tasks support `description`, `sortOrder`, `parentTaskId`, status and an optional exact dueAt. The Tasks tab now displays descriptions and clickable HTTP(S) references rather than titles alone. Numbered text and line breaks are preserved; Markdown links and bare URLs are supported. Content is escaped and arbitrary HTML or unsafe link protocols are not executed.

When creating a task, include:

1. A concrete action title.
2. Numbered steps that can be completed without reconstructing the workflow from chat.
3. The verified official URL or exact named internal resource, such as a Writing Library draft.
4. A “Done when” criterion describing evidence of completion.
5. Dependencies, source and deadline notes. Distinguish account login IDs, applicant/student IDs, admission deadlines and scholarship deadlines.

Do not mark “contact supervisor” complete merely because a draft exists. Do not mark a university submission complete because an internal Application was created. Reference drafts need referee approval and the university's own submission process.

Existing tasks should be updated in place, preserving title meaning, status, hierarchy, priority and exact dates. The PATCH task DTO uses replacement semantics for several fields, so carry forward parentTaskId, dueAt, priority, taskType and source rather than omitting them unintentionally. Do not overwrite the applicant's progress. Use sortOrder for the intended sequence.

If an email provides a date without a time or timezone, retain that date in the guide; do not invent a dueAt timestamp. A personal preparation target is separate from a university cutoff.

All newly created default checklist tasks now include starting guidance and completion criteria. Agents should tailor the guidance with the programme-specific official links before handing the checklist to the applicant.

A read-only `GET /api/v1/agent/applications` route, requiring `applications`, lists the authenticated workspace's applications and helps agents find the correct existing checklist. Deploy this update before using it. Do not bootstrap another approved proposal merely to discover an existing application.

No schema migration is needed for this display/default-guide update.

Task titles and “View details and links” now open an accessible dialog with the full guide. Completion checkboxes remain separate from opening details.

The Writing Library uses server pagination with 10, 20, 50 or 100 samples per page. Its page and pageSize are retained in the URL. `GET /api/v1/writing-samples?page=1&pageSize=20` returns `meta.count`, `total`, `page`, `pageSize` and `totalPages`; requests without pagination parameters retain the previous collection contract. Results are ordered by updated time and ID for deterministic ties. The paginated frontend and backend should be deployed together; no migration is required.
