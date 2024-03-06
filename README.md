## Conduktor ctl

How to build:
```
go build .
```

How to run:
```
read CDK_TOKEN
export CDK_TOKEN
export CDK_BASE_URL=http://localhost:8080/public/v1
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

