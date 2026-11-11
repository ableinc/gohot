# 🔥 gohot

**gohot** is a hot-reloading tool for Go projects that watches your source files and automatically recompiles and restarts your application when files change.

Ideal for fast development cycles in the Go ecosystem — no more manual builds or restarts!

---

## 📑 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
  - [Commands](#commands)
- [Configuration File](#-configuration-file-gohotyaml)
- [Multiple Apps (Monorepos)](#-multiple-apps-monorepos)
  - [Running apps](#running-apps)
  - [Which app runs?](#which-app-runs)
  - [Precedence](#precedence)
  - [Notes](#notes)
- [Shell Completion](#-shell-completion)
- [Examples](#-examples)
- [Validation](#-validation)
- [Contributing](#-contributing)

---

## 🚀 Features

- 🔁 Auto-reload on `.go` file changes (or any extension)
- 📂 Directory and file extension filtering
- 🧠 Smart CPU usage: uses `go run` or compiles to binary based on your system
- ⚙️ Config file support (`gohot.yaml`)
- ✅ Config validation before execution
- ⏱️ Debounce file system events to avoid noisy reloads
- 🎯 Cross-platform (Linux, macOS, Windows)

---

## 📦 Installation

```bash
go install github.com/ableinc/gohot@latest
```

or **clone locally**
```bash
git clone https://github.com/yourname/gohot
cd gohot
go build -ldflags="-w -s" -o gohot .
```

## 🧠 Usage

```text
NAME:
   gohot - Auto-reload Go apps when source files change

USAGE:
   gohot [global options] command [command options]

COMMANDS:
   exec      run a named app from a multi-app gohot.yaml
   list, ls  list apps defined in gohot.yaml
   init, i   create default gohot.yaml file
   version   Print the version number
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path value, -p value                                 Directory to watch (default: "./")
   --ext value, -e value [ --ext value, -e value ]        File extension to watch (default: ".go", ".yaml")
   --ignore value, -i value [ --ignore value, -i value ]  File paths to ignore (default: ".git", "vendor")
   --out value, -o value                                  Output binary name when compiling (default: "./appb")
   --entry value, -m value                                Main Go file entry point (default: "main.go")
   --debounce value, -d value                             Debounce time in milliseconds (default: 1000)
   --envs value, -v value [ --envs value, -v value ]      Environment variables to set before go build or go run
   --env_file value, --env value                          Path to .env file to load environment variables from
   --flags value, -f value [ --flags value, -f value ]    Build flags to pass to go build
   --cli value, -c value [ --cli value, -c value ]        CLI arguments to pass to the compiled binary
   --help, -h                                             show help
```

`gohot exec` accepts the same options as the global ones.

### Commands

| Command | Description |
|---|---|
| `gohot` | Run the app defined in `gohot.yaml` (single app, or a config with exactly one entry under `apps:`) |
| `gohot exec [name]` | Run a named app from a multi-app `gohot.yaml` |
| `gohot list` (`ls`) | List the apps defined in `gohot.yaml` with their `path` and `entry` |
| `gohot init` (`i`) | Create a default single-app `gohot.yaml` in the current directory |
| `gohot init --multi` | Create a `gohot.yaml` with an `apps:` section |
| `gohot version` | Print the version |

## 📝 Configuration File (gohot.yaml)

```yaml
# gohot.yaml
path: ./
ext:
  - .go
  - .yaml
ignore:
  - .git
  - vendor
entry: main.go
out: ./appb
debounce: 500
envs:
  - GOEXPERIMENT=jsonv2
env_file: .env
flags:
  - ldflags="-w -s"
cli:
  - verbose
```

**Supported Formats**

- gohot.yaml

- gohot.yml

Gohot looks for the config file in this order and uses the first one it finds:

1. The current directory
2. Each parent directory, up to the filesystem root
3. `~/.gohot/`

(In each location `gohot.yaml` is preferred over `gohot.yml`.)

**CLI flags override config file values.**

## 🧩 Multiple Apps (Monorepos)

Define several apps under `apps:` in one `gohot.yaml`, then run each one in its own terminal window with `gohot exec <name>`.

```yaml
# gohot.yaml (at the monorepo root)

# Shared defaults: every app inherits these unless it overrides them
ext:
  - .go
  - .yaml
ignore:
  - .git
  - vendor
debounce: 500

apps:
  api:
    path: ./services/api
    entry: ./services/api/main.go
    out: ./bin/api
    env_file: .env
  cli:
    path: ./cmd/cli
    entry: ./cmd/cli/main.go
    out: ./bin/cli
    cli:
      - serve
```

### Running apps

```bash
# Terminal 1
gohot exec api

# Terminal 2
gohot exec cli

# Override a setting for one run (flags work before or after the name)
gohot exec api --debounce 200
gohot exec --debounce 200 api

# See what is defined
gohot list
# api   ./services/api   ./services/api/main.go
# cli   ./cmd/cli        ./cmd/cli/main.go

# Scaffold a multi-app config
gohot init --multi
```

### Which app runs?

| `gohot.yaml` contents | `gohot` | `gohot exec [name]` |
|---|---|---|
| No `apps:` section (single app) | Runs the app, as always | `exec` with no name runs it; a name is an error |
| `apps:` with one app | Runs it | Runs it (name optional; must match if given) |
| `apps:` with two or more apps | Error listing the app names | A valid name is required; an unknown name lists the valid ones |

```text
$ gohot
multiple apps defined: api, cli. Use `gohot exec <name>`

$ gohot exec nope
unknown app "nope". Available apps: api, cli
```

### Precedence

Settings resolve in this order, highest first:

1. CLI flag (`gohot exec api --out ./tmp/api`)
2. The app's block (`apps.api.out`)
3. Top-level value in `gohot.yaml` (shared default)
4. Built-in default

### Notes

- App names may contain letters, digits, `-` and `_`. Matching on the command line is case-insensitive.
- Relative paths (`path`, `entry`, `out`, `env_file`) resolve from the **current working directory**, so run `gohot exec` from the directory the paths are relative to (usually the monorepo root).
- Log lines are prefixed with the app name (`[api] ...`) so terminals are easy to tell apart.
- Give each app a distinct `out` path. gohot warns when two apps share one, since they would overwrite each other's binary.

## 🐚 Shell Completion

`gohot exec <TAB>` can complete app names. gohot uses urfave/cli's completion support (`--generate-bash-completion`); source the matching script from the [urfave/cli autocomplete directory](https://github.com/urfave/cli/tree/v2-maint/autocomplete) in your shell, for example in zsh:

```bash
PROG=gohot source path/to/zsh_autocomplete
```

You can also list names directly: `gohot exec --generate-bash-completion`.

## 💡 Examples

**Watch a directory and recompile a binary**
```bash
gohot --path ./server --entry main.go --out ./bin/server
```

**Use only ```go run``` for simple dev apps**
```bash
gohot --entry ./main.go
```

**Run one app from a monorepo config**
```bash
gohot exec api
```

**Custom file types (e.g. ```.go```, ```.html```, ```.env```)**
```bash
gohot --ext .go,.html,.env
```

**Load environment variables from a .env file**
```bash
gohot --env_file .env
# or use the short alias
gohot --env .env
```

**Combine .env file with additional CLI environment variables (CLI envs override file)**
```bash
gohot --env_file .env --envs DEBUG=true
```

## 🚨 Validation

Before starting, gohot validates:

- Watched path exists

- File extensions start with .

- Main file (if set) exists and is a .go file

- Env file (if set) exists

- Debounce is positive

- Output path is not a directory

- (Multi-app) app names are valid and each app block is a mapping

## 🧑‍💻 Contributing

Pull requests are welcome! Please follow idiomatic Go style and document your changes.