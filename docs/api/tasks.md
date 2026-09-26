# Application tasks API

- `GET|POST /api/v1/applications/{applicationID}/tasks`
- `PATCH|DELETE /api/v1/applications/{applicationID}/tasks/{taskID}`
- `POST .../{taskID}/complete`
- `POST .../{taskID}/reopen`

Tasks support one parent, due dates, priority, task type, ordering, and completion timestamps. Parents must belong to the same application; self-parenting and cycles are rejected. `ExternalTaskReference` persists a future provider mapping, but no external synchronization is implemented.
