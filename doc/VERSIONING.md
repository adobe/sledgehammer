# Versioning

Sledgehammer supports semantic versioning for all tools.
However there is a special handling for container versions.

A tool needs to have a version that complies to the [semantic versioning standard](https://semver.org/).
Sledgehammer can also work with tools violating the standard (eg. `latest`) but proper versioning support is not available in that case.

A version for a tool consists of two parts, the tool version and the container version.

The tool version represents the version of the tool itself e.g. `1.2.0`.

The container version is the version of the tool container itself, it can be increased if there is i.e. an update on system tools or the container itself.
The container version is appended to the tool version as a prerelease version e.g. `-1`

The whole version for the tool would then be `1.2.0-1`.
If we update the container but not the tool, we will then upgrade to `1.2.0-2`.

Sledgehammer will sort the version including the prerelease versions so you should restrain from publishing a version `1.2.0` because `1.2.0-1 < 1.2.0-2 < 1.2.0 < 1.3.0-1`

If we update the tool itself we reset the container version to `1.3.0-1`


Sledgehammer will make sure that always the newest version will be used if possible. If Sledgehammer detects a new version it will download it in the background and uses that during the next execution.