# GlitchMesh Local Lab

Small reproducible distributed systems lab for end-to-end GlitchMesh testing. All service-to-service traffic goes through the **real** GlitchMesh runtime proxy — no mocks.

## Architecture

```text
traffic-gen (local)
        ↓
   gateway :8080
        ↓
 glitchmesh :9000   ← fault injection + reverse proxy
        ↓
 feed-service :8081 (internal only)
```

**Critical rule:** gateway never calls `feed-service` directly. It always calls GlitchMesh:

```text
http://glitchmesh:9000/redirect/feed-service/feed
```

GlitchMesh reads `lab/proxy.yaml`, applies faults, then forwards to `http://feed-service:8081/feed`.

## Quick Start

### 1. Start the lab stack

From the repo root:

```bash
cd lab
docker compose up --build
```

Services:

| Service       | Port (host) | Role                                      |
|---------------|-------------|-------------------------------------------|
| gateway       | 8080        | External entrypoint                       |
| glitchmesh    | 9000        | Real runtime proxy + fault injector       |
| feed-service  | (internal)  | Mock downstream                           |

### 2. Smoke test (manual)

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/feed
curl http://localhost:9000/metrics
```

### 3. Generate traffic (local)

With the stack running:

```bash
go run ../cmd/trafficgen/main.go -url http://localhost:8080/api/feed -concurrency 5 -count 50
```

Flags:

- `-url` — gateway URL (default `http://localhost:8080/api/feed`)
- `-concurrency` — parallel workers
- `-count` — total requests
- `-timeout` — per-request timeout

Watch logs in the `docker compose` terminal. You should see:

- `[gateway]` — incoming + GlitchMesh upstream calls
- GlitchMesh runtime — proxy timing + fault messages
- `[feed-service]` — downstream handling

### 4. Change faults

Edit `lab/proxy.yaml` (or copy from `lab/scenarios/`), then restart runtime:

```bash
docker compose restart glitchmesh
```

Example: 500ms latency on every request to `feed-service` (see `scenarios/latency-500ms.yaml`).

`probability: 0` means fault applies on every request when `enabled: true`.

## End-to-End Experiment Flow

1. Define your experiment scenarios in a config YAML file (or copy/modify one from `lab/scenarios/`).
2. Copy the YAML config into `lab/proxy.yaml`.
3. `docker compose restart glitchmesh` — runtime reloads config at startup.
4. Run `traffic-gen` against gateway.
5. Observe latency/errors in traffic-gen output and container logs.
6. Check `curl http://localhost:9000/metrics` for fault counters.

## Docker Networking

Containers use Compose service names as hostnames:

- `glitchmesh` — runtime
- `gateway` — entry gateway
- `feed-service` — mock feed API

Gateway env (set in `docker-compose.yml`):

```env
GLITCHMESH_URL=http://glitchmesh:9000
FEED_SERVICE_NAME=feed-service
```


## Teardown

```bash
docker compose down
```

## Future Extensions

- Add Redis/Postgres as new services; route through GlitchMesh in `proxy.yaml`
- Add more dummy services under `services/`
- Prometheus/Grafana sidecars

