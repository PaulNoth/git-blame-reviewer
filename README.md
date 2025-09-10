# git-review-blame

A Git tool that shows GitHub pull request approvers for each line of code, instead of the original commit authors that `git blame` shows.

## Features

- 🔍 Shows who approved the changes for each line of code
- ⏰ Displays approval timestamps  
- 🎨 Colorized terminal output
- 🚀 Built with Go - single binary, no dependencies

## Installation

### From Source

```bash
git clone https://github.com/your-username/git-review-blame.git
cd git-review-blame
make build
```

### Binary

The compiled binary will be available as `git-review-blame`.

## Usage

### Basic Usage

```bash
git-review-blame -file src/main.go
```

### With GitHub Token

```bash
git-review-blame -file src/main.go -token ghp_xxxxxxxxxxxx
```

### Using Environment Variable

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
git-review-blame -file src/main.go
```

### Command Line Options

- `-file` - Path to file to analyze (required)
- `-token` - GitHub personal access token 
- `-verbose` - Enable verbose output
- `-help` - Show help message

## GitHub Token

You'll need a GitHub personal access token with `repo` scope to access pull request information.

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate new token with `repo` scope
3. Use it with `-token` flag or set `GITHUB_TOKEN` environment variable

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

1. Analyzes the specified file using `git blame` to get commit hashes per line
2. For each commit, queries GitHub API to find the associated pull request  
3. Retrieves PR approval information (approvers and timestamps)
4. Formats output similar to `git blame` but shows approver info instead

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make check` to ensure quality
5. Submit a pull request