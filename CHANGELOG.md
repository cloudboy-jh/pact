# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- ASCII logo branding for CLI with colored output using lipgloss
- `--version` / `-v` flag to display logo and version
- Logo displayed on `pact init` welcome message
- Logo displayed on `pact --help` output
- Logo displayed when running `pact` without initialization
- Scoop package manager support for Windows users
- Automated Scoop bucket updates via GoReleaser on release

### Changed
- Version variable moved to `internal/ui` package for centralized branding
