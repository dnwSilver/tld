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

## GitLab protection in Settings

The `branches` check requires explicit protected-branch rules for `master` and
`dev`. The `tags` check requires all three protected-tag patterns:
`release/*.*.*`, `hotfix/*.*.*`, and `v*.*.*`. These checks verify rule presence;
`protect` continues to validate branch access levels and force-push settings.

Press `u` on a selected project to open operations, then choose **Protect
master/dev branches** or **Protect release/hotfix/version tags**. These operations
create only missing rules and preserve existing rules. New branch rules allow
Developer merges, Maintainer pushes, and disable force pushes. New tag rules
allow Maintainers to create tags. The GitLab token needs Maintainer access and
`api` scope. After applying an operation, project checks refresh automatically.

## Release timeline

Stable tags `vX.Y.Z` and `release/X.Y.Z` appear as releases (rocket).
`hotfix/X.Y.Z` appears as a hotfix (`󰀯`). Versions contain three numeric
components; unrelated tags and prerelease tags are excluded. When timeline
marks overlap in one displayed cell, the hotfix icon takes precedence.
Existing date-only cache entries remain readable as releases; refresh Releases
to reload tag names and classify hotfixes.

The rocket and ambulance columns between the project and the first month count
release and hotfix tags in the selected timeline period. Each column is three
characters wide. Multiple tags in the same timeline slot count separately.
