# 🔥 gohot

**gohot** is a hot-reloading tool for Go projects that watches your source files and automatically recompiles and restarts your application when files change.

Ideal for fast development cycles in the Go ecosystem — no more manual builds or restarts!

---

## 🚀 Features

- 🔁 Auto-reload on `.go` file changes (or any extension)
- 📂 Directory and file extension filtering
- 🧠 Smart CPU usage: uses `go run` or compiles to binary based on your system
- ⚙️ Config file support (`gohot.yaml`, `gohot.json`, etc.)
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
go build -ldflags="-w -s" -o gohot ./cmd/gohot/gohot.go
```

## 🧠 Usage

```bash
gohot --path ./cmd/api --ext .go,.yaml --entry main.go --out ./build/app
```


| Flag          | Alias | Description                  | Default |
| ------------- | ----- | -----------------------------| ------- |
| --path        | -p    | Directory to watch           | ./      |
| --ext         | -e    | File extensions to watch (comma-separated) | .go,.yaml |
| --entry       | -m    | Main Go file to run          | (auto-detect) |
| --out         | -o    | Output binary name/path      | ./appb |
| --debounce    | -d    | Debounce time in ms          | 500 |


## ⚙️ Configuration File (gohot.yaml)

```yaml
# gohot.yaml
path: ./
ext: .go,.yaml
entry: main.go
out: ./appb
debounce: 500
```

**Supported Formats**

- gohot.yaml

- gohot.json

- gohot.toml

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

## 📋 TODO / Future Ideas

- --init to create a sample gohot.yaml

- Export merged config with --export

- Logging to .gohot.log

- WebSocket/HTTP integration for browser auto-refresh

## 🧑‍💻 Contributing

Pull requests are welcome! Please follow idiomatic Go style and document your changes.