#!/bin/bash -eu

echoerr() { echo "$@" 1>&2; }

SCRIPTDIR=$(dirname "$0")

function cleanup {
	docker compose -f "$SCRIPTDIR/docker/docker-compose.yml" down
}

trap cleanup EXIT
main() {
	cd "$SCRIPTDIR"
	docker compose -f docker/docker-compose.yml build
	docker compose -f docker/docker-compose.yml up -d mock mockGateway
	sleep 1
	docker compose -f docker/docker-compose.yml run --rm conduktor apply -f /test_resource.yml
	docker compose -f docker/docker-compose.yml run --rm conduktor apply -f /
	docker compose -f docker/docker-compose.yml run --rm conduktor delete -f /test_resource.yml
	docker compose -f docker/docker-compose.yml run --rm conduktor apply -f /
	docker compose -f docker/docker-compose.yml run --rm conduktor get Topic yolo --cluster=my-cluster
	docker compose -f docker/docker-compose.yml run --rm conduktor delete Topic yolo -v --cluster=my-cluster
	docker compose -f docker/docker-compose.yml run --rm -e CDK_USER=admin -e CDK_PASSWORD=secret conduktor login
	docker compose -f docker/docker-compose.yml run --rm -e CDK_USER=admin -e CDK_PASSWORD=secret -e CDK_API_KEY="" conduktor get KafkaCluster my_kafka_cluster
	docker compose -f docker/docker-compose.yml run --rm conduktor token list admin
	docker compose -f docker/docker-compose.yml run --rm conduktor token list application-instance -i=my_app_instance
	docker compose -f docker/docker-compose.yml run --rm conduktor token create admin a_admin_token
	docker compose -f docker/docker-compose.yml run --rm conduktor token create application-instance -i=my_app_instance a_admin_token
	docker compose -f docker/docker-compose.yml run --rm conduktor token delete 0-0-0-0-0

	# Gateway
	docker compose -f docker/docker-compose.yml run --rm conduktor apply -f /test_resource_gw.yml
	docker compose -f docker/docker-compose.yml run --rm conduktor delete VClusters vcluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor get VClusters
	docker compose -f docker/docker-compose.yml run --rm conduktor get VClusters vcluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor get GatewayGroups
	docker compose -f docker/docker-compose.yml run --rm conduktor get GatewayGroups g1
	docker compose -f docker/docker-compose.yml run --rm conduktor get AliasTopics --show-defaults --name=yo --vcluster=mycluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor get ConcentrationRules --show-defaults --name=yo --vcluster=mycluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor get ServiceAccounts --show-defaults --name=yo --vcluster=mycluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor get interceptors --group=g1 --name=yo --username=me --vcluster=mycluster1
	docker compose -f docker/docker-compose.yml run --rm conduktor delete aliastopic aliastopicname --vcluster=v1
	docker compose -f docker/docker-compose.yml run --rm conduktor delete concentrationrule cr1 --vcluster=v1
	docker compose -f docker/docker-compose.yml run --rm conduktor delete serviceaccounts s1 --vcluster=v1
}

main "$@"
