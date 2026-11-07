# lms-worker

> Scheduled jobs: due-date notices, fines accrual, catalog reindexing

Part of the **LMS Library** distributed system — team `lms-library`, Grupo 2.
Governance and documentation live in [`library-docs`](https://github.com/code-corhuila/library-docs).

User story: **HU-08 Late-Return Penalty System**.

Go, hexagonal. The worker is **polling/cron-triggered**: an internal ticker runs every job every
`POLL_INTERVAL` (15 minutes by default). It does not consume a message broker or event bus. There
is no broker anywhere in this system (`AT-002`, `05-architecture/overview.md`), and
`ADR-009-worker-scheduling-model.md` keeps it that way. Earlier wording that called this service
"event-driven" was generic scaffold text, not the design.

## Structure

```
cmd/worker/              → entry point (main.go): config, logger, clients, jobs, ticker
internal/
├── config/                → configuration from environment variables
├── domain/notice/         → driven ports only (CirculationClient, CatalogClient); no business logic
├── application/usecase/   → DueDateNotices, FinesAccrual, CatalogReindex (one job each, with tests)
└── infrastructure/
    ├── httpclient/          → CirculationClient, CatalogClient, internal HS256 token (with tests)
    ├── scheduler/           → the ticker that runs every job, isolated per tick (with tests)
    └── logger/              → structured JSON logger (zap)
deploy/Dockerfile        → multi-stage build; the image exposes no ports
```

No database and no migrations: the service is stateless (`ADR-006`, `ADR-009`). The only dedupe
state (`notifiedAt` on a loan) lives in `lms-circulation-api`'s own record.

## How it runs

1. `main.go` loads the configuration and fails fast if `INTERNAL_JWT_SECRET` is missing or
   `POLL_INTERVAL` is not a valid duration.
2. It builds the circulation and catalog clients and registers three jobs:
   `due-date-notices`, `fines-accrual` and `catalog-reindex`.
3. The ticker runs every job once at start and then on each interval. A failing job is logged and
   does not stop the others.
4. `SIGINT` or `SIGTERM` stops the ticker and the process exits cleanly.

Every call to another service carries a one-minute HS256 service token signed with
`INTERNAL_JWT_SECRET`. Any status other than 200 counts as a failure.

## Configuration

Every variable is listed in `.env.example`.

| Variable | Default | Purpose |
|---|---|---|
| `INTERNAL_JWT_SECRET` | none (required) | Secret for the service-to-service token; must match the other services |
| `CIRCULATION_SERVICE_URL` | `http://circulation-service:8080` | Base URL of `lms-circulation-api` |
| `CATALOG_SERVICE_URL` | `http://catalog-service:8080` | Base URL of `lms-catalog-api` |
| `POLL_INTERVAL` | `15m` | Time between ticks (Go duration, ADR-009) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |

## Run, test and build

```bash
make test         # go test ./...
make test-cover   # coverage report
make dev          # go run ./cmd/worker/...
make build        # bin/worker
```

Docker:

```bash
docker build -f deploy/Dockerfile -t lms-worker:dev .
docker run --rm -e INTERNAL_JWT_SECRET=dev-secret lms-worker:dev
```

Without the other services running, each job logs `job failed` with `no such host`, and the
worker keeps running. That is the expected behaviour.

## Known integration gap

The worker calls endpoints that `lms-circulation-api` and `lms-catalog-api` do not expose yet:

| Route | Service | Effect until it exists |
|---|---|---|
| `POST /api/v1/loans/{id}/notify` | `lms-circulation-api` | `MarkNotified` fails |
| `POST /api/v1/loans/accrue-fines` | `lms-circulation-api` | `FinesAccrual` fails |
| `POST /api/v1/books/reindex` | `lms-catalog-api` | `CatalogReindex` fails |
| `GET /api/v1/loans/overdue` filtered by "not yet notified" | `lms-circulation-api` | the route exists, the filter does not |

Building those routes is outside this repository. They must answer 200.

## Migration scope

**From scratch**: nothing in `lms-library` covered this before. Delivered in PRs #2 (scaffold),
#3 (ports and use cases), #4 (HTTP clients and ticker) and #5 (entry point and Dockerfile).

The full map lives in `library-docs`.

---

## Branching

Three permanent branches. **None of them accepts a direct commit** — you enter through a child
branch and leave through a Pull Request.

```
develop  <--PR--  feat/... fix/... chore/...
qa       <--PR--  qa/...
main     <--PR--  release/...  hotfix/...
```

Promotion happens **by re-application** (`git cherry-pick -x`), never by merging one permanent
branch into another: `merge develop -> qa` and `merge qa -> main` do not exist in this model.

`main` requires **1 approval from `ariel5253`**. On `develop` and `qa` the team sets its own review
rule.

Full policy: `00-governance/branching-policy.md` in `library-docs`.
