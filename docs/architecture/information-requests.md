# Information Requests

An `InformationRequest` is the safe alternative to guessing. It targets an Application Question or Application Field, describes the missing information and expected response type, and deduplicates open requests for the same target.

Lifecycle: `pending → answered → processing → completed`. Processing transactionally applies the latest response, records provenance, and only then completes the request. Every submitted response is retained in `information_request_responses`. Completed requests stay closed unless explicitly reopened.
