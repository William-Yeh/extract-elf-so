extract-elf-so
==============

[![Circle CI](https://circleci.com/gh/William-Yeh/extract-elf-so.svg?style=shield)](https://circleci.com/gh/William-Yeh/extract-elf-so) [![Build Status](https://travis-ci.org/William-Yeh/extract-elf-so.svg?branch=master)](https://travis-ci.org/William-Yeh/extract-elf-so)


This program extracts .so files from specified ELF executables, and packs them in a tarball.

It was initially invented as a tool to generate minimal Docker images. See the slides ["Quest for minimal Docker images"](http://william-yeh.github.io/docker-mini) for more details.


## Usage


```
Extract .so files from specified ELF executables, and pack them in a tarball.

Usage:
  extract-elf-so  [options]  [(--add <so_file>)...]  <elf_files>...
  extract-elf-so  --help
  extract-elf-so  --version

Options:
  -d <dir>, --dest <dir>          Destination dir in the tarball to place the elf_files;
                                    [default: /usr/local/bin].
  -n <name>, --name <name>        Generated tarball filename (without .gz/.tar.gz);
                                    [default: rootfs].
  -a <so_file>, --add <so_file>   Additional .so files to be appended into the tarball.
  -s <so_dir>, --sodir <so_dir>   Directory in the tarball to place additional .so files;
                                    [default: /usr/lib].
  -z                              Compress the output tarball using gzip.
  --nss-net                       Install networking stuff of NSS;  [default: false].
  --cert                          Install necessary root CA certificates;  [default: false].

```

## Download and install

See the [releases](https://github.com/William-Yeh/extract-elf-so/releases) page.

Or, use the `install.sh` script:

```bash
$ ./install.sh  [VERSION]  [INSTALL PATH]
```

Or, install the latest version to `/usr/local/bin` with the following one-liner:

```bash
$ curl -sSL http://bit.ly/install-extract-elf-so \
    | sudo bash
```


## Runtime dependencies

The program expects the following executables in `$PATH`:

- `cp` - copy files and directories.
- `ldd` - print shared library dependencies.
- `tar` - The GNU version of the tar archiving utility.


## Build

Build the executable for your platform (before compiling, please make sure that you have [Go](https://golang.org/) compiler installed):

```
$ go install github.com/docopt/docopt-go
$ go build
```

Or, build the *linux-amd64* executables with Docker:

```
$ ./build.sh
```

Or, build the *linux-amd64* executables with Vagrant:

```
$ vagrant up
```

It will place the `extract-elf-so_linux-amd64` and `extract-elf-so_static_linux-amd64` executables into the `dist` directory.


## Caveats

This program only handles parts of the [*Name Service Switch (NSS)*](http://www.gnu.org/software/libc/manual/html_node/Name-Service-Switch.html) stuff. If this is important for you, read the article: ["Creating minimal Docker images from dynamically linked ELF binaries"](http://blog.oddbit.com/2015/02/05/creating-minimal-docker-images/).


## History

- 0.6 - Use `en_US.UTF-8` locale to avoid messy `ldd` output.
- 0.5 - Treat some NSS .so files as "optional".
- 0.4 - Add "/etc/ssl/certs/ca-certificates.crt" to rootfs.
- 0.3 - Fix "not a dynamic executable" handling.
- 0.2 - Handle parts of [*Name Service Switch (NSS)*](http://www.gnu.org/software/libc/manual/html_node/Name-Service-Switch.html) stuff.
- 0.1 - Initial release.


## Author

William Yeh, william.pjyeh@gmail.com

## License

Apache License V2.0.  See [LICENSE](LICENSE) file for details.
