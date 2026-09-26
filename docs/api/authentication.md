# Authentication API

Scholarship OS recognizes three principals through one header:

```http
Authorization: Bearer <token>
```

- `usr_...`: an expiring human user session.
- `agt_...`: a scoped agent credential associated with one user workspace.
- `sys_...`: the environment-provided root system key.

Only token prefixes are identifiers. Security comes from cryptographically random token material, hashed persistence, expiry/revocation checks, scope checks, and ownership enforcement.

## System bootstrap

`SYSTEM_API_KEY` is equivalent to root access. Generate it outside the repository, store it in `.env` or a production secret manager, and never expose it to browsers or agents.

```bash
export SYSTEM_API_KEY="sys_$(openssl rand -hex 32)"
```

Register a user with the system key. Registration normalizes the email, hashes the password with Argon2id, and creates a default Master Profile unless `createDefaultProfile` is false.

```http
POST /api/v1/auth/register
Authorization: Bearer sys_<system-secret>
Content-Type: application/json

{
  "email": "alex@example.test",
  "password": "a development password",
  "displayName": "Alex Example",
  "createDefaultProfile": true
}
```

There is no public self-registration endpoint.

## Human sessions

`POST /api/v1/auth/login` is public and accepts `email` and `password`. It returns an opaque `usr_` token once, along with its expiration and safe user data. Only a SHA-256 token digest is stored. SHA-256 is suitable here because session tokens contain at least 256 random bits; passwords use the deliberately expensive Argon2id algorithm instead.

Authenticated user endpoints:

- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/change-password` with `currentPassword` and `newPassword`

Changing a password revokes every other user session while preserving the current session. Logout revokes the current session. Disabled users, expired sessions, and revoked sessions stop authenticating immediately.

Listing and selectively revoking all browser sessions is deferred; Phase 4.5 implements current-session logout plus all-session administrative/password-reset revocation.

The development frontend keeps its user token in `sessionStorage`. This is intentionally abstracted and must be reconsidered—typically in favor of secure HTTP-only cookies—before production browser deployment. If cookies are introduced, CSRF protection must also be revisited. Login requires production rate limiting before public exposure.

## Agent credentials

Only SYSTEM may manage agent credentials:

- `POST /api/v1/admin/agent-credentials`
- `GET /api/v1/admin/agent-credentials`
- `POST /api/v1/admin/agent-credentials/{credentialID}/revoke`

Creation accepts camelCase `userId`, `name`, `scopes`, and optional `expiresAt`. Supported scopes are `research`, `applications`, `questionnaires`, `information_requests`, `profiles`, and `tasks`. The plaintext `apiKey` is returned exactly once; listing exposes only its prefix.

Agent credentials use only `/api/v1/agent/...` routes. Scope and workspace ownership are both required. Agents cannot register users, manage credentials, change passwords, use normal user routes, or approve Application Proposals.

## System user administration

SYSTEM-only endpoints:

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{userID}`
- `PATCH /api/v1/admin/users/{userID}`
- `POST /api/v1/admin/users/{userID}/disable`
- `POST /api/v1/admin/users/{userID}/enable`
- `POST /api/v1/admin/users/{userID}/reset-password`

Administrative password reset revokes every user session. Disabling a user revokes sessions and also prevents associated agent credentials from authenticating.

## Public routes

Only `GET /health` and `POST /api/v1/auth/login` are public. The legacy unauthenticated `POST /api/v1/users` route is no longer mounted.
