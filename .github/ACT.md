# Running GitHub Actions Locally with act

This guide explains how to test GitHub Actions workflows locally using [act](https://github.com/nektos/act) before pushing to GitHub.

## Why Use act?

- **Test workflows locally** before pushing
- **Faster feedback loop** - no waiting for CI
- **Save CI minutes** and reduce commit noise
- **Debug workflow issues** more easily

## Installation

### Linux

```bash
# Download and install
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Or using package managers
# Arch Linux
sudo pacman -S act

# Ubuntu/Debian (via GitHub releases)
curl -s https://api.github.com/repos/nektos/act/releases/latest \
  | grep "browser_download_url.*linux_amd64.tar.gz" \
  | cut -d : -f 2,3 \
  | tr -d \" \
  | wget -qi - -O /tmp/act.tar.gz
sudo tar -xzf /tmp/act.tar.gz -C /usr/local/bin act
```

### macOS

```bash
brew install act
```

### Windows

```bash
choco install act-cli
# or
scoop install act
```

## Quick Start

### 1. Run All Workflows

```bash
# Test the default event (push)
act

# Test pull request event
act pull_request

# Test specific workflow
act -W .github/workflows/ci.yml
```

### 2. Run Specific Job

```bash
# Run only the lint-and-test job
act -j lint-and-test

# Run with specific Go version
act -j lint-and-test --matrix go-version:1.24
```

### 3. Debug Mode

```bash
# Verbose output
act -v

# Very verbose (includes Docker commands)
act -vv

# List available workflows/jobs without running
act -l
```

## Common Commands for This Project

### Test CI Workflow

```bash
# Run the full CI workflow
act push -W .github/workflows/ci.yml

# Run with specific Go version matrix
act push -W .github/workflows/ci.yml --matrix go-version:1.22

# Run only linting
act push -W .github/workflows/ci.yml -j lint-and-test
```

### Test Release Workflow

```bash
# Test semantic-release (dry-run)
act push -W .github/workflows/release.yml --secret GITHUB_TOKEN=your_token
```

## Configuration

### .actrc File

Create `.actrc` in the project root for default options:

```bash
# Use medium-sized container (recommended)
-P ubuntu-latest=catthehacker/ubuntu:act-latest

# Or use larger container with more tools
# -P ubuntu-latest=catthehacker/ubuntu:full-latest

# Bind local Docker socket (if needed)
# --bind

# Verbose output by default
# -v
```

### Secrets

Create `.secrets` file (add to .gitignore):

```bash
GITHUB_TOKEN=your_github_token_here
```

Then run with:

```bash
act --secret-file .secrets
```

## Container Images

act uses Docker containers to simulate GitHub runners:

- **Micro** (~200MB): Basic tools only
- **Medium** (~500MB): Most common tools (recommended)
- **Large** (~12GB): Full GitHub runner equivalent

Specify with `-P` flag:

```bash
# Medium (recommended)
act -P ubuntu-latest=catthehacker/ubuntu:act-latest

# Large (if you need specific tools)
act -P ubuntu-latest=catthehacker/ubuntu:full-latest
```

## Troubleshooting

### Docker Permission Issues

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Logout and login again
```

### Container Platform Issues

```bash
# Specify platform explicitly
act --container-architecture linux/amd64
```

### Missing Tools in Container

Some actions may require tools not in the default container. Options:

1. Use larger container: `-P ubuntu-latest=catthehacker/ubuntu:full-latest`
2. Install tools in workflow with `run: sudo apt-get install ...`
3. Use official action containers when possible

### Workflow Not Running

```bash
# Check available workflows
act -l

# Verify workflow syntax
act --dryrun -W .github/workflows/ci.yml
```

## Limitations

- Some GitHub-specific features don't work (e.g., artifact upload/download)
- Secrets need to be provided manually
- Matrix builds run sequentially (not parallel)
- Some third-party actions may not work perfectly

## Best Practices

1. **Test before pushing**: Always run `act` before `git push`
2. **Use .actrc**: Set up defaults to avoid repetitive flags
3. **Keep secrets safe**: Never commit `.secrets` file
4. **Use medium container**: Good balance of size vs. functionality
5. **Check the logs**: Use `-v` for detailed output when debugging

## Example Workflow

```bash
# 1. Make code changes
vim server.go

# 2. Run tests locally
go test ./...

# 3. Test CI with act
act push -W .github/workflows/ci.yml

# 4. If successful, commit and push
git add .
git commit -m "feat: add new feature"
git push origin main
```

## Resources

- [act GitHub Repository](https://github.com/nektos/act)
- [act Documentation](https://nektosact.com/)
- [Container Images](https://github.com/catthehacker/docker_images)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

---

**Pro Tip**: Add `act` to your pre-push Git hook to automatically test workflows before pushing!
