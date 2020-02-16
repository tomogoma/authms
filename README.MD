# Authentication Micro-Service

## Introduction

A Golang based authentication and Authorization micro-service that handles
various kinds of authentication:

1. username/password based authentication
1. phone/password based authentication
1. OAuth2 authentication
1. auth method verification e.g. phone number

In this Readme you will come accross the variable `<version>`,
it is the major version code of the `VersionFull` variable in
[config/consts.go](config/consts.go) as per http://semver.org

e.g. 
```
	Name        = "seedms"
	VersionFull = "0.1.2"
```
yields

`<version>` => `0`


## Installation/Deployment

Refer to the respective readme files for instructions:

1. [Using micro or standalone service](MICRO.MD)
2. [Deploy to Google AppEngine](cmd/gcloud/README.MD)

## Manual build

### Pre-requisites

1. [Go](https://golang.org) 1.8+ is required for this.
1. APIDoc installed and added to `PATH` variable: http://apidocjs.com

### Build

Run the build command from inside the root folder:
```
make build
```

`make build` can also be supplied with `goos=[GOOS]` `goarch=[GOARCH]`
`goarm=[GOARM]` values as per the documentation at https://golang.org/cmd/go (untested)

build uses the go toolchain to build binaries. It also generates API docs and
configuration templates.
