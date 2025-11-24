# Repository Guidelines

## Project Structure & Module Organization
- Core microservices live under `cmd/` (`api`, `user`, `video`, `interaction`, `follow`, `chat`). Each service follows the same layout: `dal/` for data stores, `service/` for business logic, `pack/` for DTO translation, and RPC glue generated from `idl/`.
- Shared utilities (`pkg/constants`, `pkg/errno`, `pkg/middleware`, etc.) host cross-service helpers. Keep new shared code here rather than duplicating logic.
- Configuration assets are under `config/` (YAML, loader code, SQL migrations). Deployment artifacts mirror this in `deploy/`.
- Tests sit next to the code they cover (e.g., `cmd/video/service/feed_video_test.go`) and integration suites reside in `test/<service>/`.
- Generated Kitex/Hertz code is in `kitex_gen/`; never edit these files manually—regenerate from `idl/` instead.

## Build, Test, and Development Commands
- `make env-up` — start the Docker Compose stack (MySQL, Redis, etcd, Kafka, observability). Run first when bootstrapping the repo.
- `make <service>` (e.g., `make api`) — build and run a specific microservice; append `ci=1` to skip auto-run.
- `make build-all` — compile all microservices into `output/` without starting them.
- `go test ./...` — execute all Go unit tests from the repo root.
- `sh docker-run.sh [service]` — run prebuilt Docker images; omit `[service]` to start the entire suite.

## Coding Style & Naming Conventions
- Go code must be `gofmt`-ed, use camelCase for locals, and export identifiers with GoDoc-friendly comments where appropriate.
- Keep business logic inside `service/`, serialization/mapping in `pack/`, and persistence in `dal/`. Middleware or shared helpers belong in `pkg/`.
- Avoid editing generated `kitex_gen/` or router files directly; re-run the Hertz/Kitex generators after IDL updates.

## Testing Guidelines
- Leverage Go’s standard `testing` package; filenames end with `_test.go` and reside beside the implementation.
- Prefer table-driven tests for service logic. Mock external RPCs or DAL calls to keep tests deterministic.
- Run `go test ./...` (or targeted packages) before submitting changes. Document skipped or flaky tests in the PR if unavoidable.

## Commit & Pull Request Guidelines
- Follow the existing concise verb-phrase style, optionally scoped (e.g., `feat: 删除日志钩子`). Keep messages in English or bilingual when helpful.
- Each PR should include: summary of changes, affected service(s), test evidence (`go test ./...` output or CI link), and deployment/config notes (new env vars, schema changes, etc.).
- Rebase before merging, avoid force-pushing over reviewed commits, and group logical changes into focused commits.

## Security & Configuration Tips
- Never hardcode secrets; services expect them from the config center (e.g., `ETCD_ADDR`, DB credentials). Update `config/sql/` migrations when schemas change.
- Upload limits (e.g., `constants.MaxVideoSize`) and OSS prefixes come from `config/config.yaml`; validate these when touching API surfaces.
- Deleting observability components (logging, Prometheus, Grafana) should not impact business logic, but document such removals for future reference.
