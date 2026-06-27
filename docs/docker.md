# Docker

## Quick Start

```bash
docker pull veloper/goard
docker run -p 8300:8300 \
  -e GOARD_ADMIN_USERNAME=admin \
  -e GOARD_ADMIN_PAT=pat_admin \
  veloper/goard
```

Open http://localhost:8300/login and sign in.

## Build from Source

```bash
docker build -t goard .
```

The Dockerfile builds both `goard` and `goardctl` binaries in a multi-stage build. The final image is Alpine-based and includes both binaries.

## Docker Compose

```yaml
services:
  goard:
    image: veloper/goard
    ports:
      - "8300:8300"
    environment:
      GOARD_ADMIN_USERNAME: admin
      GOARD_ADMIN_PAT: pat_admin
    volumes:
      - goard-data:/data

volumes:
  goard-data:
```

The database persists in the `goard-data` volume at `/data/goard.db` (set via `ENV GOARD_DB_PATH` in the Dockerfile).

## Using goardctl

The Docker image includes `goardctl` — use it for scripting and automation:

```bash
# One-off commands against a running container
docker compose exec goard \
  goardctl projects create "Game" GAME

docker compose exec goard \
  goardctl issues create GAME "Fix login" --type bug --priority 1
```

## Automated Setup Service

For first-time bootstrapping, use a separate service with a `setup` profile:

```yaml
services:
  goard:
    image: veloper/goard
    ports: ["8300:8300"]
    environment:
      GOARD_ADMIN_USERNAME: admin
      GOARD_ADMIN_PAT: pat_admin
    volumes:
      - goard-data:/data

  setup:
    image: veloper/goard
    profiles: ["setup"]
    environment:
      GOARD_HOST: http://goard:8300
      GOARD_PAT: pat_admin
    depends_on:
      goard:
        condition: service_started
    command: >
      goardctl projects create "Asteroid Game" ASTEROID-GAME &&
      goardctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
```

```bash
docker compose --profile setup run setup
```

This creates the admin user (via the main service) and then seeds projects and issues automatically.

## Environment Variables

| Env var | Default | Description |
|---------|---------|-------------|
| `GOARD_ADMIN_USERNAME` | — | Admin username **(required)** |
| `GOARD_ADMIN_PAT` | — | Admin PAT **(required)** |
| `GOARD_HOST` | `""` | Listen host |
| `GOARD_PORT` | `"8300"` | Listen port |
| `GOARD_DB_PATH` | `/data/goard.db` | SQLite database path |
