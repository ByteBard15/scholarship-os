# Creating applications from legacy approved proposals

Older research records can contain `proposedInstitution`, `proposedProgramme` and `proposedScholarship` as JSON strings, for example `"The American University in Cairo"`, rather than structured objects. Previously, materializing these approved proposals returned HTTP 500 while decoding catalogue metadata.

The bootstrap now interprets a legacy string as the approved catalogue name. For an institution, a missing country is obtained from the approved proposal's country; if neither exists, creation returns a validation error. Other unspecified metadata remains absent. Existing structured proposals and supplied catalogue IDs continue to work. Approval records and source provenance are retained unchanged. Existing catalogue records are reused, and application/profile creation remains transactional.

Deploy the API before retrying the existing approved proposal through `POST /api/v1/agent/application-proposals/{proposalID}/application`. No migration, new proposal or renewed approval is required. Do not create a replacement proposal or manually insert an Application to work around the error.

Validation includes unit coverage of legacy/structured names and a PostgreSQL regression using real migrations that creates applications from all three legacy string fields, checks catalogue reuse and verifies the approved proposal was not rewritten. Run the workflow integration test with `WORKFLOW_TEST_DATABASE_URL` pointing to a disposable localhost database.
