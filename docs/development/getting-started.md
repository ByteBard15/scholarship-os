# Getting started

1. Install Go 1.23+, Node.js 20+, Docker, npm, and the golang-migrate CLI.
2. Run `cp .env.example .env` from the repository root. Generate a secure development key with `printf 'sys_%s\n' "$(openssl rand -hex 32)"` and assign it to `SYSTEM_API_KEY` in `.env`.
3. Run `make setup` to download Go and npm dependencies.
4. Run `make db-bootstrap` to start PostgreSQL and apply every migration using Docker. Alternatively run `make db-up`, wait for `docker compose ps` to report PostgreSQL healthy, and run `make migrate-up` with a locally installed golang-migrate CLI.
5. Optionally run `make seed` for fictional Phase 3 catalog records.
6. Run `make api`; verify `curl http://localhost:8080/health` returns an OK response.
7. In another terminal run `make web` and open `http://localhost:5173`.

Create a development user with the root-equivalent system key. Never place this key in frontend configuration:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Authorization: Bearer sys_<system-secret>' \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.test","password":"a development password","displayName":"Alex Example"}'
```

Open `http://localhost:5173/login` and sign in with that email and password. The browser stores the opaque development session in `sessionStorage`; this is not the final production browser credential strategy. The values above are fictional.

The gitignored `tmp/requests.http` contains manual Phase 2 and Phase 3 flows for JetBrains HTTP Client-compatible editors. The Phase 3 section creates catalog records and an application profile/application, exercises each child resource, runs the deterministic research provider, reviews findings, applies accepted data, and checks dashboard/deadline summaries.

Run `make test` before committing. `make lint` runs Go vet and the frontend type checker. Use `make fmt` for Go and frontend formatting. The API never invokes AutoMigrate: every schema change needs matching numbered `.up.sql` and `.down.sql` files.
