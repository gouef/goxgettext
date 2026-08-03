# goxgettext

A lightweight gettext-style extractor for Go projects.

goxgettext scans Go source files and GoHTML/HTML templates, collects translatable strings, and generates POT/PO catalogs for your project.

## ✨ Features

- Extracts messages from Go calls such as `gettext`, `t`, `T`, and `i18n`
- Scans GoHTML and HTML templates for visible text and translation calls
- Supports recursive directory scanning
- Generates POT and PO output
- Can create language-specific PO files and a POTFILES list from `LINGUAS`
- Supports custom keywords and additional file extensions

## 🚀 Installation

Install the latest release binary with Go:

```bash
go install github.com/gouef/goxgettext@latest
```

Install from the repository source:

```bash
make build
```

The binary will be created at ./bin/goxgettext and can be run directly.

Install from a downloaded release binary:


### Linux
```bash
curl -L -o /tmp/goxgettext https://github.com/gouef/goxgettext/releases/latest/download/goxgettext-linux-amd64
sudo install -m 0755 /tmp/goxgettext /usr/local/bin/goxgettext
/usr/local/bin/goxgettext --help
```


### macOS
```bash
curl -L -o /tmp/goxgettext https://github.com/gouef/goxgettext/releases/latest/download/goxgettext-darwin-amd64
install -m 0755 /tmp/goxgettext /usr/local/bin/goxgettext
/usr/local/bin/goxgettext --help
```

### Windows
```powershell
# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/gouef/goxgettext/releases/latest/download/goxgettext-windows-amd64.exe -OutFile "$env:USERPROFILE\bin\goxgettext.exe"
```

After installation, you can run `goxgettext` directly from your shell.

## 🧪 Usage

Run the extractor against a source tree:

```bash
./bin/goxgettext .
```

Generate a POT catalog:

```bash
./bin/goxgettext --output messages.pot .
```

Generate a PO catalog:

```bash
./bin/goxgettext --format po --output messages.po .
```

Generate language files from a LINGUAS file:

```bash
./bin/goxgettext --output-dir locale .
```

Generate a single language file explicitly:

```bash
./bin/goxgettext --output-dir locale --language cs .
```

Build release binaries for multiple platforms:

```bash
make release
```

## 📦 Release binaries

When a tag matching `v*` is pushed, the GitHub Actions workflow in [.github/workflows/release.yml](.github/workflows/release.yml) builds release artifacts for Linux, macOS, and Windows and uploads them to the GitHub Release.

You can also trigger the workflow manually from the Actions tab and choose a tag name.

Download the appropriate binary for your platform from the release page and run it directly.

## ⚙️ Useful flags

- --output: write the generated catalog to a file
- --format: select pot or po output
- --keyword: add custom translation function names
- --extension: include additional file extensions to scan
- --output-dir: write generated language files to a directory
- --language: generate or update a specific language file

## 🛠️ Development

Run the test suite:

```bash
make test
```

Generate a coverage report:

```bash
make coverage
```

## Contributors

<div>
<a href="https://github.com/gouef/goxgettext/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=gouef/goxgettext" />
</a>
</div>

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

