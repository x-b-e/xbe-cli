# xbe-cli

A Go-based CLI for XBE.

## Install
Download the release archive for your OS/arch from the GitHub Releases page, extract it, and move `xbe` into your PATH.

Example (macOS arm64):

```
VERSION=0.1.0
curl -L https://github.com/x-b-e/xbe-cli/releases/download/v${VERSION}/xbe_${VERSION}_darwin_arm64.tar.gz | tar -xz
sudo mv xbe /usr/local/bin/xbe
```

## Update
Repeat the install steps with the latest release version. Checksums are published alongside each release.

## Build
make build

## Run
./xbe --help
./xbe version
./xbe view newsletters list
./xbe view brokers list
