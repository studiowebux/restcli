---
title: Installation
---

# Installation

## Pre-built Binaries

Download from [Releases](https://github.com/studiowebux/restcli/releases).

Extract and move to your PATH:

```bash
xattr -c restcli-macos-latest
chmod +x restcli-macos-latest
mv restcli-macos-latest /usr/local/bin/restcli
```

## From Source

### Requirements

1. Go 1.24 or later

### Build

```bash
git clone https://github.com/studiowebux/restcli
cd restcli/src
go build -o ../bin/restcli ./cmd/restcli
```

### Install to PATH

```bash
mv ../bin/restcli /usr/local/bin/
```

## Shell Completion

### Zsh (macOS)

Create completions directory:

```bash
mkdir -p ~/.zsh/completions
```

Add to `~/.zshrc`:

```bash
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

Generate completions:

```bash
restcli completion zsh > ~/.zsh/completions/_restcli
```

Reload shell:

```bash
source ~/.zshrc
```

## Verify Installation

```bash
restcli --version
```
