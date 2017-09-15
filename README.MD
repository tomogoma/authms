# Authentication Micro-Service

## Introduction

A Golang based authentication and Authorization micro-service that handles
various kinds of authentication:

1. username/password based authentication
1. phone/password based authentication
1. OAuth2 authentication
1. auth method verification e.g. phone number

## Installation

Note: One can use the installer
[here](https://github.com/tomogoma/repo-installer) to install all the
dependencies together with the micro-service by configuring using the
provided installer links.

### Pre-requisites

1. A [cockroachdb](https://www.cockroachlabs.com/) instance for
persistance. A systemd installer can be found here:
https://github.com/tomogoma/cockroach-installer
1. [consul](https://www.consul.io/) for service discovery. A systemd
installer can be found here:
https://github.com/tomogoma/consul-installer
1. [micro](https://github.com/micro/micro) as a gateway and load balancer.
A systemd installer can be found here:
https://github.com/tomogoma/micro-installer

## Installation and Uninstallation

1. clone the repository and `cd` into it
    ```
    git clone https://github.com/tomogoma/authms
    cd authms
    ```
    The app can either be installed or
    [run directly](#running-the-micro-service).
1. install (Systemd only)
    ```
    make install
    ```
1. uninstall (Systemd only)
    ```
    make uninstall
    ```
    
## Manual build

To run a manual build run the build command from inside the root folder:
```
make build
```

## Configuration and deployment

### Configuring the service

Edit the configuration file. It is located at
`/etc/authms/autms<version>.conf.yml` and is documented for required
values.

### Running the Micro-Service

This section will be managed by systemd if the respective installers were
used.

1. Make sure CockroachDB is running
    - Lack or misconfiguration of this will not stop the micro-service
     from starting, but requests will yield internal server errors until
     a connection to the db is established.
1. Start consul
1. Start micro (api, web or both)
    1. Start micro api if planning to use RPC communication between the
    gateway and the micro-service while maintaining http communication
    between the clients and the gateway.
        ```
        micro api
        ```
        More here:
        https://github.com/micro/examples/tree/master/greeter#sidecar
    1. Start micro web if planning to use http between all components.
        ```
        micro web
        ```
        More here:
        https://github.com/micro/examples/tree/master/greeter#web
1. Start the micro-service
    1. Using systemd after installation:
        ```
        sudo systemctl start authms<version>
        ```
    1. Running directly from the repo folder
       ```
       cd bin
       ./app
       ```
    1. Using the go toolchain e.g.
        ```
        go build
        ```
        or
        ```
        go install
        ```

### Accessing the Authentication Services

`micro` is the recommended gateway for accessing the micro-service.
    - `micro web` runs on port `8082` by default
    - `micro api` runs on port `8080` by default
        These options can be configured:
            ```
            micro --help
            ```
            or
            ```
            micro [option] --help
            ```

1. Access using `micro web` is direct:
    ```
    http://localhost:8082/authms<version>/...
    ```
1. Access using `micro api` has to be through POST methods e.g. (with curl):
   ```
   curl -H 'Content-Type: application/json' -d '{}' http://localhost:8080/authms<version>/AuthMS/...
   ```