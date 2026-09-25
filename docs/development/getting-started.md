# Getting started

1. Install Go 1.23+, Node.js 20+, Docker, npm, and the golang-migrate CLI.
2. Run `cp .env.example .env` from the repository root.
3. Run `make setup` to download Go and npm dependencies.
4. Run `make db-up` and wait for `docker compose ps` to report PostgreSQL healthy.
5. Run `make migrate-up`.
6. Run `make api`; verify `curl http://localhost:8080/health` returns an OK response.
7. In another terminal run `make web` and open `http://localhost:5173`.

Create a development user with:

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.test","displayName":"Alex Example"}'
```

Add the returned ID as `VITE_DEMO_USER_ID` in `.env`, restart Vite, and create profiles through the API. The values above are fictional.

Run `make test` before committing. `make lint` runs Go vet and the frontend type checker. Use `make fmt` for Go and frontend formatting. The API never invokes AutoMigrate: every schema change needs matching numbered `.up.sql` and `.down.sql` files.
