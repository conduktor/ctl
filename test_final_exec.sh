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
	docker compose -f docker/docker-compose.yml up -d mock
	sleep 1
	docker compose -f docker/docker-compose.yml run conduktor apply -f /test_resource.yml
	docker compose -f docker/docker-compose.yml run conduktor get Topic yolo --cluster=my-cluster
	docker compose -f docker/docker-compose.yml run conduktor delete Topic yolo -v --cluster=my-cluster
	docker compose -f docker/docker-compose.yml run -e CDK_USER=admin -e CDK_PASSWORD=secret conduktor login
	docker compose -f docker/docker-compose.yml run -e CDK_USER=admin -e CDK_PASSWORD=secret  -e CDK_API_KEY="" conduktor get KafkaCluster my_kafka_cluster
    docker compose -f docker/docker-compose.yml run conduktor token list admin
	docker compose -f docker/docker-compose.yml run conduktor token list application-instance -i=my_app_instance
	docker compose -f docker/docker-compose.yml run conduktor token create admin a_admin_token
	docker compose -f docker/docker-compose.yml run conduktor token create application-instance -i=my_app_instance a_admin_token
	docker compose -f docker/docker-compose.yml run conduktor token delete 0-0-0-0-0
}

main "$@"
