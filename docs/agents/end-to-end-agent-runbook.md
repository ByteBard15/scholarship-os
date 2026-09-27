# End-to-end agent runbook

This runbook describes how an authenticated Scholarship OS agent processes research, waits for human approval, creates an Application, performs follow-up preparation, and requests missing information.

It is an operational sequence, not permission to bypass service boundaries. Agents use only the typed `/api/v1/agent` endpoints and never access PostgreSQL, GORM models, user credentials, or the SYSTEM key.

## 1. Authentication and scope requirements

Authenticate every agent request using the assigned credential:

```http
Authorization: Bearer agt_<key>
```

Never print or persist the complete key in research findings, activity summaries, source notes, or repository files.

The agent normally needs these scopes:

- `research`: poll tasks, claim work, save sources and findings, and create proposals.
- `applications`: poll approved proposals, create Applications, and run application prefill.
- `profiles`: read the Effective Profile associated with research or an Application.
- `information_requests`: create, read, and process controlled requests for missing information.
- `writing`: list writing tags, retrieve matching source samples, and save generated drafts.

Agents cannot approve or reject Application Proposals. That action belongs to a USER or SYSTEM principal.

## 2. Research execution lifecycle

### Step 1: Poll queued work

```http
GET /api/v1/agent/research-tasks?limit=10
```

Process the returned tasks in queue order. Do not repeatedly claim many tasks that cannot be completed promptly.

### Step 2: Atomically claim one task

```http
POST /api/v1/agent/research-tasks/{taskID}/start
```

The response contains both the claimed task and its `researchRun`. Retain both IDs for subsequent writes. If the server reports a state conflict, another agent may have claimed the task; return to polling.

Starting a task performs no simulated research. The external agent is responsible for the actual research.

### Step 3: Load all approved inputs

Before researching, retrieve:

```http
GET /api/v1/agent/research-tasks/{taskID}
GET /api/v1/agent/research-tasks/{taskID}/links
GET /api/v1/agent/research-tasks/{taskID}/contexts
GET /api/v1/agent/research-tasks/{taskID}/effective-profile
GET /api/v1/agent/research-tasks/{taskID}/runs
```

Read the task title, description, custom instructions, supplied links, reusable Research Context, and Effective Profile.

Important distinctions:

- `ResearchContext` is reusable question-and-answer guidance supplied before research. It may describe interests, preferences, goals, or existing narrative material. It is not canonical profile data.
- `EffectiveProfile` is the controlled, inherited applicant profile suitable for factual comparisons.
- Never treat an unverified statement in Research Context as stronger evidence than the Effective Profile or Profile Evidence.
- Never modify the Master Profile during research.

### Step 4: Research official sources first

Use this source order:

1. Official scholarship provider
2. Official university
3. Official programme or faculty
4. Official government source
5. Official supervisor profile
6. Trusted secondary source
7. General web source

Record retrieval dates. Preserve vague deadline wording as vague wording. Do not convert “January 2027” or “early December” into an invented timestamp.

### Step 5: Persist each source

```http
POST /api/v1/agent/research-tasks/{taskID}/sources
Content-Type: application/json

{
  "researchRunId": "<run-id>",
  "url": "https://official.example/programme",
  "title": "Official programme page",
  "publisher": "Example University",
  "sourceType": "programme",
  "isOfficial": true,
  "retrievedAt": "2027-01-10T12:00:00Z"
}
```

Persist sources before the claims that depend on them. User-supplied third-party links may be retained as discovery sources but must not be marked official.

### Step 6: Persist typed findings

```http
POST /api/v1/agent/research-tasks/{taskID}/findings
Content-Type: application/json

{
  "researchRunId": "<run-id>",
  "sourceId": "<source-id>",
  "category": "deadline",
  "field": "internationalAdmissionDeadline",
  "value": {
    "rawText": "Applications close in January 2027",
    "deadlineAt": null,
    "datePrecision": "month"
  },
  "confidence": 0.95,
  "verificationStatus": "verified",
  "notes": "Exact day is not published."
}
```

Use `verified`, `supported`, `conflicting`, `unverified`, or `inferred` accurately. Confidence describes confidence in the evidence, not admission probability or proposal priority.

Keep conflicting official facts as separate findings linked to their respective sources. Mark them `conflicting` and require human review.

## 3. Creating a proposal shortlist

A Research Task may produce zero, one, or many viable proposals. Do not collapse distinct programmes or scholarships into one proposal merely because they share an institution.

Create one proposal per viable programme/scholarship combination:

```http
POST /api/v1/agent/application-proposals
Content-Type: application/json

{
  "researchTaskId": "<task-id>",
  "researchRunId": "<run-id>",
  "name": "Example MSc and Graduate Fellowship 2027",
  "country": "Canada",
  "intake": "Fall 2027",
  "intakeYear": 2027,
  "priority": "highest",
  "rank": 1,
  "confidence": 0.92,
  "summary": "Concise opportunity summary.",
  "reasoningSummary": "Why this is a strong evidence-supported fit and what remains unresolved.",
  "proposedInstitution": {},
  "proposedProgramme": {},
  "proposedScholarship": {},
  "sourceIds": ["<official-source-id>"]
}
```

Proposal priority values are:

- `highest`: best actionable fit in the researched shortlist.
- `high`: strong opportunity worth serious consideration.
- `medium`: plausible opportunity with material uncertainty or weaker fit.
- `low`: retained for consideration but clearly less suitable or actionable.

Use `rank` when proposals can be meaningfully ordered. Rank is one-based. Do not rank using unsupported admission probability. Consider verified thematic fit, funding, eligibility, deadline feasibility, and unresolved risks.

The user may approve multiple proposals. Every proposal remains independently reviewable.

### Step 7: Submit the research for review

After all sources, findings, and proposals have been persisted:

```http
POST /api/v1/agent/research-tasks/{taskID}/complete
Content-Type: application/json

{
  "researchRunId": "<run-id>"
}
```

The task becomes `review_required`. Stop research processing at this point. Do not approve a proposal or create an Application before human approval.

## 4. Human approval boundary

The human reviews each proposal and may approve, reject, or leave it pending. Approval records the selected Master or Domain parent profile but does not create an Application.

The agent must not call these user endpoints:

```text
POST /api/v1/application-proposals/{proposalID}/approve
POST /api/v1/application-proposals/{proposalID}/reject
POST /api/v1/application-proposals/{proposalID}/reopen
```

Wait by polling the approved-proposal agent queue rather than attempting to infer approval from task state.

## 5. Creating Applications from approved proposals

### Step 1: Poll approved, unmaterialized proposals

```http
GET /api/v1/agent/application-proposals/approved
```

Only proposals explicitly approved by a human and without an `applicationId` are returned.

### Step 2: Create the Application workspace

```http
POST /api/v1/agent/application-proposals/{proposalID}/application
```

This controlled transaction:

- resolves or creates the Institution, Programme, and Scholarship;
- creates an isolated Application Profile from the human-selected parent;
- creates the Application and default checklist;
- links the retained Research Run, Sources, and Findings;
- transfers supported structured research;
- records outputs, provenance, agent activity, and audit data;
- stores the resulting `applicationId` on the proposal.

Do not retry a successful materialization. If a network failure makes the outcome uncertain, fetch the proposal or poll again. A proposal with `applicationId` has already been materialized.

## 6. Using the writing library

Writing retrieval is intentionally staged so private writing is not loaded unnecessarily.

### Step 1: Inspect available tags

```http
GET /api/v1/agent/writing-tags
```

Correlate tag names with the Application, programme, scholarship, requirements, and questionnaire. Do not fetch every sample.

### Step 2: Retrieve up to four relevant samples

```http
POST /api/v1/agent/writing-samples/match
Content-Type: application/json

{
  "applicationId": "<application-id>",
  "documentType": "motivation_letter",
  "tags": ["biomedical engineering", "medical imaging"],
  "limit": 4
}
```

Use only high-confidence relevant matches. Retrieval confidence is tag overlap, not document quality.

### Step 3: Save generated writing with provenance

```http
POST /api/v1/agent/writing-samples
Content-Type: application/json

{
  "applicationId": "<application-id>",
  "title": "Example Scholarship motivation draft",
  "documentType": "motivation_letter",
  "content": "<generated draft>",
  "tagNames": ["biomedical engineering", "medical imaging"],
  "sourceSampleIds": ["<sample-id-1>", "<sample-id-2>"]
}
```

Generated writing is forced to `origin=agent` and `status=suggested`. Never claim a generated recommendation represents a recommender's opinion, and never fabricate personal facts.

## 7. Application follow-up and prefill

Load the newly created Application and Effective Profile:

```http
GET /api/v1/agent/applications/{applicationID}
GET /api/v1/agent/applications/{applicationID}/effective-profile
```

Then run controlled prefill:

```http
POST /api/v1/agent/applications/{applicationID}/prefill
Content-Type: application/json

{
  "trigger": "manual",
  "regenerate": false
}
```

The prefill agent may:

- fill factual Application Fields supported by evidence;
- create suggested questionnaire answers;
- create Application Tasks;
- create Information Requests for unsupported or conflicting facts.

The prefill agent may not:

- overwrite approved answers by default;
- approve subjective writing;
- modify the Master Profile;
- invent applicant facts;
- submit an application or contact third parties.

## 8. Requesting additional information

Create an Information Request only when the required information cannot be safely resolved from the Effective Profile, accepted research, approved answers, completed prior requests, or other controlled evidence.

```http
POST /api/v1/agent/information-requests
Content-Type: application/json

{
  "applicationId": "<application-id>",
  "questionId": "<question-id>",
  "requestType": "missing_information",
  "title": "Passport expiry date",
  "prompt": "What is the expiry date shown on your current passport?",
  "context": "Required by the application questionnaire.",
  "responseType": "date"
}
```

Do not create duplicate open requests for the same target. The service deduplicates supported targets, but the agent should still inspect existing work.

The lifecycle is:

```text
pending → answered → processing → completed
```

Completed requests must not be reopened or asked again automatically.

### Processing an answered request

Poll or retrieve the request through the controlled endpoint:

```http
GET /api/v1/agent/information-requests/{requestID}
```

Only process it after the user has supplied a response and its status is `answered`:

```http
POST /api/v1/agent/information-requests/{requestID}/process
```

Processing transactionally applies the response to its permitted Application Field or Application Answer, records `information_request` provenance, and only then marks the request `completed`.

If processing fails, do not discard or replace the user's response. Leave the request recoverable and report the operational error without sensitive content.

### Follow-up prefill

After successfully processing new information, rerun prefill conservatively:

```http
POST /api/v1/agent/applications/{applicationID}/prefill
Content-Type: application/json

{
  "trigger": "information_updated",
  "regenerate": false
}
```

This run must skip completed Information Requests and approved or user-edited answers. Regeneration requires an explicit user action.

## 9. Stop conditions

Stop and await human action when:

- research is `review_required`;
- all proposals are pending human review;
- a questionnaire requires unsupported personal information;
- official sources conflict on a material fact;
- a subjective answer needs approval;
- an Information Request is pending user input;
- the requested operation would submit, email, upload, or otherwise act on an external portal.

Fail safely when:

- authentication or the required scope is missing;
- ownership validation fails;
- an Application Proposal is not approved;
- a source cannot be verified;
- the target has already been completed or materialized;
- the API reports a state conflict.

## 10. Minimal end-to-end checklist

```text
[ ] Poll queued Research Tasks
[ ] Claim one task and retain task/run IDs
[ ] Load task, links, contexts, profile, and runs
[ ] Research official sources first
[ ] Persist every source
[ ] Persist typed findings with provenance
[ ] Create every viable proposal with priority/rank
[ ] Complete the task for human review
[ ] Stop until a human approves proposals
[ ] Poll approved proposals
[ ] Create one Application per approved proposal
[ ] Load Application and Effective Profile
[ ] Inspect writing tags and retrieve only relevant samples
[ ] Run conservative prefill
[ ] Save generated writing as suggested with source IDs
[ ] Create Information Requests for missing facts
[ ] Process answered requests transactionally
[ ] Rerun conservative prefill after new information
[ ] Stop before submission or external communication
```
