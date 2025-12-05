# Conduktor CLI Documentation

The Conduktor CLI is a command-line tool for interacting with Conduktor Console and Gateway services. 

It provides a kubectl-like experience for managing [Conduktor resources](https://docs.conduktor.io/guide/reference).

## Configuration
Before using the CLI, ensure you have set the necessary environment variables for connecting to Conduktor Console and/or Gateway. 

See [Environment Variable Configuration](./env-var-config.md) for details.

## Global Flags

All commands support these global flags:

- `-v, --verbose`: Verbose output (can be repeated: `-v` for debug, `-vv` for trace)
- `--permissive`: Permissive mode, allow undefined environment variables

## Commands Overview

### Core Resource Management Commands

#### `apply`
Upsert (create or update) resources on Conduktor from YAML/JSON files.

**Usage:**
```bash
conduktor apply -f <file>
conduktor apply -f <folder> --recursive
```

**Flags:**
- `-f, --file`: File or folder path (required, can be repeated)
- `-r, --recursive`: Apply all .yaml/.yml files in folder and subfolders
- `--dry-run`: Test changes without applying them
- `--print-diff`: Show differences between current and new resource
- `--parallelism`: Number of parallel operations (1-100, default: 1)
- `--enable-state`: Enable state management (see [State Management](./state_management.md))
- `--state-file`: Custom state file path (see [State Management](./state_management.md))

**Examples:**
```bash
# Apply single file
conduktor apply -f resource.yaml

# Apply multiple files
conduktor apply -f file1.yaml -f file2.yaml

# Apply all YAML files recursively
conduktor apply -f ./configs --recursive

# Dry run with diff
conduktor apply -f resource.yaml --dry-run --print-diff
```

#### `get`
Retrieve resources from Conduktor.

**Usage:**
```bash
conduktor get <resource-kind> [name]
conduktor get all
```

**Flags:**
- `-o, --output`: Output format (yaml|json|name, default: yaml)

**Examples:**
```bash
# List all resources
conduktor get all

# Get specific resource type
conduktor get topics

# Get specific resource by name
conduktor get User alice@mycompany.io

# Output as JSON
conduktor get Groups -o json

# Filter by backend (only useful for dual setup)
conduktor get all --gateway
conduktor get all --console
```

#### `delete`
Delete resources from Conduktor.

**Usage:**
```bash
conduktor delete -f <file>
conduktor delete <resource-kind> <name>
```

**Flags:**
- `-f, --file`: File or folder path
- `-r, --recursive`: Delete from all files in folder and subfolders
- `--dry-run`: Test deletion without executing
- `--enable-state`: Enable state management (see [State Management](./state_management.md))
- `--state-file`: Custom state file path (see [State Management](./state_management.md))

**Examples:**
```bash
# Delete from file
conduktor delete -f resource.yaml

# Delete specific resource
conduktor delete topic my-topic

# Dry run deletion
conduktor delete -f resource.yaml --dry-run
```

#### `edit`
Edit a resource in a text editor and apply changes.

**Usage:**
```bash
conduktor edit <resource-kind> <name>
```

**Examples:**
```bash
# Edit a topic
conduktor edit topic my-topic

# Edit a consumer group
conduktor edit consumergroup my-group
```

### Template and Development Commands

#### `template`
Generate YAML templates for resources.

**Usage:**
```bash
conduktor template <resource-kind>
```

**Flags:**
- `-o, --output`: Write template to file
- `-e, --edit`: Edit template after creation (requires --output)
- `-a, --apply`: Apply template after editing (requires --edit)

**Examples:**
```bash
# Print template to stdout
conduktor template topic

# Save template to file
conduktor template topic -o topic-template.yaml

# Create, edit, and apply template
conduktor template topic -o topic.yaml -e -a
```

### Utility Commands

#### `login`
Authenticate to Console backend using username/password provided as env var and obtain JWT token used for following queries.

**Usage:**
```bash
conduktor login
```

**Requirements:**
- CDK_USER and CDK_PASSWORD environment variables must be set

#### `version`
Display CLI version information.

**Usage:**
```bash
conduktor version
```

#### `sql`
Execute SQL queries on indexed topics (when available).

**Usage:**
```bash
conduktor sql "SELECT * FROM my_topic LIMIT 10"
```

**Flags:**
- `-n, --num-line`: Number of lines to display (default: 100)

#### `run`
Execute predefined actions/operations.

**Usage:**
```bash
conduktor run <action-name>
```

**Note:** Available actions depend on your Conduktor configuration and enabled features.

#### `token`
Manage admin and application instance tokens.

**Usage:**
```bash
conduktor token list admin
conduktor token list application-instance
```

## Resource Types

The CLI supports various resource types. See [Conduktor resource reference](https://docs.conduktor.io/guide/reference) documentation for more details.

## Working with Files

### File Format
Resources are defined in YAML or JSON format following this structure:

```yaml
apiVersion: v1
kind: Topic
metadata:
  name: my-topic
spec:
  # resource-specific configuration
```

### Batch Operations
You can work with multiple resources in several ways:

1. **Multiple files**: `conduktor apply -f file1.yaml -f file2.yaml`
2. **Folder**: `conduktor apply -f ./configs`
3. **Recursive folder**: `conduktor apply -f ./configs --recursive`
4. **Multiple resources in one file**: Separate resources with `---`

### Error Handling
- The CLI will report errors for individual resources while continuing to process others
- Use `--dry-run` to validate configurations before applying
- Check exit codes: 0 for success, non-zero for errors



