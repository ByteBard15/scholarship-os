# Application Proposals API

- `GET /api/v1/application-proposals?userId=&status=`
- `GET /api/v1/application-proposals/{proposalID}`
- `POST .../{proposalID}/approve` with `{ "parentProfileId": "..." }`
- `POST .../{proposalID}/reject`
- `POST .../{proposalID}/reopen`

Approval is the only research-discovery path that creates an Application. It is transactional and cannot be repeated.
