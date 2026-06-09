# AtomGit CLI (ag)

AtomGit command line tool, developed based on GitHub CLI (gh).

## Installation

```bash
# Buuild from the source code.
go build ./cmd/ag

# Install to $GOPATH/bin.
go install ./cmd/ag
```

## Configuration

Before using the tool, configure an access token. Create a token file as follows:

### (Recommended) XDG-compliant

```sh
$XDG_CONFIG_HOME/ag-cli/token.json
```

Default path: `~/.config/ag-cli/token.json`

### Compatible with the Old Method

```sh
~/.atomgit_personal_token.json
```

File content:

```json
{
  "access_token": "your-token-here",
  "user": "your-username"
}
```

## Command

### Authentication

```bash
# Check the authentication status.
ag auth status

# Display the current token.
ag auth token
```

### Repository

```bash
# List repositories.
ag repo list

# View repository details.
ag repo view owner/repo

# Create a repository.
ag repo create my-project --public --description "My project"

# Clone a repository.
ag repo clone owner/repo
ag repo clone owner/repo --branch develop

# Fork a repository.
ag repo fork owner/repo
ag repo fork owner/repo --name my-fork --public

# Delete a repository.
ag repo delete owner/repo --yes
```

### Pull Request (PR)

```bash
# List PRs.
ag pr list owner/repo
ag pr list owner/repo --state closed

# View PRs.
ag pr view owner/repo 123

# Create a PR.
ag pr create owner/repo --title "Fix bug" --body "Description" --base master --head feature-branch

# Close a PR.
ag pr close owner/repo 123
```

### PR Comment

```bash
# Creatr a comment.
ag pr comment create owner/repo 123 --body "LGTM!"
ag pr comment create owner/repo 123 --body-file review.md

# View all commands (tree-struction).
ag pr comment view owner/repo 123

# Edit a command (interactive mode).
ag pr comment edit owner/repo 123 456
ag pr comment edit owner/repo 123 456 --body "Updated comment"

# Delete a command.
ag pr comment delete owner/repo 123 456
ag pr comment delete owner/repo 123 456 --yes

# Reply to a comment (specific to PRs).
ag pr comment reply owner/repo 123 456 --body "Thanks for the feedback!"
```

### Issue

```bash
# List issues.
ag issue list owner/repo
ag issue list owner/repo --state all

# View issues.
ag issue view owner/repo 42

# Create an issue.
ag issue create owner/repo --title "Bug report" --body "Description"
```

### Issue Comment

```bash
# Create a comment.
ag issue comment create owner/repo 42 --body "I can reproduce this issue"
ag issue comment create owner/repo 42 --body-file details.md

# View all comments.
ag issue comment view owner/repo 42

# Edit a comment (interactive mode).
ag issue comment edit owner/repo 42 789
ag issue comment edit owner/repo 42 789 --body "Updated information"

# Delete a comment.
ag issue comment delete owner/repo 42 789
ag issue comment delete owner/repo 42 789 --yes
```

### License

```bash
# Check license compliance.
ag license check MIT
ag license check Apache-2.0
ag license check GPL-3.0
```

### SSH Key

```bash
# Add an SSH key.
ag ssh-key add ~/.ssh/id_rsa.pub --title "My Laptop"
cat ~/.ssh/id_rsa.pub | ag ssh-key add --title "My Laptop"
```

## Project Structure

```sh
ag-cli/
├── cmd/ag/main.go              # Entry
├── internal/
│   ├── agcmd/cmd.go            # Core command processing
│   ├── config/config.go        # Configuration management
│   └── api/
│       ├── client.go           # API client
│       └── types.go            # Data type
├── pkg/
│   ├── cmdutil/factory.go      # Command factory
│   └── cmd/
│       ├── root/root.go        # Root command
│       ├── auth/auth.go        # Authentication command
│       ├── repo/               # Repository command
│       │   ├── repo.go
│       │   ├── create.go
│       │   ├── clone.go
│       │   ├── delete.go
│       │   └── fork.go
│       ├── pr/                 # PR command
│       │   ├── pr.go
│       │   └── comment/        # PR comment command
│       ├── issue/              # Issue command
│       │   ├── issue.go
│       │   └── comment/        # Issue comment command
│       ├── license/            # License command
│       │   ├── license.go
│       │   └── check.go
│       └── ssh-key/ssh_key.go  # SSH key command
└── go.mod
```

## API

See AtomGit API v5: `https://api.atomgit.com/api/v5`.

## Reference

- [AtomGit API](https://docs.atomgit.com/docs/apis/)
- [GitHub CLI](https://cli.github.com/)

## Licensing

[Mulan Permissive Software License, Version 2](LICENSE)

Copyright (c) 2026 AtomGit CLI Contributors
