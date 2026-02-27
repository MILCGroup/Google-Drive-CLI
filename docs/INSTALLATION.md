# Installation Guide

## Install Script (Recommended)

```bash
# Install to ~/.local/bin (ensure it's on your PATH)
curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Drive-CLI/master/install.sh | bash
```

## Homebrew (Tap)

```bash
brew tap dl-alexandre/tap
brew install gdrv
```

## Download Binary

Download the latest release from the [releases page](https://github.com/dl-alexandre/Google-Drive-CLI/releases).

```bash
# Make executable and move to PATH
chmod +x gdrv
sudo mv gdrv /usr/local/bin/
```

## Build from Source

```bash
git clone https://github.com/dl-alexandre/Google-Drive-CLI.git
cd Google-Drive-CLI
go build -o gdrv ./cmd/gdrv
```

## Verify Installation

```bash
gdrv --version
```
