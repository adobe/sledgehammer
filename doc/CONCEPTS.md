# CONCEPTS

Sledgehammer has a few concepts that are described here

## Registries

A registry is a construct that contains tools and tool kits.
Sledgehammer can hold any number of registries and ships with a default registry that contains all open sourced tools and kits.

There are three types of registries

### Git

The git registry is the most common one as the default registry has the same type.
As the name suggest, the metadata, tools and tool kits are stored inside a git repository.

The best example is the default registry at `https://github.com/adobe/sledgehammer-registry`.

A git registry can be remote or local and can be added with the following command:

    slh create registry git <path|url>

### File

The file registry is useful when developing locally.
In this case the registry is a simple json file.
This makes it easy to change the registry on the fly.

Example:
```
{
    "name": "<name>",
    "description": "<desc>",
    "maintainer": "<maintainer>",
    "kits": [
        {
            "name": "<name>",
            "description": "<desc>",
            "tools": [
                {
                    "name": "<toolname>"
                },
            ]
        }
    ],
    "tools":[
        {
            "name":"<name>",
            "description": "<desc>",
            "image": "<image>",
            "entry": ["<entry>"],
            "type": "<type>"
        }
    ]
}
```

It can be created with 

    slh create registry file <url>

### URL

A url registry is the last type of registry and also references a single file. In this case the file is reachable from the public internet with a simple GET request.

Sledgehammer will download and update the file on demand.

It can be created with 

    slh create registry url <url>

## Tools

In each registry there is a set of tools.
A tool represents a single docker container that can be executed. e.g. `jq` would be a tool as well as `aws`.

Tools always have the following metadata:

```
 {
    "name":"aws",
    "description": "The famous aws cli",
    "image": "mikesir87/aws-cli",
    "entry": ["aws"],
    "type": "hub"
}
```

| Variable      | Description |
| --------- | ----------- |
| Name|The name of the tool.|
| Description|Describes in a short sentence what this tool does.|
| image |The full name of the image as found on docker hub or any private registry|
| Entry|The initial command that will be called in the container. If empty will take the default of the docker container|
|Type|The type of the image, supported are `hub`, `jfrog` and `local`|

To see which tools are available you can use

    slh get tools <searchparam>

It will then show all tools available for installation.

To install a tool you can use the following command:
```
$ slh install -h
Will install the given tool on the system and symlinks it to the Sledgehammer executable

Usage:
  slh install <tool> [flags]

Flags:
      --alias string     The alias which should be used. It then can be called by this alias (e.g. py2)
      --force            True if the installation should be forced. Will overwrite previous installed tools.
  -h, --help             help for install
      --kit              True if the type is a kit that should be installed
      --version string   The version constraint that should be used (e.g. '^2' to stay on major version 2). (default "latest")
```

### Tool types

There are different types of tools because Sledgehammer needs to fetch the versions for each tool.
Unfortunately the way to fetch those versions depends on the repository those tools are located at.

#### `hub` tools

Hub tools are tools that can be found on docker hub. If the type of the tools is left empty, `hub` will be assumed.
This is the most common type of tool and should be used if the tool is present in the default registry.

#### `jfrog` tools

JFrog tools are tools that are not on docker hub but any artifactory repository. E.g. AWS offers artifactory registries for teams where images can be stored.

#### `local` tools

Local tools are again useful when developing a new tool.
With a local tool you can still add it to Sledgehammer and develop it further.
In this case Sledgehammer will only look for new version on the local system.

## Tool kits

A registry can contain tools and tool kits. Tool kits are a set of related tool under a certain group name.

The best example in this case is the `slh-dev` toolkit in the default registry that contains all tools needed to develop Sledgehammer.

The structure of a tool kit looks like this:
```
{
    "name": "<name>",
    "description": "<desc>",
    "tools": [
        {
            "name": "<toolname>",
            "version": "<version>",
            "alias": "<alias>"
        },
    ]
}
```
| Variable      | Description |
| --------- | ----------- |
|Name|The name of the tool kit. Will be used to install all tools in this kit.|
|Description|A description what this tool kit does and which problems it solves|

For each tool inside the kit:

| Variable      | Description |
| --------- | ----------- |
|Name|The name of the tool, needs to match, otherwise it cannot be installed.|
|Version|The version string for the tool that should be used. e.g. `~1.1`|
|Alias|The alias that should be installed for the tool. Will be the name of the symlink generated.|

A tool kit can only reference tools within its own registry. If you need to reference tools in other registries you need to create the same toolkit in that registry.

When installing a tool kit Sledgehammer will install all tool kits with that name regardless of the registry.

## Mounts

A mount is a local directory that Sledgehammer will mount into all tool containers.
As Sledgehammer is just a simple wrapper around docker, a mount is the same as a volume mount in the docker command.

So if the following mounts are present:

    $ slh get mounts
    Mounts
    ------
    /Users

then `/Users` will be mounted to `/Users` in the container.

That is also the reason, that you cannot mount `/`.

The reason behing this is simple: If a command contains a relative or absolute path, then Sledgehammer needs to make sure that this path is also valid inside the container.

The naive solution therefore is to mount the directories at the same location inside the container as on they are on the host.