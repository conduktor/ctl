<a name="readme-top" id="readme-top"></a>

<p align="center">
  <img src="https://raw.githubusercontent.com/conduktor/conduktor.io-public/main/logo/transparent.png" width="256px" />
</p>
<h1 align="center">
    <strong>Conduktor CLI</strong>
</h1>

<p align="center">
    <a href="https://docs.conduktor.io/"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/conduktor/ctl/issues">Report Bug</a>
    ·
    <a href="https://github.com/conduktor/ctl/issues">Request Feature</a>
    ·
    <a href="https://support.conduktor.io/">Contact support</a>
    <br />
    <br />
    <a href=""><img alt="GitHub Release" src="https://img.shields.io/github/v/release/conduktor/ctl?sort=semver&logo=github&color=BCFE68"></a>
    ·
    <img alt="License" src="https://img.shields.io/github/license/conduktor/ctl?color=BCFE68">
    <br />
    <br />
    <a href="https://conduktor.io/"><img src="https://img.shields.io/badge/Website-conduktor.io-192A4E?color=BCFE68" alt="Scale Data Streaming With Security and Control"></a>
    ·
    <a href="https://twitter.com/getconduktor"><img alt="X (formerly Twitter) Follow" src="https://img.shields.io/twitter/follow/getconduktor?color=BCFE68"></a>
    ·
    <a href="https://conduktor.io/slack"><img src="https://img.shields.io/badge/Slack-Join%20Community-BCFE68?logo=slack" alt="Slack"></a>
</p>

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

To use Conduktor CLI, you need to:

**Define 2 environment variables for Console**:
 -   The URL of Conduktor Console
 -   Your API token (either a User Token or Application Token). You can generate an API token on `/settings/public-api-keys` page of your Console instance, or create one through the [CLI](https://docs.conduktor.io/platform/reference/cli-reference/#admin-api-key).
````yaml
CDK_BASE_URL=http://localhost:8080
CDK_API_KEY=<admin-token>
````

**Define 3 environment variables for Gateway**:
 -   The URL of the Conduktor Gateway API
 -   Your Gateway User for the API 
 -   Your Gatway Password for the API
````yaml
CDK_GATEWAY_BASE_URL=http://localhost:8888
CDK_GATEWAY_USER=admin
CDK_GATEWAY_PASSWORD=conduktor
````

### Commands Usage
````
You need to define the CDK_API_KEY and CDK_BASE_URL environment variables to use this tool.
You can also use the CDK_KEY,CDK_CERT to use a certificate for tls authentication.
If you have an untrusted certificate you can use the CDK_INSECURE=true variable to disable tls verification or you can use CACERT.

Usage:
  conduktor [flags]
  conduktor [command]

Available Commands:
  apply       Upsert a resource on Conduktor
  completion  Generate the autocompletion script for the specified shell
  delete      Delete resource of a given kind and name
  get         Get resource of a given kind
  help        Help about any command
  login       Login user using username password to get a JWT token
  token       Manage Admin and Application Instance Token
  version     Display the version of conduktor

Flags:
      --cert string   set pem cert for certificate authentication (useful for teleport)
  -h, --help          help for conduktor
      --key string    set pem key for certificate authentication (useful for teleport)
  -v, --verbose       show more information for debugging

Use "conduktor [command] --help" for more information about a command.
````

You can find more usage details on our:
 - [Console CLI documentation](https://docs.conduktor.io/platform/reference/cli-reference/)
 - [Gateway CLI documentation](https://docs.conduktor.io/gateway/reference/cli-reference/)


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
read CDK_API_KEY
export CDK_API_KEY
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

## Note about `v` in version

Before v0.2.6, versions used to follow the following format `X.Y.Z` like any other conduktor product.
When the conduktor team started to work on a terraform plugin, we wanted it to use gomod to reuse the client part of the CLI.
We realize that gomod requires version tags to start with a `v` (see: https://github.com/golang/go/issues/32945).
Therefore now, conduktor ctl version like any other go project starts with a v

