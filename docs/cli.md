# CLI (tktrctl)

The `tktrctl` binary is included in the Docker image. It's also buildable from source:

```bash
go build -o tktrctl ./cmd/tktrctl
```

## Configuration

The CLI reads two environment variables — no flags needed:

```bash
export TICKETER_HOST=http://localhost:8300
export TICKETER_PAT=pat_admin
```

The PAT must belong to an admin user for user management commands.

## Commands

### Info

```bash
tktrctl info
```

Server metadata: valid states/types/priorities, registered users, projects, and the authenticated user.

### Users

```bash
# List all users
tktrctl users list

# Show one user
tktrctl users show 1

# Create a user (PAT is auto-generated, shown in response)
tktrctl users create bot-builder --display-name "Bot Builder"

# Create an admin user
tktrctl users create admin2 --admin

# Update a user's display name or PAT
tktrctl users update 2 --display-name "New Name"
tktrctl users update 2 --pat "pat_newtoken"

# Delete a user
tktrctl users delete 2
```

### Projects

```bash
# List all projects
tktrctl projects list

# Show one project
tktrctl projects show ASTEROID-GAME

# Create a project (name and slug are required positional args)
tktrctl projects create "Asteroid Game" ASTEROID-GAME

# With a description
tktrctl projects create "Asteroid Game" ASTEROID-GAME --description "A space game"

# Update a project
tktrctl projects update 1 --name "Asteroid Game v2" --description "Updated"
tktrctl projects update ASTEROID-GAME --slug NEW-SLUG

# Delete a project
tktrctl projects delete 1
```

### Issues

```bash
# List issues in a project
tktrctl issues list ASTEROID-GAME

# Filter by state or assignee
tktrctl issues list ASTEROID-GAME --state qa
tktrctl issues list ASTEROID-GAME --assignee 1

# Show one issue
tktrctl issues show ASTEROID-GAME-42
tktrctl issues show 42

# Create an issue (project and title are required positional args)
tktrctl issues create ASTEROID-GAME "Add rotation" --type feature --priority 2
tktrctl issues create ASTEROID-GAME "Fix login bug" --type bug --priority 1

# Update issue fields
tktrctl issues update 42 --title "Add ship rotation" --state qa

# Show current state
tktrctl issues state 42

# Update state (most common operation)
tktrctl issues state-update ASTEROID-GAME-42 qa
tktrctl issues state-update 42 done
```

## References

Projects and issues can be referenced by numeric ID or slug. Examples:

```
# These are equivalent:
tktrctl projects show 1
tktrctl projects show ASTEROID-GAME

tktrctl issues show 42
tktrctl issues show ASTEROID-GAME-42
```
