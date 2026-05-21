# Contributing

## Requirements

- Go 1.26.3 or newer
- Make
- C toolchain for the SQLCipher-backed SQLite driver

## Setup

```sh
make deps
```

## Development

```sh
make dev
```

## Build

```sh
make build
```

## Tests

```sh
make test
```

## Local database

The application stores its encrypted local database at the OS user config path:

- macOS: `~/Library/Application Support/tld/tld.db`
- Linux: `$XDG_CONFIG_HOME/tld/tld.db` or `~/.config/tld/tld.db`
- Windows: `%AppData%\tld\tld.db`

The database is opened with a passphrase entered interactively at startup. The passphrase is not stored by the application. If it is lost, the local database cannot be recovered.
