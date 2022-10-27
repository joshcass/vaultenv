# vaultenv

This tool is composed of two parts that together create a plugin for [direnv][1] which allows it to automatically fetch secrets from a vault instance and automatically load/unload them for a project.
The secrets are written to files in memory (`/dev/shm` on GNU/Linux and a ramdisk on macOS) so they are not added to a persistent file on disk. It's not 100% perfect or foolproof but it's an improvement over
populating a `.env` or `.envrc` with a bunch of secrets that persist on disk.

The two parts are:

1. A binary application written in Go that takes a path to some vault secrets and writes them to a file in memory.
2. A shell script for direnv that executes the binary and tells direnv to load the secrets from the resulting file.

## Usage

Assuming you have some secrets in a vault kv v2 store at "kv/global/test" that look like:

```shell
key: SECRET_ONE value: "secrets"
key: SECRET_TWO value: "more secrets"
```

You would add the following to your `.envrc`:

```shell
vaultenv "kv/global/test"
```

When you `cd` into your project vaultenv will fetch those secrets and use direnv to export them as ENV variables named `SECRET_ONE` and `SECRET_TWO` with their respective values.

## Installation

Download the release binary that fits your platform and operating system and move it to a location that is on your `PATH` such as `/usr/local/bin`
Rename or alias the binary as `venv`
Make sure it is executable

Download the `vaultenv.sh` script and move it to `$HOME/.config/direnv/lib`
