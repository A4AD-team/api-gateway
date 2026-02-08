# Git Flow Configuration for API Gateway

This repository uses Lefthook to enforce Git Flow conventions and protect important branches.

## Installation

1. Install Lefthook:
```bash
# macOS
brew install lefthook

# Or download from GitHub releases
# https://github.com/evilmartians/lefthook/releases
```

2. Install hooks:
```bash
lefthook install
```

## Git Flow Rules

### Protected Branches
Direct pushes to these branches are **blocked**:
- `main` / `master` - Production branch
- `develop` - Development integration branch  
- `release/*` - Release preparation branches
- `hotfix/*` - Production hotfix branches

### Allowed Branch Names
Branch names must follow Git Flow patterns:
- `feature/<description>` - New features
- `bugfix/<description>` - Bug fixes
- `hotfix/<description>` - Production hotfixes
- `release/<version>` - Release preparation
- `develop` - Development branch
- `main`/`master` - Production branches

### Commit Message Format
Commits must follow **Conventional Commits** format:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Code style (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Tests
- `chore` - Maintenance tasks
- `perf` - Performance improvements
- `ci` - CI/CD changes
- `build` - Build system changes
- `revert` - Revert previous commit

**Examples:**
```
feat(auth): add JWT token validation
fix: resolve memory leak in rate limiter
docs(api): update authentication endpoints
```

## Hooks

### Pre-commit
- Runs `go fmt` and `go vet`
- Runs tests on staged files
- Validates Go code quality

### Commit-msg
- Validates commit message format
- Enforces conventional commits
- Checks message length limits

### Pre-push
- Validates branch naming conventions
- Runs full test suite
- Blocks pushes to protected branches

## Local Development

For stricter local validation, copy the example config:
```bash
cp lefthook-local.yml.example lefthook-local.yml
```

This enables:
- Additional code formatting checks
- Protected branch blocking
- Stricter validation rules

## Bypassing Hooks (Not Recommended)

If absolutely necessary, you can bypass hooks:
```bash
git commit --no-verify -m "message"
git push --no-verify
```

**Use with caution** - this skips all safety checks.

## Workflow Example

1. Create feature branch:
```bash
git checkout -b feature/jwt-authentication
```

2. Make changes and commit:
```bash
git add .
git commit -m "feat(auth): add JWT token validation"
```

3. Push and create PR:
```bash
git push origin feature/jwt-authentication
# Create pull request in GitHub/GitLab
```

4. Merge via PR (never push directly to main branches)