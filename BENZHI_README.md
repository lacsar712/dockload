# Dockload

码头吊具称重流

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh dockload linux/amd64
./build_benzhi_docker.sh dockload linux/arm64
```