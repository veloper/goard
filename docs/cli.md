# CLI (goardctl)

The `goardctl` binary is included in the Docker image. It's also buildable from source:

```bash
go build -o goardctl ./cmd/goardctl
```

## Configuration

The CLI reads two environment variables — no flags needed:

```bash
export GOARD_HOST=http://localhost:8300
export GOARD_PAT=pat_admin
```

The PAT must belong to an admin user for user management commands.

## Commands

### Info

```bash
goardctl info
```

Server metadata: valid states/types/priorities, registered users, projects, and the authenticated user.

### Users

```bash
# List all users
goardctl users list

# Show one user
goardctl users show 1

# Create a user (PAT is auto-generated, shown in response)
goardctl users create bot-builder

# Create an admin user
goardctl users create admin2 --admin

# Update a user's PAT
goardctl users update 2 --pat "pat_newtoken"

# Delete a user
goardctl users delete 2
```

### Projects

```bash
# List all projects
goardctl projects list

# Show one project
goardctl projects show ASTEROID-GAME

# Create a project (name and slug are required positional args)
goardctl projects create "Asteroid Game" ASTEROID-GAME

# With a description
goardctl projects create "Asteroid Game" ASTEROID-GAME --description "A space game"

# Update a project
goardctl projects update 1 --name "Asteroid Game v2" --description "Updated"
goardctl projects update ASTEROID-GAME --slug NEW-SLUG

# Delete a project
goardctl projects delete 1
```

### Issues

```bash
# List issues in a project
goardctl issues list ASTEROID-GAME

# Filter by state or assignee
goardctl issues list ASTEROID-GAME --state review
goardctl issues list ASTEROID-GAME --assignee 1

# Show one issue
goardctl issues show ASTEROID-GAME-42
goardctl issues show 42

# Create an issue (project and title are required positional args)
goardctl issues create ASTEROID-GAME "Add rotation" --type feature --priority 2
goardctl issues create ASTEROID-GAME "Fix login bug" --type bug --priority 1

# Update issue fields
goardctl issues update 42 --title "Add ship rotation" --state review

# Show current state
goardctl issues state 42

# Update state (most common operation)
goardctl issues state-update ASTEROID-GAME-42 review
goardctl issues state-update 42 done
```

## References

Projects and issues can be referenced by numeric ID or slug. Examples:

```
# These are equivalent:
goardctl projects show 1
goardctl projects show ASTEROID-GAME

goardctl issues show 42
goardctl issues show ASTEROID-GAME-42
```
