# Contributing

## First time

- Fork the repo on GitHub.
- Clone your fork, add the upstream remote:

```bash
git clone https://github.com/<you>/goard.git
cd goard
git remote add upstream https://github.com/veloper/goard.git
```

## Making changes

Create a branch, make your changes, commit, keep it updated:

```bash
git checkout -b my-change
# make changes
git commit -m "area: brief description of what changed"
git fetch upstream
git rebase upstream/main
```

Try to keep one branch focused on one thing. Commit messages should be readable, not a specific format.

Before you push, make sure tests still pass:

```bash
go test ./... -count=1 -timeout 60s
```

## Pull request

Push your branch and open a PR on GitHub against `veloper/goard main`. Say what changed and why. If it fixes an open issue, mention it.

That's it. 

---

## Running locally

```bash
# Run the server
GOARD_ADMIN_USERNAME=admin GOARD_ADMIN_PAT=pat_admin go run ./cmd/goard

# Run the CLI against it
GOARD_HOST=http://localhost:8300 GOARD_PAT=pat_admin go run ./cmd/goardctl projects list
```

---

Happy hacking!