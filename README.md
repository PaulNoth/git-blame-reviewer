# git-review-blame

A Git tool that mimics the behavior of `git blame` but shows GitHub pull request approvers for each line of code instead of the original commit authors.

## Features

- 🔍 Shows who approved the changes for each line of code
- ⏰ Displays approval timestamps instead of commit timestamps
- 📋 **Identical interface** to `git blame` - same command-line arguments and output format
- 🎨 Supports both human-readable and porcelain output formats
- 🚀 Built with Go - single binary, no external dependencies
- ⚡ Optimized with commit-level caching to minimize GitHub API calls

## Installation

### From Source

```bash
git clone <repository-url>
cd git-review-blame
make build
```

### Binary

The compiled binary will be available as `git-review-blame`.

## Usage

**git-review-blame uses the exact same command-line interface as git blame:**

### Basic Usage

```bash
git-review-blame src/main.go
```

### With Line Range

```bash
git-review-blame -L 10,20 src/main.go
```

### Porcelain Format (Machine-Readable)

```bash
git-review-blame -porcelain src/main.go
```

### Show Email Addresses

```bash
git-review-blame -show-email src/main.go
```

### Command Line Options

- `-L <start>,<end>` - Show only lines in given range (same as git blame)
- `-porcelain` - Show in a format designed for machine consumption
- `-show-email` - Show author email instead of author name  
- `-help` - Show help message

**Note:** The file path is provided as a positional argument, just like `git blame`.

## GitHub Token

You'll need a GitHub personal access token with `repo` scope to access pull request information.

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate new token with `repo` scope
3. Set the `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
git-review-blame src/main.go
```

The tool requires the token to be set as an environment variable (no command-line flag for the token).

## Development

### Prerequisites

- Go 1.21+
- Make
- golangci-lint (install with `make install-tools`)

### Building

```bash
make build
```

### Testing

```bash
make test
make test-coverage  # with coverage report
```

### Linting

```bash
make lint
```

### All Checks

```bash
make check  # runs tests + linting
```

## How It Works

1. **Git Repository Detection** - Finds the git repository root and validates it's a git directory
2. **Repository Info Extraction** - Extracts GitHub owner/repository name from git remote origin
3. **Git Blame Execution** - Runs `git blame` on the specified file to get commit hashes per line  
4. **GitHub API Integration** - For each unique commit hash:
   - Queries GitHub API to find the associated pull request
   - Retrieves PR approval information (approvers and timestamps)
   - Caches results to avoid duplicate API calls
5. **Output Formatting** - Displays results in the same format as `git blame`, but with:
   - PR approver name instead of commit author
   - PR approval timestamp instead of commit timestamp
   - Falls back to original commit info if no PR/approval found

## Output Format

The output format is identical to `git blame`:

```
a1b2c3d4 (Jane Smith    2023-12-01 14:30:25  1) package main
b2c3d4e5 (Bob Wilson    2023-12-01 09:15:10  2) 
c3d4e5f6 (Alice Johnson 2023-11-30 16:45:30  3) import "fmt"
```

Where:
- `a1b2c3d4` - Commit hash (shortened)
- `Jane Smith` - **PR approver name** (not commit author)
- `2023-12-01 14:30:25` - **PR approval timestamp** (not commit timestamp)
- `1` - Line number
- `package main` - Line content

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make check` to ensure quality
5. Submit a pull request