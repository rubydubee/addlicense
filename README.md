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
    -s silent mode: do not log anything to the console
    -y year (defaults to current year)
    -check check only mode: verify presence of license headers and exit with non-zero code if missing
    -config yaml config file: see examples/config.yml

The pattern argument can be provided multiple times, and may also refer to single files.

For example to check a project for license headers:

    addlicense -check /path/to/my/project
    

## Run as a container

Pull from dockerhub

    docker pull nokia/addlicense-nokia

The image can now be run using the `run_addlicense` script:

    ./run_addlicense.sh "addlicense-options"

Where the addlicense-options are the options for addlicense (in quotes).

The current working directory is used as the starting point to run addlicense.

To use with the default configuration and BSD 3 Clause copyright texts use: 
    
    ./run_addlicense.sh "-f /copyright-texts/bsd-3-clause"

or 
    
    docker run -e OPTIONS="-f /copyright-texts/bsd-3-clause" --rm -it -v $(pwd):/myapp nokia/addlicense-nokia:latest

## Configuration

Paths to be ignored by addlicense can be configured in a yaml config file and passed to addlicense using the `-config` option. Patterns that determine whether a file already has a license and file extension comment commenting formats can also be configured there. 

For more concrete examples, see the example [config.yml](https://github.com/nokia/addlicense/blob/master/examples/config.yml).

### Ignored paths
The ignored paths can be configured using glob patterns.
Example of excluding everything inside the tests and docs folders:

```yaml
ignorePaths:
  - 'tests/**'
  - 'docs/**'
```

### Comment patterns

Example of configuring comment patterns for some file extensions:
```yaml
fileExtensions:
  - extensions: ['.js', '.mjs', '.cjs', '.jsx', '.tsx', '.css', '.tf', '.ts']
    top: '/**'
    mid: ' * '
    bot: '*/'
```
The "top" value speciefies how to begin the comment, the "mid" value how to continue the comment on new lines and the "bot" how to close the comment.

### Has-license patterns
Patterns that determine whether a file already has a license can be configured like this:
```yaml
hasLicensePatterns: 
  - 'copyright'
  - 'license'
```

## License

Apache 2.0
