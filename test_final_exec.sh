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
}

main "$@"
