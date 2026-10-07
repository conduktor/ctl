#!/bin/sh
# Lays out the single node and creates the bucket and key the remote state tests use, through Garage's admin API.
set -eu
apk add --no-cache curl jq >/dev/null
admin="http://garage:3903/v2"
auth="Authorization: Bearer integration-test-admin-token"
call() { curl -fsS -H "$auth" -H "Content-Type: application/json" "$@"; }

until call "$admin/GetClusterStatus" >/dev/null 2>&1; do sleep 1; done
node=$(call "$admin/GetClusterStatus" | jq -r '.nodes[0].id')
call -X POST "$admin/UpdateClusterLayout" \
  -d "{\"roles\":[{\"id\":\"$node\",\"zone\":\"dc1\",\"capacity\":1073741824,\"tags\":[]}]}" >/dev/null
version=$(call "$admin/GetClusterLayout" | jq -r '.version + 1')
call -X POST "$admin/ApplyClusterLayout" -d "{\"version\":$version}" >/dev/null

call -X POST "$admin/ImportKey" \
  -d "{\"name\":\"ctl-tests\",\"accessKeyId\":\"$ACCESS_KEY\",\"secretAccessKey\":\"$SECRET_KEY\"}" >/dev/null
bucket=$(call -X POST "$admin/CreateBucket" -d '{"globalAlias":"conduktor-state"}' | jq -r '.id')
call -X POST "$admin/AllowBucketKey" \
  -d "{\"bucketId\":\"$bucket\",\"accessKeyId\":\"$ACCESS_KEY\",\"permissions\":{\"read\":true,\"write\":true,\"owner\":true}}" >/dev/null
echo "Garage bucket created successfully"
