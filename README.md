# Conduktor ctl
![Release](https://img.shields.io/github/v/release/conduktor/ctl?sort=semver&logo=github)
![License](https://img.shields.io/github/license/conduktor/ctl)
[![Roadmap](https://img.shields.io/badge/Roadmap-click%20here-blueviolet)](https://product.conduktor.help/tabs/1-in-development)
[![twitter](https://img.shields.io/twitter/follow/getconduktor.svg?style=social)](https://twitter.com/getconduktor)

This repository contain Conduktor CLI source code. 
Conduktor CLI is a command line tool to interact with Conduktor Console. 
It is strongly inspired by Kubernetes kubectl CLI tool and reuse some of it's concepts.


## How to install

### From binaries (Linux, MacOS, Windows)

Look for assets of the last release at https://github.com/conduktor/ctl/releases 

### Using Docker image
How to get the latest docker image:
```
docker pull conduktor/conduktor-ctl:latest
```

### From source 
You will need Go 1.22+ installed and configured on your machine.

To build simply run
```
go build -o conduktor .
```
You will find the `conduktor` binary at the root of the project.

## Usage

To run the CLI you will need to provide the Conduktor Console URL and an API access token.

### Configure

To use Conduktor CLI, you need to define 2 environment variables:
-   The URL of Conduktor Console
-   Your API token (either a User Token or Application Token). You can generate an API token on `/settings/public-api-keys` page of your Console instance.
````yaml
CDK_BASE_URL=http://localhost:8080
CDK_TOKEN=<admin-token>
````

### Commands Usage
````
You need to define the CDK_TOKEN and CDK_BASE_URL environment variables to use this tool.
You can also use the CDK_KEY,CDK_CERT instead of --key and --cert flags to use a certificate for tls authentication.
If you have an untrusted certificate you can use the CDK_INSECURE=true variable to disable tls verification

Usage:
  conduktor [flags]
  conduktor [command]

Available Commands:
  apply       Upsert a resource on Conduktor
  completion  Generate the autocompletion script for the specified shell
  delete      Delete resource of a given kind and name
  get         Get resource of a given kind
  help        Help about any command
  version     Display the version of conduktor

Flags:
      --cert string   set pem cert for certificate authentication (useful for teleport)
  -h, --help          help for conduktor
      --key string    set pem key for certificate authentication (useful for teleport)
  -v, --verbose       show more information for debugging

Use "conduktor [command] --help" for more information about a command.
````

You can find more usage details on our [documentation](https://docs.conduktor.io/platform/reference/cli-reference/)


#### How to use behind teleport
If you are using Conduktor behind a teleport proxy, you will need to provide the certificate and key to the CLI using `CDK_CERT` and `CDK_KEY` environment variables.

First login to your teleport proxy, for example:
```
tsh login --proxy="$TELEPORT_SERVER" --auth="$TELEPORT_AUTH_METHOD"
tsh apps login console
export CDK_CERT=$(tsh apps config --format=cert)
export CDK_KEY=$(tsh apps config --format=key)
conduktor get application
```

### Development

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

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
