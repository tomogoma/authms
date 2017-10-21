# Installation

####Note
One can use the installer
[here](https://github.com/tomogoma/repo-installer) to install all the
dependencies together with the micro-service by configuring using the
provided installer links.

## Pre-requisites

1. A [cockroachdb](https://www.cockroachlabs.com/) instance for
persistance. A systemd installer can be found here:
https://github.com/tomogoma/cockroach-installer
1. [consul](https://www.consul.io/) for service discovery. A systemd
installer can be found here:
https://github.com/tomogoma/consul-installer
1. [micro](https://github.com/micro/micro) as a gateway and load balancer.
A systemd installer can be found here:
https://github.com/tomogoma/micro-installer

## Install/Uninstall  (Systemd only)

The app can either be installed or [run directly](#running-the-micro-service).

1. install
    ```
    make install
    ```
1. uninstall
    ```
    make uninstall
    ```

# Configuration and deployment

## Configuring the service

Edit the configuration file. It is located at
`/etc/authms/authms<version>.conf.yml` after installation, and is documented for required values.

run the app with `--help` flag for options on custom configuration file and
other options.

## Running the Micro-Service

This section will be managed by SystemD if the respective installers were
used.

1. Make sure CockroachDB is running
    - Lack or misconfiguration of this will not stop the micro-service
     from starting, but requests will yield internal server errors until
     a connection to the db is established.
1. Start consul
1. **Recommended**: Start micro api with the proxy handler for access to the http API.
    ```
    micro api --handler=proxy
    ```
    More here: https://github.com/micro/examples/tree/master/greeter/api/rest
    
    You can set the systemd micro binary to use the proxy handler if
    installed using the [systemd installer](https://github.com/tomogoma/micro-installer)
    by following these steps:
    1. Open the unit file using your editor of choice
        ```
        nano /etc/systemd/system/microapi.service
        ```
        and append ` --handler=proxy` to the `ExecStart=...micro api` entry to
        end up with something like:
        ```
        ExecStart=/usr/local/bin/micro api --handler=proxy
        ```
    1. Reload the SystemD daemon
        ```
        systemctl daemon-reload
        ```
    1. Restart micro api
        ```
        systemctl restart microapi
        ```
1. Start the micro-service using one of the following methods:
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
        ...then executing the binary

## Accessing the Services

The API docs for service access can be accessed on your browser:
1. Via `micro api` with proxy handler
    ```
    http://localhost:8082/<version>/authms/docs
    ```
    
    **NOTE** The app has to be running for this option to work.
1. Static htm site in [install/docs](install/docs).
