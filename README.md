# 🔥 gohot

**gohot** is a hot-reloading tool for Go projects that watches your source files and automatically recompiles and restarts your application when files change.

Ideal for fast development cycles in the Go ecosystem — no more manual builds or restarts!

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
go build -ldflags="-w -s" -o gohot ./gohot.go
```

## 🧠 Usage

```bash
NAME:
   gohot - Auto-reload Go apps when source files change

USAGE:
   gohot [global options] command [command options]

COMMANDS:
   init, i  create default gohot.yaml file
   version  Print the version number
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path value, -p value                                 Directory to watch (default: "./")
   --ext value, -e value [ --ext value, -e value ]        File extension to watch (default: ".go & .yaml")
   --ignore value, -i value [ --ignore value, -i value ]  File paths to ignore (default: ".git & vendor")
   --out value, -o value                                  Output binary name when compiling (default: "./appb")
   --entry value, -m value                                Main Go file entry point (default: "main.go")
   --debounce value, -d value                             Debounce time in milliseconds (default: 500)
   --envs value, -v value [ --envs value, -v value ]      Environment variables to set before go build or go run )
   --flags value, -f value [ --flags value, -f value ]    Build flags to pass to go build
   --cli value, -c value [ --cli value, -c value ]        CLI arguments to pass to the compiled binary
   --help, -h                                             show help
```

## ⚙️ Example Configuration File (gohot.yaml)

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
flags:
  - ldflags="-w -s"
cli:
  - verbose
```

**Supported Formats**

- gohot.yaml

- gohot.yml

Files are loaded automatically from:

- Current directory

- ~/.gohot/gohot.yaml

**CLI flags override config file values.**

## Examples

**Watch a directory and recompile a binary**
```bash
gohot --path ./server --entry main.go --out ./bin/server
```

**Use only ```go run``` for simple dev apps**
```bash
gohot --entry ./main.go
```

**Custom file types (e.g. ```.go```, ```.html```, ```.env```)**
```bash
gohot --ext .go,.html,.env
```

## 🚨 Validation

Before starting, gohot validates:

- Watched path exists

- File extensions start with .

- Main file (if set) exists and is a .go file

- Debounce is positive

- Output path is not a directory

## 🧑‍💻 Contributing

Pull requests are welcome! Please follow idiomatic Go style and document your changes.