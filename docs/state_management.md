# State Management

Conduktor CLI provides optional state management capabilities to track resources managed by the CLI. This feature helps maintain consistency and enables advanced workflows like detecting drift and cleaning up orphaned resources.

## Quick Start

```bash
# Local default location state (simplest)
conduktor apply -f resources.yaml --enable-state # store in default OS user application directory

# Local custom location state
conduktor apply -f resources.yaml --enable-state --state-file ./my-state.json

# Remote state with S3 (team collaboration)
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://my-bucket/conduktor/state/?region=us-east-1"

# Remote state with environment variable
export CDK_STATE_REMOTE_URI="s3://my-bucket/conduktor/state/?region=us-east-1"
conduktor apply -f resources.yaml --enable-state
```

## Overview

State management in Conduktor CLI:

- Tracks resources that have been applied using the CLI.
- Stores state information locally or in remote object storage.
- Enables detection of resources that were created via CLI but no longer defined in your files that need to be deleted.
- Supports multiple storage backends: Local file, Amazon S3, Google Cloud Storage, Azure Blob Storage.

## Enabling State Management

State management is **disabled by default** and must be explicitly enabled using command-line flags or environment variables.

### Flags

- `--enable-state`: Enable state management for the operation
- `--state-file`: Specify a custom path for the local state file (optional)
- `--state-remote-uri`: Specify a remote storage URI for the state file (optional)

### Environment Variables

- `CDK_STATE_ENABLED`: Enable state management globally (`true`, `1`, or `yes`)
- `CDK_STATE_FILE`: Specify a custom local state file path globally
- `CDK_STATE_REMOTE_URI`: Specify a remote storage URI globally

### Storage Backend Selection

The CLI automatically selects the appropriate storage backend:

- If `--state-remote-uri` is provided, uses remote object storage
- Otherwise, uses local file storage

### Default Local State File Location

When `--enable-state` is used without `--state-file` or `--state-remote-uri`, the default local location depends on the current system:

- **Linux**: `$XDG_DATA_HOME/.local/share/conduktor/cli-state.json`
- **Darwin (MacOS)**: `$HOME/Library/Application\ Support/conduktor/cli-state.json`
- **Windows**: `$APPDATA/conduktor/cli-state.json` or `$USERPROFILE/AppData/Roaming/conduktor/cli-state.json`

## Supported Commands

State management is currently supported by:

- `apply`: Records applied resources in state and auto-deletes removed resources
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

## Remote State Backends

Conduktor CLI supports storing state in remote object storage using URI-based configuration. This is powered by [gocloud.dev/blob](https://gocloud.dev/howto/blob/), which provides a unified interface across cloud providers.

### Supported Storage Providers

- **Amazon S3** and S3-compatible storage (MinIO, DigitalOcean Spaces, Ceph, etc.)
- **Google Cloud Storage (GCS)**
- **Azure Blob Storage**

### Remote URI Format

The `--state-remote-uri` flag accepts provider-specific URIs:

#### Amazon S3

```
s3://bucket-name/path/prefix/?option=value
```

**URI Components:**

- Bucket: S3 bucket name
- Path: Optional prefix path (folder structure)
- Query parameters (optional):
  - `region`: AWS region (e.g., `us-east-1`)
  - `endpoint`: Custom endpoint for S3-compatible storage
  - `profile`: The shared config profile to use; sets SharedConfigProfile.
  - `anonymous`: Forces use of anonymous credentials (`true`/`false`)
  - `disable_https`: Use HTTP instead of HTTPS (`true`/`false`)
  - `s3ForcePathStyle`: Use path-style addressing (`true`/`false`)
  - `use_path_style` : A value of true sets the UsePathStyle option.
  - `ssetype`: The type of server side encryption used (`AES256`, `aws:kms`, `aws:kms:dsse`)
  - `kmskeyid`: The KMS key ID for server side encryption
  - `accelerate`: Uses the S3 Transfer Accleration endpoints (`true`/`false`)
  - `hostname_immutable`: Make the hostname immutable, only works if endpoint is also set.
  - `dualstack`: Enables dual stack (IPv4 and IPv6) endpoints (`true`/`false`)
  - `fips`: Enables the use of FIPS endpoints (`true`/`false`)
  - `rate_limiter_capacity`: A integer value configures the capacity of a token bucket used in client-side rate limits. If no value is set, the client-side rate limiting is disabled. See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/retries-timeouts/#client-side-rate-limiting.
  - `request_checksum_calculation`: Request checksum calculation mode (`when_supported`, `when_required`)
  - `response_checksum_validation`: Response checksum validation mode (`when_supported`, `when_required`)

**Examples:**

```bash
# Standard AWS S3
s3://my-bucket/conduktor/state/

# With region
s3://my-bucket/conduktor/state/?region=eu-west-1

# MinIO (S3-compatible)
s3://my-bucket/state/?region=us-east-1&endpoint=http://minio.example.com:9000&disableSSL=true&s3ForcePathStyle=true

# DigitalOcean Spaces
s3://my-space/state/?region=nyc3&endpoint=https://nyc3.digitaloceanspaces.com
```

#### Google Cloud Storage (GCS)

```
gs://bucket-name/path/prefix/?option=value
```

**URI Components:**

- Bucket: GCS bucket name
- Path: Optional prefix path
- Query parameters (optional):
  - `access_id`: Sets Options.GoogleAccessID; only used in SignedURL, except that a value of "-" forces the use of an unauthenticated client.
  - `private_key_path`: Path to read for Options.PrivateKey; only used in SignedURL.
  - `anonymous`: Forces the use of an unauthenticated client (`true`/`false`)

**Examples:**

```bash
# Standard GCS
gs://my-bucket/conduktor/state/

# With project ID
gs://my-bucket/conduktor/state/?projectid=my-gcp-project
```

#### Azure Blob Storage (Azure Storage Account)

```
azblob://container-name/path/prefix/?option=value
```

**URI Components:**

- Container: Azure Blob container name
- Path: Optional prefix path
- Query parameters (optional):
  - `storage_account`: Azure storage account name
  - `domain`: Overrides Options.StorageDomain.
  - `cdn`: Overrides Options.IsCDN.
  - `protocol`: Overrides Options.Protocol.
  - `localemu`: Overrides Options.IsLocalEmulator.

**Examples:**

```bash
# Standard Azure Blob
azblob://my-container/conduktor/state/

# With storage account
azblob://my-container/state/?storage_account=myaccount
```

### Custom State Filename

By default, the state file is named `cli-state.json`. You can specify a custom filename by ending the URI with `.json`:

```bash
# Default filename (cli-state.json)
s3://my-bucket/conduktor/state/

# Custom filename (my-custom-state.json)
s3://my-bucket/conduktor/my-custom-state.json

# Custom filename with path (prod-state.json)
s3://my-bucket/environments/prod-state.json
```

### Authentication

Authentication is handled through standard provider-specific mechanisms:

#### AWS/S3

1. `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables
2. `AWS_SESSION_TOKEN` for temporary credentials
3. IAM role (for EC2, ECS, Lambda)
4. `~/.aws/credentials` file
5. `AWS_PROFILE` environment variable

#### Google Cloud Storage

1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (path to service account JSON)
2. Application Default Credentials (`gcloud auth application-default login`)
3. GCE metadata service (automatic on GCE/GKE)

#### Azure Blob Storage

1. `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` or `AZURE_STORAGE_SAS_TOKEN` environment variables
2. Managed identity (automatic on Azure VMs/AKS)

## Usage Examples

### Basic State Management

```bash
# Apply resources with local state tracking
conduktor apply -f resources.yaml --enable-state

# Delete resources with local state tracking
conduktor delete -f resources.yaml --enable-state
```

### Local Custom State File

```bash
# Use a project-specific state file
conduktor apply -f resources.yaml --enable-state --state-file ./project-state.json

# Multiple environments with separate state files
conduktor apply -f prod-resources.yaml --enable-state --state-file ./prod-state.json
conduktor apply -f dev-resources.yaml --enable-state --state-file ./dev-state.json
```

### Remote State with S3

```bash
# Using CLI flag
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://my-state-bucket/conduktor/prod/?region=us-east-1"

# Using environment variable
export CDK_STATE_REMOTE_URI="s3://my-state-bucket/conduktor/prod/?region=us-east-1"
conduktor apply -f resources.yaml --enable-state
```

### Remote State with GCS

```bash
# Set credentials
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Apply with remote state
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "gs://my-state-bucket/conduktor/prod/"
```

### Remote State with Azure

```bash
# Set credentials
export AZURE_STORAGE_ACCOUNT="myaccount"
export AZURE_STORAGE_KEY="base64-encoded-key=="

# Apply with remote state
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "azblob://conduktor-state/prod/"
```

### MinIO (S3-Compatible) for Testing

```bash
# Start MinIO locally (Docker)
docker run -d -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Create bucket
docker run --rm -it --network host minio/mc \
  alias set myminio http://localhost:9000 minioadmin minioadmin && \
  mc mb myminio/conduktor-state

# Use with CLI
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://conduktor-state/test/?region=us-east-1&endpoint=http://localhost:9000&disable_https=true&s3ForcePathStyle=true"
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

### Choosing Storage Backend

**Use Local State When:**

- Working on personal projects
- Single-user scenarios
- No need for state sharing
- Use with provisioner chart along with PVC

**Use Remote State When:**

- Team collaboration required
- CI/CD pipelines
- Multiple environments (dev/staging/prod)
- State backup and versioning needed
- Working from multiple machines

### Environment Separation

**Local State:**
```bash
# Different local state files for different environments
conduktor apply -f resources.yaml --enable-state --state-file ./dev-state.json
conduktor apply -f resources.yaml --enable-state --state-file ./staging-state.json
conduktor apply -f resources.yaml --enable-state --state-file ./prod-state.json
```

**Remote State (Recommended):**
```bash
# Different remote paths/files for different environments
conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://state-bucket/my-app/dev/"

conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://state-bucket/my-app/staging/"

conduktor apply -f resources.yaml \
  --enable-state \
  --state-remote-uri "s3://state-bucket/my-app/prod/"
```

### CI/CD Integration

**GitHub Actions Example:**
```yaml
name: Deploy to Production
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Apply Conduktor Resources
        env:
          CDK_BASE_URL: ${{ secrets.CONDUKTOR_URL }}
          CDK_API_KEY: ${{ secrets.CONDUKTOR_API_KEY }}
          CDK_STATE_REMOTE_URI: s3://my-state-bucket/prod/?region=us-east-1
        run: |
          conduktor apply -f resources/ --recursive --enable-state
```

### Version Control

**Local State:**
```bash
# Include state files in version control for team collaboration
git add *.yaml
git add *-state.json
git commit -m "Update resources and state"
```

**Remote State:**
```bash
# No need to commit state files - they're in remote storage
# Only commit resource definitions
git add resources/
git commit -m "Update resources"
```

### Security Best Practices

**Remote State:**
```bash
# 1. Use environment variables for credentials (never hardcode)
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# 2. Use IAM roles when possible (no credentials needed)
# When running on AWS EC2/ECS/Lambda, IAM roles are automatic

# 3. Enable encryption at rest
aws s3api put-bucket-encryption \
  --bucket my-state-bucket \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# 4. Restrict bucket access with IAM policies
# Ensure only authorized users/services can access state bucket

# 5. Enable bucket versioning for audit trail
aws s3api put-bucket-versioning \
  --bucket my-state-bucket \
  --versioning-configuration Status=Enabled
```

## Limitations and Considerations

### Current Limitations

- State doesn't include resource specifications, only metadata
- No built-in state locking mechanism

### File Management

- **Local State**: Should be backed up regularly or included in version control
- **Remote State**: Enable versioning on your storage bucket for backup/rollback

### Concurrency

**Local State:**

- Multiple CLI operations on the same state file should be avoided
- State files are not locked during operations
- Last-write-wins behavior for concurrent modifications

**Remote State:**

- Object storage provides atomic operations for individual file writes
- However, no distributed locking is implemented
- Coordinate CLI operations in CI/CD pipelines to avoid conflicts
- Consider using separate state files/paths per environment/resources set to reduce conflicts


## Troubleshooting

### Local State File Corruption

If the local state file becomes corrupted:
```bash
# Backup corrupted file
mv ~/.conduktor/ctl/cli-state.json ~/.conduktor/ctl/cli-state.json.corrupted

# Start fresh (you'll lose state tracking for existing resources)
conduktor apply -f resources.yaml --enable-state
```

### Remote State Connection Issues

**Error: failed to open bucket**

```bash
# Verify credentials are set
echo $AWS_ACCESS_KEY_ID
echo $AWS_SECRET_ACCESS_KEY

# Test bucket access directly
aws s3 ls s3://my-bucket/

# Check URI format
# Correct: s3://bucket-name/path/
# Incorrect: s3://bucket-name.s3.amazonaws.com/path/
```

**Error: failed to read/write state file from remote storage**

```bash
# Check if state file exists
aws s3 ls s3://my-bucket/conduktor/state/cli-state.json

# Verify permissions
# Required: s3:GetObject, s3:PutObject, s3:ListBucket

# Check network connectivity
curl -I https://s3.amazonaws.com
```

**MinIO/S3-Compatible Storage Issues**

```bash
# Ensure all required query parameters are set
# Correct URI format for MinIO:
s3://bucket/path/?region=us-east-1&endpoint=http://minio:9000&disable_https=true&s3ForcePathStyle=true

# Common issues:
# - Missing s3ForcePathStyle=true for MinIO
# - Wrong endpoint (should not include bucket name)
# - Certificate issues (use disable_https=true for self-signed certs)
```

### Migration Between State Backends

**From Local to Remote:**
```bash
# 1. Ensure current local state is up to date
conduktor apply -f resources.yaml --enable-state --state-file ./current-state.json

# 2. Copy local state to remote (manual, using AWS CLI)
aws s3 cp ./current-state.json s3://my-bucket/conduktor/state/cli-state.json

# 3. Switch to using remote state
export CDK_STATE_REMOTE_URI="s3://my-bucket/conduktor/state/"
conduktor apply -f resources.yaml --enable-state
```

**From Remote to Local:**
```bash
# Download remote state
aws s3 cp s3://my-bucket/conduktor/state/cli-state.json ./local-state.json

# Use local state
conduktor apply -f resources.yaml --enable-state --state-file ./local-state.json
```

### Authentication Failures

**AWS/S3:**
```bash
# Check AWS credentials
aws sts get-caller-identity

# Use specific profile
export AWS_PROFILE=my-profile
conduktor apply -f resources.yaml --enable-state --state-remote-uri "s3://bucket/state/"

# For temporary credentials (MFA)
export AWS_SESSION_TOKEN=your-token
```

**GCS:**
```bash
# Check service account
echo $GOOGLE_APPLICATION_CREDENTIALS
cat $GOOGLE_APPLICATION_CREDENTIALS

# Verify credentials work
gcloud auth list
gsutil ls gs://my-bucket/
```

**Azure:**
```bash
# Check credentials
echo $AZURE_STORAGE_ACCOUNT
echo $AZURE_STORAGE_KEY

# Test access
az storage container list --account-name myaccount
```