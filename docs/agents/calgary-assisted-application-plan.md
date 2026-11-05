# Assisted applications: Calgary case study

This is the execution design for a future portal adapter, not an enabled external submission feature. The shipped foundation stores task criteria and evidence. Current internal prefill, profiles and suggested Writing drafts remain available. No Calgary account has been created, portal form completed, email sent or payment made by this change.

## What an agent should be able to do

Each programme needs a verified capability assessment: `manual_only`, `assisted_draft` or, eventually, `reviewed_submission`. Public admissions instructions do not establish permission or technical support for automated portal interaction. Calgary is currently **unassessed for external automation**; its instructions describe saved online drafts but no supported application API has been verified.

An adapter should first inspect the actual form and permitted access method. With applicant authorization and a supported session, it could create a draft, fill supported factual fields, attach reviewed files, save progress and produce an audit package. It must pause for unresolved facts, account verification, CAPTCHA, MFA, applicant declarations, payment or unsupported controls. A draft save must not be represented as university submission.

## Calgary task sequence and evidence

| Step | Agent preparation / future permitted action | Applicant or external dependency | Completion evidence |
| --- | --- | --- | --- |
| 1. Identify the workspace | Locate the existing Calgary Application and resolved Application Profile | Resolve the older Calgary profile versus newly linked workspace before portal work | Internal application/profile IDs and selected programme/intake |
| 2. Access the account | Open the official account/portal entry point; reuse an existing account | Applicant completes account identity checks and provides an authorized session through a future secure integration | eID and portal URL; no password, token or session cookie |
| 3. Begin a draft | Select Biomedical Engineering, thesis MSc, Fall 2027; verify the live choices; save a draft if supported | Applicant confirms programme and intake if ambiguous | External draft/application ID, stable portal URL and save confirmation |
| 4. Prepare forms and documents | Fill sourced facts from EffectiveProfile; prepare CV and statement as suggested Writing drafts; check required files against live prompts | Applicant reviews subjective writing and supplies unknown facts | Writing IDs, reviewed document references, saved sections and unresolved Information Requests |
| 5. Supervisor and referees | Prepare tailored outreach and referee materials | Applicant authorizes sending; faculty confirms supervision/funding; referees approve and submit their own letters | Actual correspondence/confirmation and portal receipt, not draft existence |
| 6. Audit admission | Assemble the exact saved answers, file versions, intake, supervisor state, fee shown by portal, declarations and outstanding items | Applicant reviews this concrete package | Audit-package version/hash and explicit user decision in the future review system |
| 7. Submit admission | Future supported adapter only after scoped approval; current workflow is applicant submission | Applicant performs declarations/payment/submission in current implementation | University-issued submission confirmation, submission date and UCID when issued |
| 8. Prepare GAC | After admission submission and UCID, open Student Centre → Apply for Awards → Graduate Award Competition; check the award-specific evidence | Programme may need to upload transcripts; applicant reviews award answers | Separate GAC draft ID, file references, eligibility-source notes |
| 9. Audit and submit GAC | Prepare a separate audit package; do not reuse admission approval | Applicant submits under current workflow | GAC-specific submission confirmation and outstanding actions |

Official account entry: https://acctman.ucalgary.ca/register/

Official admission instructions: https://grad.ucalgary.ca/future-students/graduate/admissions/how-apply

Admission portal: https://apply.ucalgary.ca/StudentAdmission/Login.aspx?AppType=A

GAC instructions: https://grad.ucalgary.ca/awards-and-funding/award-opportunities/graduate-award-competition

Student Centre: https://my.ucalgary.ca/

Official instructions checked 5 November 2026. They describe eID login, saving application progress and separate submission. The GAC page directs applicants through Student Centre and distinguishes its transcript requirements from admission. Do not assume an unofficial admission upload satisfies GAC.

Deadline source: Andrea Cortes's 4 November 2026 reply, saved in Writing as **Calgary BME Programme Reply — 4 November 2026**. BME internal GAC deadline **11 January 2027**; international admission deadline **1 March 2027**, with supervisor confirmation by then. Admission must be submitted well before the GAC deadline to obtain the UCID. Closing times were not supplied; no exact timestamp is inferred.

## Implementation stages for the future runner

1. Add per-programme capability records with official evidence, permitted domains, supported actions and the date verified. Avoid a global “can automate” assumption.
2. Add secure browser/API sessions outside task evidence, credential references rather than credentials, resumable execution runs, leases and action-level idempotency keys. On ambiguous network failure, reconcile the portal before retrying account creation or submission.
3. Add typed actions such as start draft, save section, attach document and inspect confirmation. Resolve EffectiveProfile and record source versions. Human-only actions must stop the run with a concrete handoff.
4. Add immutable audit packages with exact field values, file hashes/versions, programme, intake, destination, current portal state, fee and declarations. A USER-only approval must bind to that version, action and destination, expire and be invalidated by any change. Proposal approval is not portal submission approval. Agents cannot approve themselves.
5. Only after review endpoints, enforced approval checks, tested adapters and explicit policy support exist, consider a submission action. Admission and funding submissions require separate approvals. Keep a manual route available.
6. Capture actual portal receipts as completion evidence. Do not infer success from a click, prepared PDF, saved draft or paid fee alone.

Current project instructions prohibit autonomous external submission. This design does not change that rule or implement a submission executor. The human-reviewed flow needs the explicit implementation and policy changes above before it can run.
