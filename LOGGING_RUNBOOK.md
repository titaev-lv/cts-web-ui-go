# Web UI Logging Runbook

## Purpose

Operational checks for Web UI logging in Docker:
- `request_id` propagation (`X-Request-ID`)
- JSON logs in `docker logs`
- `access.log` / `error.log` presence and rotation
- quick diagnostics for write-permission issues

## Quick Validation

### 1) Service is running

```bash
docker compose up -d web-ui
docker compose ps web-ui
```

### 2) Request ID propagation

```bash
curl -sSI -H 'X-Request-ID: runbook-rid-001' http://127.0.0.1/login | grep -i '^X-Request-Id:'
```

Expected: response header `X-Request-Id: runbook-rid-001`.

### 3) Access log fields in stdout

```bash
docker logs --since 2m ct-system-web-ui 2>&1 | grep -E 'HTTP access|request_id' | tail -n 20
```

Expected JSON fields:
- `request_id`
- `module`
- `method`
- `path`
- `status`
- `latency_ms`
- `user_id`

### 4) Log files and rotation

```bash
ls -lh services/web-ui-go/logs/
ls -1 services/web-ui-go/logs | grep -E '^(access|error)\.log(\.|$)|^(access-|error-)'
```

Expected:
- active files `access.log`, `error.log`
- rotated artifacts (`access-*`, `error-*`, optionally `.gz`)

## Debug/Release Regression Profile Check

### A) Switch to release temporarily

```bash
cp services/web-ui-go/config/config.yaml services/web-ui-go/config/config.yaml.tmp
sed -i 's/^  mode: .*/  mode: release  # debug, release, test/' services/web-ui-go/config/config.yaml
docker compose up -d --force-recreate web-ui
```

### B) Validate release startup and request flow

```bash
docker logs --since 1m ct-system-web-ui 2>&1 | grep '"mode":"release"'
curl -sSI -H 'X-Request-ID: runbook-rid-release' http://127.0.0.1/login | grep -i '^X-Request-Id:'
```

### C) Restore debug mode

```bash
mv services/web-ui-go/config/config.yaml.tmp services/web-ui-go/config/config.yaml
docker compose up -d --force-recreate web-ui
docker logs --since 1m ct-system-web-ui 2>&1 | grep '"mode":"debug"'
```

## Troubleshooting

### Logger init fails: `/app/logs is not writable`

Symptoms in container logs:
- `Failed to initialize logger: ... log dir /app/logs is not writable`

Checks:
```bash
ls -ldn services/web-ui-go/logs
docker exec ct-system-web-ui sh -lc 'id && test -w /app/logs && echo writable'
```

Recommended fix:
- run Web UI with explicit non-root user mapping in compose
- ensure bind-mount owner/mode allows write (`750` dir, `640` files)

### No records in `docker logs`

Checks:
```bash
docker compose ps web-ui
docker logs --since 5m ct-system-web-ui
```

Ensure logging config uses `output: both` in `services/web-ui-go/config/config.yaml`.
