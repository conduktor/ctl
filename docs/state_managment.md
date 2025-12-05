# State Management

Conduktor CLI provides optional state management capabilities to track resources managed by the CLI. This feature helps maintain consistency and enables advanced workflows like detecting drift and cleaning up orphaned resources.

## Overview

State management in Conduktor CLI:
- Tracks resources that have been applied using the CLI
- Stores state information in a local file (by default)
- Enables detection of resources that were created via CLI but no longer defined in your files that need to be deleted

## Enabling State Management

State management is **disabled by default** and must be explicitly enabled using command-line flags or environment variables.

### Flags

- `--enable-state`: Enable state management for the operation
- `--state-file`: Specify a custom path for the state file (optional)

### Environment Variables
- `CDK_ENABLE_STATE`: Enable state management globally
- `CDK_STATE_FILE`: Specify a custom state file path globally

### Default State File Location

When `--enable-state` is used without `--state-file`, the default location depend on current system :
- **Linux** : `$XDG_DATA_HOME/.local/share/conduktor/cli-state.json` or `$HOME/.config/conduktor/cli-state.json`
- **Darwin (MacOS)** : `$HOME/Library/Application\ Support/conduktor/cli-state.json`
- **Windows** : `$APPDATA/conduktor/cli-state.json` or `$USERPROFILE/AppData/Roaming/conduktor/cli-state.json`

## Supported Commands

State management is currently supported by:
- `apply`: Records applied resources in state
- `delete`: Removes deleted resources from state

## State File Format

The state file is stored in JSON format with the following structure:

```json
{
  "version": "v1",
  "lastUpdated": "2023-12-05T10:30:00Z",
  "resources": [
    {
      "apiVersion": "v1",
      "kind": "Topic",
      "metadata": {
        "name": "my-topic"
      }
    },
    {
      "apiVersion": "v1",
      "kind": "User",
      "metadata": {
        "name": "alice@my-company.io"
      }
    }
  ]
}
```

### State File Fields

- `version`: State file format version (currently `v1`)
- `lastUpdated`: ISO 8601 timestamp of last state modification
- `resources`: Array of resource states, each containing:
  - `apiVersion`: Resource API version
  - `kind`: Resource type
  - `metadata`: Resource metadata (name and other identifying information)

## Usage Examples

### Basic State Management

```bash
# Apply resources with state tracking
conduktor apply -f resources.yaml --enable-state

# Delete resources with state tracking
conduktor delete -f resources.yaml --enable-state
```

### Custom State File

```bash
# Use a project-specific state file
conduktor apply -f resources.yaml --enable-state --state-file ./project-state.yaml

# Multiple environments with separate state files
conduktor apply -f prod-resources.yaml --enable-state --state-file ./prod-state.yaml
conduktor apply -f dev-resources.yaml --enable-state --state-file ./dev-state.yaml
```

### Combined with Other Features

```bash
# Dry run with state (state won't be modified)
conduktor apply -f resources.yaml --enable-state --dry-run

# Parallel apply with state tracking
conduktor apply -f resources/ --recursive --enable-state --parallelism 5
```

## State Lifecycle

### Apply Operations
1. CLI loads existing state file (creates new if doesn't exist)
2. If resource missing in files but present in state, it is considered orphaned and will be deleted
3. Resources are applied to Conduktor
4. Successfully applied resources are added to state
5. State file is updated with new timestamp

### Delete Operations
1. CLI loads existing state file
2. Resources are deleted from Conduktor
3. Successfully deleted resources are removed from state
4. State file is updated

### Error Handling
- If state cannot be loaded, the operation fails immediately
- State is saved after resource operations, even if some resources fail
- Both operation errors and state save errors are reported

## Best Practices

### Project Organization
```bash
# Per-project state files
my-project/
├── resources/
│   ├── topics.yaml
│   └── consumers.yaml
├── project-state.yaml
└── scripts/
    └── deploy.sh
```

```bash
#!/bin/bash
# deploy.sh
conduktor apply -f resources/ --recursive --enable-state --state-file ./project-state.yaml
```

### Environment Separation
```bash
# Different state files for different environments
conduktor apply -f resources.yaml --enable-state --state-file ./dev-state.yaml
conduktor apply -f resources.yaml --enable-state --state-file ./staging-state.yaml
conduktor apply -f resources.yaml --enable-state --state-file ./prod-state.yaml
```

### Version Control
```bash
# Include state files in version control for team collaboration
git add *.yaml
git add *-state.yaml
git commit -m "Update resources and state"
```

### Backup and Recovery
```bash
# Backup state before major changes
cp project-state.yaml project-state.yaml.backup

# Restore from backup if needed
cp project-state.yaml.backup project-state.yaml
```

## Limitations and Considerations

### Current Limitations
- State is only stored locally (no remote/shared storage)
- No automatic drift detection (planned for future releases)
- No garbage collection of orphaned resources (planned for future releases)
- State doesn't include resource specifications, only metadata

### File Management
- State files should be backed up regularly
- Consider including state files in version control for team environments
- State files can grow large with many resources

### Concurrency
- Multiple CLI operations on the same state file should be avoided
- State files are not locked during operations
- Last-write-wins behavior for concurrent modifications

## Troubleshooting

### State File Corruption
If the state file becomes corrupted:
```bash
# Backup corrupted file
mv ~/.conduktor/ctl/state.yaml ~/.conduktor/ctl/state.yaml.corrupted

# Start fresh (you'll lose state tracking for existing resources)
conduktor apply -f resources.yaml --enable-state
```

### Migration Between State Files
Currently, there's no built-in migration tool. To move to a new state file:
```bash
# Apply all current resources with new state file
conduktor apply -f all-resources.yaml --enable-state --state-file ./new-state.yaml
```

### State File Location Issues
```bash
# Check default location
ls -la ~/.conduktor/ctl/

# Create directory if missing
mkdir -p ~/.conduktor/ctl/

# Use absolute paths for custom state files
conduktor apply -f resources.yaml --enable-state --state-file /absolute/path/to/state.yaml
```

## Future Enhancements

Planned features for state management include:
- Drift detection (comparing state vs actual resources)
- Garbage collection (removing resources not in files but in state)
- Remote state backends (cloud storage, databases)
- State locking for concurrent access
- Resource specification tracking for better drift detection
