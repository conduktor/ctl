## Conduktor ctl
How to download binary:
```
Look for assets of the last release:
https://github.com/conduktor/ctl/releases
```
How to get the latest docker image:
```
docker pull harbor.cdkt.dev/conduktor/conduktor-ctl
```

How to build:
```
go build .
```

How to run:
```
read CDK_TOKEN
export CDK_TOKEN
export CDK_BASE_URL=http://localhost:8080
go run . 
```

How to run unit test:
```
go test ./...
```

How to run integration test:
```
./test_final_exec.sh
```

## How to use behind teleport

First login to your teleport proxy, for example:
```
tsh login --proxy=teleport-01.prd.tooling.cdkt.dev --auth=github
```

```
conduktor get application --cert $(tsh apps config --format=cert) --key $(tsh apps config --format=key)
```

Or:
```
export CDK_CERT=$(tsh apps config --format=cert)
export CDK_KEY=$(tsh apps config --format=key)
conduktor get application
```
