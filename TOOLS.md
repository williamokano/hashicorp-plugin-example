# Tool Versions

This document lists all the development tools used in this project and their versions.
Both `make deps` and GitHub Actions CI use the same versions to ensure consistency.

## Required Tools

### Core Build Tools
- **Go**: 1.24.6
- **protoc-gen-go**: v1.36.8
- **protoc-gen-go-grpc**: v1.5.1

### Linting and Code Quality
- **golangci-lint**: v2.4.0
- **gosec**: v2.21.4
- **govulncheck**: v1.1.3
- **gocyclo**: v0.6.0
- **ineffassign**: v0.1.0

### Testing
- **counterfeiter**: v6.11.2

## Installation

To install all required tools with the correct versions:

```bash
make deps
```

To check installed tool versions:

```bash
make check-tools
```

## Version Alignment

The following files contain tool version specifications:
- `Makefile` - `deps` target installs specific versions
- `.github/workflows/ci.yml` - CI pipeline uses the same versions

## Updating Tool Versions

When updating tool versions:
1. Update version in `Makefile` deps target
2. Update version in `.github/workflows/ci.yml`
3. Update this document
4. Test locally with `make deps && make check-tools`
5. Run `make quality` to ensure everything still works