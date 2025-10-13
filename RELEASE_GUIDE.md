# EggyByte Core Release Guide

## Quick Start

### 1. Set up GitHub Token

Before creating a release, you need to set up a GitHub Personal Access Token:

1. Go to [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select the following scopes:
   - `repo` (Full control of private repositories)
   - `write:packages` (Write packages to GitHub Package Registry)
4. Copy the generated token
5. Set the environment variable:

```bash
export GITHUB_TOKEN=your_token_here
```

### 2. One-Click Release

```bash
# Create a complete GitHub release with binaries
make create-release VERSION=v1.0.0

# Or force create if tag already exists
make create-release-force VERSION=v1.0.0
```

### 3. Manual Steps (Alternative)

If you prefer manual control:

```bash
# 1. Prepare the release
make prepare-release VERSION=v1.0.0

# 2. Push changes to GitHub
make github-update

# 3. Create and push tag
make github-release VERSION=v1.0.0

# 4. Create release with goreleaser
goreleaser release --clean
```

## What the Release Includes

### Binaries
- `ebcctl-darwin-amd64` - macOS Intel
- `ebcctl-darwin-arm64` - macOS Apple Silicon
- `ebcctl-linux-amd64` - Linux x86_64
- `ebcctl-linux-arm64` - Linux ARM64
- `ebcctl-windows-amd64.exe` - Windows x86_64
- `ebcctl-windows-arm64.exe` - Windows ARM64

### Archives
- Source code archives for each platform
- Checksums for verification
- Installation instructions

### GitHub Release Features
- Automatic changelog generation
- Release notes with installation instructions
- Binary downloads for all platforms
- Checksum verification files

## Scripts Overview

### `scripts/sh/create-release.sh`
Complete release automation script that:
- Validates environment and dependencies
- Runs tests (optional)
- Pushes changes to GitHub
- Creates and pushes git tag
- Uses goreleaser to create GitHub release
- Provides detailed feedback and next steps

### `scripts/sh/github-release.sh`
Creates and pushes git tags to GitHub (without binaries)

### `scripts/sh/github-update.sh`
Pushes current changes to GitHub repository

### `scripts/sh/prepare-release.sh`
Prepares project for release (tests, version updates, tag creation)

## Troubleshooting

### Common Issues

1. **GITHUB_TOKEN not set**
   ```
   Error: GITHUB_TOKEN environment variable is not set
   ```
   Solution: Set the environment variable as described above

2. **Tag already exists**
   ```
   Error: Tag v1.0.0 already exists
   ```
   Solution: Use `--force` flag or `make create-release-force`

3. **Tests failing**
   ```
   Error: Tests failed
   ```
   Solution: Fix failing tests or use `--skip-tests` flag

4. **goreleaser not found**
   ```
   Error: goreleaser is not installed
   ```
   Solution: `go install github.com/goreleaser/goreleaser@latest`

### Environment Requirements

- Go 1.24.5+
- goreleaser (latest)
- git
- GITHUB_TOKEN with repo and write:packages permissions

## Release Checklist

- [ ] All tests passing
- [ ] Version updated in relevant files
- [ ] CHANGELOG.md updated
- [ ] README.md updated (if needed)
- [ ] GITHUB_TOKEN set
- [ ] goreleaser installed
- [ ] Git repository clean
- [ ] Remote origin points to correct repository

## Post-Release

After creating a release:

1. Verify binaries work on different platforms
2. Test installation instructions
3. Update documentation if needed
4. Announce the release to users
5. Monitor for any issues

## Support

For issues or questions:
- Check the [GitHub Issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- Review the [Documentation](docs/)
- Contact the development team
