#!/bin/sh
# Mock editor for testing the edit command
# This script replaces the content of the file with a modified topic YAML

FILE="$1"

cat > "$FILE" << 'EOF'
# WARNING: Your file will be applied automatically once saved. If you do not want to apply anything, save an empty file.
---
apiVersion: v2
kind: Topic
metadata:
  name: edit-test-topic
  cluster: edit-cluster
spec:
  replicationFactor: 1
  partitions: 3
  configs:
    retention.ms: 604800000
EOF