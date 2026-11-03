# Queue imported tasks through the agent API

A workspace-associated agent with `research` may inspect existing tasks and queue selected drafts after the API update is deployed. No migration is required.

- `GET /api/v1/agent/research-tasks/list?status=draft&page=1&pageSize=100`: normal paginated listing, including filters available to the user listing. Ownership is enforced even if a different userId is supplied. Omit status to include all states.
- `POST /api/v1/agent/research-tasks/{taskID}/queue`: queue an owned draft or ready task with an owned, valid selected profile. Already queued tasks return successfully without rewriting state. Running, completed, failed and cancelled tasks return 409. A conditional database update prevents a stale queue request from resetting a worker's state.
- `GET /api/v1/agent/research-tasks?limit=50`: existing worker poll; still returns queued tasks only.

Match an import against existing tasks before queueing. Do not import another copy. An exact title, task type and profile match is a useful starting point; review ambiguous duplicates. Queue requests do not begin research, approve proposals or create applications. Agent tools expose matching typed list and queue operations.

The helper `scripts/queue-imported-research.py` reads all task pages before mutation and matches title, task type and profile ID to an import file. Instructions disambiguate repeated titles. Missing or ambiguous matches stop all writes. It never creates tasks or resets running/completed work. Preview is the default; use `--apply` only to queue matched eligible tasks.

```sh
python3 scripts/queue-imported-research.py tmp/new-tasks-import/research-tasks.json --credentials-file tmp/requests.http --apply --report tmp/new-tasks-import/queue-results.json
```

Alternatively provide `SCHOLARSHIP_OS_BASE_URL` and `SCHOLARSHIP_OS_AGENT_KEY` through the environment. Do not place credentials in command arguments. A failed run can be retried after inspecting its report; already queued tasks are unchanged. Tests: `python3 -m unittest discover -s scripts/tests -v`.
