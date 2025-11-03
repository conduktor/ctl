#!/usr/bin/env bash
set -eu

echoerr() { echo "$@" 1>&2; }

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
REPO_DIR=$(dirname "$SCRIPT_DIR")
TESTDATA_DIR="$REPO_DIR/tests/integration/testdata"
echo "Using test data directory: $TESTDATA_DIR"

function cleanup {
	docker compose -f "$TESTDATA_DIR/docker-compose.yml" down
}

run() {
	docker compose -f "$TESTDATA_DIR/docker-compose.yml" run --rm "$@"
}

trap cleanup EXIT
main() {
	cd "$TESTDATA_DIR"
	docker compose -f docker-compose.yml build
	docker compose -f docker-compose.yml up -d mock mockGateway
	sleep 2
	run conduktor apply -f ./test_resource.yml
	run conduktor apply -f ./
	run conduktor delete -f ./test_resource.yml
	run conduktor apply -f ./
	run conduktor get Topic yolo --cluster=my-cluster
	run conduktor get Topic yolo --cluster=my-cluster -o yaml
	run conduktor get Topic yolo --cluster=my-cluster -o name
	run conduktor delete Topic yolo -v --cluster=my-cluster
	run -e CDK_USER=admin -e CDK_PASSWORD=secret conduktor login
	run -e CDK_USER=admin -e CDK_PASSWORD=secret -e CDK_API_KEY="" conduktor get KafkaCluster my_kafka_cluster
	run conduktor token list admin
	run conduktor token list application-instance -i=my_app_instance
	run conduktor token create admin a_admin_token
	run conduktor token create application-instance -i=my_app_instance a_admin_token
	run conduktor token delete 0-0-0-0-0
	run conduktor sql 'select * from "julien-cloud_sql_test"' -n 100

	# Edit command test
	run -e EDITOR="/mock_editor.sh" conduktor edit Topic edit-test-topic --cluster=edit-cluster

	# Gateway
	run conduktor apply -f ./test_resource_gw.yml
	run conduktor delete VirtualCluster vcluster1
	run conduktor get VirtualCluster
	run conduktor get VirtualCluster vcluster1
	run conduktor get GatewayGroup
	run conduktor get GatewayGroup g1
	run conduktor get AliasTopic --show-defaults --name=yo --vcluster=mycluster1
	run conduktor get ConcentrationRule --show-defaults --name=yo --vcluster=mycluster1
	run conduktor get GatewayServiceAccount --show-defaults --name=yo --vcluster=mycluster1
	run conduktor get interceptor --group=g1 --name=yo --username=me --vcluster=mycluster1
	run conduktor delete aliastopic aliastopicname --vcluster=v1
	run conduktor delete concentrationrule cr1 --vcluster=v1
	run conduktor delete gatewayserviceaccount s1 --vcluster=v1
	run conduktor run partnerZoneGenerateCredentials --partner-zone-name yo
	run conduktor run generateServiceAccountToken --life-time-seconds 3600 --username yo --v-cluster foyer
}

main "$@"
