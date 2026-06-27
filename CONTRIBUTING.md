# Contributing

## One-time setup

1. Fork the repo on GitHub.
2. Clone your fork:

```bash
git clone https://github.com/<your-username>/goard.git
cd goard
```

3. Add the upstream remote:

```bash
git remote add upstream https://github.com/veloper/goard.git
```

## Development

```bash
# Run tests
go test ./... -count=1 -timeout 60s

# Build both binaries
go build ./cmd/goard
go build ./cmd/goardctl

# Run the server locally
GOARD_ADMIN_USERNAME=admin GOARD_ADMIN_PAT=pat_admin go run ./cmd/goard

# Run the CLI against a running server
GOARD_HOST=http://localhost:8300 GOARD_PAT=pat_admin go run ./cmd/goardctl projects list
```

## Making changes

1. Create a branch from `main`:

```bash
git checkout -b my-change
```

2. Make your changes. Keep them focused — one branch, one concern.

3. Commit with a message that explains what and why:

```bash
git commit -m "package: short description"
```

The first word after the dash should be the package or area changed (`mcp`, `api`, `docs`, `README`, `cli`, `filter`, etc.).

4. Keep your branch updated:

```bash
git fetch upstream
git rebase upstream/main
```

5. Run tests again:

```bash
go test ./... -count=1 -timeout 60s
```

## Pull request

1. Push your branch:

```bash
git push origin my-change
```

2. Open a pull request on GitHub against `veloper/goard main`.

3. In the description, say what changed and why. If it fixes an issue, reference it.

4. That's it. No CLA, no DCO, no sign-off. If the change is small and tested, it'll get merged.
