package application

// TaskInstructions gives new checklist items a usable starting guide. An agent
// should replace generic references with the verified programme-specific links.
func TaskInstructions(title string) string {
	switch title {
	case "Research opportunity", "Verify research findings":
		return "1. Open this application's Research tab and follow the recorded official programme and funding sources.\n2. Verify the current intake, research degree, supervisor requirements and funding coverage.\n3. Record source dates and unresolved conflicts; ask the programme office about missing details.\nDone when: the programme route and funding conditions are supported by current sources, with uncertainties recorded."
	case "Verify eligibility":
		return "1. Read the official entry criteria linked in Research or Overview.\n2. Compare degree, transcript modules and language evidence with the resolved Application Profile.\n3. Ask admissions about uncertain equivalence; do not invent a GPA conversion or language waiver.\nDone when: each requirement has evidence or a recorded clarification request."
	case "Verify deadlines":
		return "1. Open the official admissions and awards pages linked to this application.\n2. Record separate application, document, supervisor, nomination and scholarship deadlines.\n3. Retain the source wording and timezone; label unpublished dates as unconfirmed.\nDone when: the applicable deadlines and required preceding steps are documented."
	case "Review required documents":
		return "1. Open the official programme checklist from the Research sources.\n2. Make a file list covering transcripts, grading scale, degree certificate, CV, statement, references and any required language evidence.\n3. Check file formats, translation rules and whether official documents are required now or only after an offer.\nDone when: each required document is ready or has a named outstanding action."
	case "Prepare application profile":
		return "1. Review this application's resolved profile in the Profile tab.\n2. Check dates, degree details, publication status and the evidence behind employment claims.\n3. Tailor supported descriptions to the programme without changing the parent profile for an application-specific preference.\nDone when: the application profile is accurate and ready for CV and statement preparation."
	case "Prepare CV":
		return "1. Review the programme's CV requirements and your Application Profile.\n2. Draft a concise academic CV with education, research, supported technical work and complete citations.\n3. Separate published papers, under-review work and portfolio articles; proofread before export.\nDone when: the applicant has reviewed the final PDF and it matches the programme requirements."
	case "Request references":
		return "1. Confirm the required number and type of referees on the official admissions page.\n2. Ask the referees to participate and share the CV, programme details, deadline and any suggested letter in Writing Library.\n3. Have each referee review their wording and submit through the university's invitation process.\nDone when: the portal confirms receipt of each required reference, not merely when a draft exists."
	case "Complete questionnaires":
		return "1. Copy the real portal prompts and limits into the Questionnaires tab.\n2. Review suggested answers against the Application Profile and Writing Library sources.\n3. Resolve unsupported facts through Information Requests and approve subjective writing yourself.\nDone when: every required prompt has a reviewed answer within its stated limit."
	case "Resolve missing information":
		return "1. Open this application's Information Requests.\n2. Supply each requested fact or document with a source, or state that it is unknown.\n3. Process the response through the controlled workflow and verify that it reached the intended field.\nDone when: required requests are applied or explicitly resolved; an unanswered request is not completed."
	case "Review and submit application":
		return "1. Open the official application portal linked in Research or Overview; verify the programme and intake.\n2. Check documents, references, supervisor arrangements, funding and all form answers.\n3. The applicant reviews declarations, pays any required fee and submits personally. Save the confirmation and monitor the portal.\nDone when: the university has issued a submission confirmation and outstanding document requests are tracked."
	default:
		return "1. Review the task's official or internal source.\n2. Complete the stated action and keep supporting evidence.\n3. Record any blocker before marking the task done.\nDone when: the action is verified by its relevant document or portal confirmation."
	}
}
