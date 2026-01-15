# Remote State Backend Configuration Examples

This directory contains example configurations for using remote object storage backends to store Conduktor CLI state files.

## Overview

The Conduktor CLI supports storing state files in remote object storage instead of locally. This is useful for:

- **Team collaboration**: Multiple users can share the same state
- **CI/CD pipelines**: State persists across pipeline runs
- **Disaster recovery**: State is backed up in cloud storage
- **Multi-environment management**: Different state files for dev/staging/prod

## Supported Storage Backends

The CLI uses [thanos-io/objstore](https://github.com/thanos-io/objstore) which supports:

- **Amazon S3** and S3-compatible storage (MinIO, Ceph, etc.)
- **Google Cloud Storage (GCS)**
- **Azure Blob Storage**
- **OpenStack Swift**
- **Alibaba Cloud OSS**
- **Oracle Cloud Infrastructure Object Storage**
- **Baidu BOS**
- **Huawei OBS**

## Usage

To use a remote backend, create a configuration YAML file and pass it to the CLI:

```bash
conduktor apply -f resources.yaml \
  --enable-state \
  --state-backend-config examples/state-backend-s3.yaml
```

## Configuration Files

### S3 Configuration (`state-backend-s3.yaml`)

```yaml
type: S3
config:
  bucket: "my-conduktor-state-bucket"
  endpoint: "s3.amazonaws.com"
  region: "us-east-1"
  access_key: "${AWS_ACCESS_KEY_ID}"
  secret_key: "${AWS_SECRET_ACCESS_KEY}"
prefix: "conduktor/state/"
```

**Authentication options:**
- Access key + secret key (shown above)
- IAM role (omit `access_key` and `secret_key`)
- Session token for temporary credentials

**For S3-compatible storage (MinIO, Ceph):**
```yaml
config:
  bucket: "my-bucket"
  endpoint: "minio.example.com:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  insecure: true  # For HTTP instead of HTTPS
```

### GCS Configuration (`state-backend-gcs.yaml`)

```yaml
type: GCS
config:
  bucket: "my-conduktor-state-bucket"
  service_account: "/path/to/service-account.json"
prefix: "conduktor/state/"
```

**Authentication options:**
- Service account JSON file (shown above)
- Application Default Credentials (omit `service_account`)

### Azure Configuration (`state-backend-azure.yaml`)

```yaml
type: AZURE
config:
  storage_account_name: "myaccount"
  storage_account_key: "${AZURE_STORAGE_KEY}"
  container_name: "conduktor-state"
prefix: "conduktor/state/"
```

**Authentication options:**
- Storage account key (shown above)
- Managed identity: Add `msi_resource: "https://storage.azure.com/"`
- User-assigned identity: Add `user_assigned_id: "client-id"`

## Configuration Structure

All configurations follow this structure:

```yaml
type: <PROVIDER_TYPE>  # S3, GCS, AZURE, SWIFT, etc.
config:
  # Provider-specific configuration
  bucket: "bucket-name"
  # Authentication credentials
  # Connection settings
prefix: "optional/prefix/"  # Optional path prefix within the bucket
```

## State File Location

By default, the state file is named `cli-state.json` and will be stored at:
- Without prefix: `<bucket>/cli-state.json`
- With prefix: `<bucket>/<prefix>/cli-state.json`

## Environment Variables

Configuration files support environment variable expansion using `${VAR_NAME}` syntax:

```yaml
config:
  access_key: "${AWS_ACCESS_KEY_ID}"
  secret_key: "${AWS_SECRET_ACCESS_KEY}"
```

## Security Best Practices

1. **Never commit credentials**: Use environment variables or IAM roles
2. **Use encryption**: Enable server-side encryption on your bucket
3. **Restrict access**: Use least-privilege IAM policies
4. **Enable versioning**: Keep history of state file changes
5. **Use HTTPS**: Ensure `insecure: false` for production

## Example: Complete Workflow

1. Create a bucket configuration file:

```bash
cat > state-config.yaml <<EOF
type: S3
config:
  bucket: "my-state-bucket"
  region: "us-east-1"
prefix: "conduktor/prod/"
EOF
```

2. Apply resources with remote state:

```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"

conduktor apply -f resources/ \
  --enable-state \
  --state-backend-config state-config.yaml
```

3. On subsequent runs, the CLI will:
   - Load existing state from S3
   - Detect removed resources
   - Update state in S3 after apply

## Troubleshooting

### Connection Issues

- Verify network connectivity to storage service
- Check credentials are valid and not expired
- Ensure bucket exists and is accessible

### Permission Errors

Required permissions:
- **S3**: `s3:GetObject`, `s3:PutObject`, `s3:ListBucket`
- **GCS**: `storage.objects.get`, `storage.objects.create`, `storage.objects.update`
- **Azure**: Read, Write permissions on container

### State Conflicts

If multiple users/processes modify resources simultaneously:
- Consider implementing locking (not currently supported)
- Use separate state files per environment
- Coordinate apply operations in CI/CD

## For More Information

See the [thanos-io/objstore documentation](https://github.com/thanos-io/objstore) for detailed configuration options for each provider.
