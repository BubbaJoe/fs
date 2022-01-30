# FS Store

## Usage

```sh
## Start server
fs-store server [flags]

## list files form server
fs-store list <localFileName> ... [flags]

## upload file to server
fs-store upload <localFileName> ... [flags]

## delete file from server
fs-store delete <serverFileName> ... [flags]
```

## Setup

### Build binary

```sh
make build
```

### Testing

```sh

make test

## includes test / vet / lint
make check
```

### Docker

```sh
make docker-build
```

### Release

```sh
make release
```


## Scope

- [ ] Cmd Test (e2e)
- [x] Client Unit Test
- [x] Client Integration Test
- [ ] Server Integration Test
- [x] Server Unit Test
- [x] Dockerfile
- [x] Makefile - test/vet/lint/release
