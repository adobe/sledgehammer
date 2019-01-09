Sledgehammer [![Build Status](https://travis-ci.com/adobe/sledgehammer.svg?token=7fDSSWxNwGMMnLrqaxnB&branch=master)](https://travis-ci.com/adobe/sledgehammer)
======

### Introduction

`Sledgehammer` is a lightweight wrapper around docker to offer you native tools executed in docker containers.
It has been created to support dependency isolated executions and reduces the need to install tools locally.

### Installation

You can start the installation of Sledgehammer with the following command:

    docker run adobe/slh

Sledgehammer will then guide you through the installation.

You need two steps in order to fully install Sledgehammer on your system.

* Mount the docker socket
* Pass in a local executable path

#### Step 1:

In the simplest case and on most systems you can mount the docker socket with
    
    docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock adobe/slh

<aside class="notice">
You can also pass `DOCKER_HOST` if needed
</aside>

#### Step 2:

You need to mount a local directory (which should be in your path) to the `/bin` direcotry of the container.
    
    docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v <localPath>:/bin adobe/slh

That should install Sledgehammer sucessfully on your system.

Sledgehammer will also tell you what to to during the installation.

### Usage

<aside class="notice">
Sledgehammer requires the docker API to be 1.35 or higher
</aside>

Once installed you can use Sledgehammer with the `slh` executable (given that the path you installed Sledgehammer to is in your PATH).

Sledgehammer - Dependency isolated executions.
    Use dockerized tools like they are installed on the  system.

    Usage:
    slh [command]

    Available Commands:
    create      Create a ressources
    delete      Delete a resource
    describe    Describe detailed information about ressources
    get         Get ressources
    help        Help about any command
    install     Install a tool on the system
    reset       Reset an alias
    update      Update all registries

    Flags:
        --confdir string     Location of configuration directory. Default is bindir/.slh
    -h, --help               help for slh
        --log-level string   Set the log level (debug|info|warning|error|fatal|panic) (default "none")
    -o, --output string      Define the output, currently supported is text|json (default "text")
        --version            version for slh

    Use "slh [command] --help" for more information about a command.

#### Quick start

Sledgehammer has the concept of registries where in each registry there can be any amount of tools and tool kits.

Sledgehammer ships with a registry by default (called `default`), so to see what tools are provided you can use

    slh get tools

This should list all available tools, the same can be done with registries, tool kits as well as mounts.

    slh get registries
    slh get kits
    slh get mounts

To install a tool you can use the install command

    slh install <toolname> <flags>

Sledgehammer will then create a symlink with the tool name, so that you can call the tool with 
    
    <toolname> <arguments>

#### Versioning

Sledgehammer supports versioned tools. That means you can have multiple versions of a single tool installed on the system.

By default Sledgehammer will take the latest version available following semantic versioning.

So with the following versions

    latest
    1.0.0
    1.2.0

Sledgehammer will use version `1.2.0` by default.
Details can be found in the [versioning documentation](./doc/VERSIONING.md).

### Build & Run

To build this project, you need Docker and Make:

    make ci

This will generate an `adobe/slh` image locally which can then be installed as stated above.

As Sledgehammer is a go project you can also place the repository in a valid go path (`<GOPATH>/src/github.com/adobe/sledgehammer`) and execute:

    go test ./...

To build the installer:

    cd installer && go build

or to build the slh executable:

    cd slh && go build

### Contributing

Contributions are welcomed! Read the [Contributing Guide](./doc/CONTRIBUTING.md) for more information.

### Licensing

This project is licensed under the Apache V2 License. See [LICENSE](LICENSE) for more information.