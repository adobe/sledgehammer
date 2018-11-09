Sledgehammer [![Build Status](https://travis-ci.com/adobe/sledgehammer.svg?token=7fDSSWxNwGMMnLrqaxnB&branch=master)](https://travis-ci.com/adobe/sledgehammer)
======

### Introduction

`Sledgehammer` is a lightweight wrapper around docker to offer you native tools executed in docker containers.
It has been created to support dependency isolated executions and reduces the need to install tools locally.

### Usage

#### Installation

Start the installation of Sledgehammer with the following command:

    docker run adobe/slh

Sledgehammer will then guide you through the installation.

#### Using Sledgehammer

<aside class="notice">
Sledgehammer requires the docker API to be 1.35 or higher
</aside>

Once installed you can use Sledgehammer with the `slh` executable (given that the path you installed Sledgehammer to is in your PATH).

##### Configuration

### Build & Run

To build this project, you need Docker and Make:

    make build

This will generate an `adobe/slh` image locally which can then be installed as stated above.

As Sledgehammer is a go project a valid command would also be:

    cd installer && go build

and

    cd slh && go build

### Contributing

Contributions are welcomed! Read the [Contributing Guide](./.github/CONTRIBUTING.md) for more information.

### Licensing

This project is licensed under the Apache V2 License. See [LICENSE](LICENSE) for more information.