# addlicense

The program ensures source code files have copyright license headers
by scanning directory patterns recursively.

It modifies all source files in place and avoids adding a license header
to any file that already has one.

This is a forked and slightly evolved variant of [google/addlicense](https://github.com/google/addlicense)

## Install as a Go program

    go get -u github.com/nokia/addlicense

## Usage as a Go program

    addlicense [flags] pattern [pattern ...]

    -c copyright holder (defaults to "Google LLC")
    -f custom license file (no default)
    -l license type: apache, bsd, mit, mpl (defaults to "apache")
    -y year (defaults to current year)
    -check check only mode: verify presence of license headers and exit with non-zero code if missing
    -config yaml config file: see examples/config.yml

The pattern argument can be provided multiple times, and may also refer
to single files.

## Pull from DockerHub

    docker pull nokia/addlicense-nokia

## Run as a container

    run_addlicense.sh "addlicense-options"

Where the addlicense-options are the options for addlicense (in quotes).

The actual working directory is used as the starting point to run addlicense.

## Configuration

Paths to be ignored by addlicense can be configured in a yaml config file and passed to addlicense using the `-config` option. Patterns that determine whether a file already has a license and file extension comment commenting formats can also be configured there. For an example, see the example [config.yml](https://github.com/nokia/addlicense/blob/master/examples/config.yml).

## License

Apache 2.0

