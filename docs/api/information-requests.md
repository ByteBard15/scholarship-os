# Information Requests API

- `GET /api/v1/information-requests?userId=&status=&applicationId=&requestType=`
- `GET /api/v1/information-requests/{requestID}`
- `POST .../{requestID}/respond` with `responseText` or `responseValue`
- `POST .../{requestID}/process`
- `POST .../{requestID}/reopen|cancel`

Responding appends immutable response history and moves the request to `answered`. Processing applies it to the controlled target and only then marks it `completed`.
