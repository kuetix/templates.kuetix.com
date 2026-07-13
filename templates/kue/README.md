# kue template bundle

This directory mirrors the checked-in `kue` CLI source layout so the template archive includes the real bundled workflows and related bootstrap files.

Included assets:

- `cmd/cli/main.go`
- `cmd/modules/main.go`
- `embed.go`
- `Makefile`
- `install.sh`
- `modules/modules.go`
- `workflows/**`
- `runtime/bin/workflows/**`

This bundle is source-faithful: it follows the `kue` Makefile flow (`make build`, `make test`, `make install`, `make clean`) rather than the older `runner.sh` pattern referenced by some legacy generic templates.
