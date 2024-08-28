#!/bin/bash -eu

echoerr() { echo "$@" 1>&2; }

SCRIPTDIR=$(dirname "$0")

function cleanup {
	docker compose -f "$SCRIPTDIR/docker/docker-compose.yml" down
}

run() {
	docker compose -f docker/docker-compose.yml run --rm "$@"
}

trap cleanup EXIT
main() {
	cd "$SCRIPTDIR"
	docker compose -f docker/docker-compose.yml build
	docker compose -f docker/docker-compose.yml up -d mock mockGateway
	sleep 2
	run conduktor apply -f /test_resource.yml
	run conduktor apply -f /
	run conduktor delete -f /test_resource.yml
	run conduktor apply -f /
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

	# Gateway
	run conduktor apply -f /test_resource_gw.yml
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
}

main "$@"
