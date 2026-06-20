# Docker

## Quick Start

```bash
docker pull veloper/ticketer
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  veloper/ticketer
```

Open http://localhost:8300/login and sign in.

## Build from Source

```bash
docker build -t ticketer .
```

The Dockerfile builds both `ticketer` and `tktrctl` binaries in a multi-stage build. The final image is Alpine-based and includes both binaries.

## Docker Compose

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports:
      - "8300:8300"
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

volumes:
  ticketer-data:
```

The database persists in the `ticketer-data` volume at `/data/ticketer.db` (set via `ENV TICKETER_DB_PATH` in the Dockerfile).

## Using tktrctl

The Docker image includes `tktrctl` — use it for scripting and automation:

```bash
# One-off commands against a running container
docker compose exec ticketer \
  tktrctl projects create "Game" GAME

docker compose exec ticketer \
  tktrctl issues create GAME "Fix login" --type bug --priority 1
```

## Automated Setup Service

For first-time bootstrapping, use a separate service with a `setup` profile:

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports: ["8300:8300"]
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

  setup:
    image: veloper/ticketer
    profiles: ["setup"]
    environment:
      TICKETER_HOST: http://ticketer:8300
      TICKETER_PAT: pat_admin
    depends_on:
      ticketer:
        condition: service_started
    command: >
      tktrctl projects create "Asteroid Game" ASTEROID-GAME &&
      tktrctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
```

```bash
docker compose --profile setup run setup
```

This creates the admin user (via the main service) and then seeds projects and issues automatically.

## Environment Variables

| Env var | Default | Description |
|---------|---------|-------------|
| `TICKETER_ADMIN_USERNAME` | — | Admin username **(required)** |
| `TICKETER_ADMIN_PAT` | — | Admin PAT **(required)** |
| `TICKETER_HOST` | `""` | Listen host |
| `TICKETER_PORT` | `"8300"` | Listen port |
| `TICKETER_DB_PATH` | `/data/ticketer.db` | SQLite database path |
