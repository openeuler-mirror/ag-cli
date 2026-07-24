package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"atomgit.com/openeuler/ag-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SyncOptions struct {
	Branch         string
	Remote         string
	UpstreamRemote string
	Strategy       string
	Push           bool
	DryRun         bool
	AllowDirty     bool
	SetUpstream    bool
}

type commandRunner interface {
	Run(name string, args ...string) (string, error)
	RunInteractive(name string, args ...string) error
}

type shellRunner struct{}

func (shellRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (shellRunner) RunInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func newCmdRepoSync(_ *cmdutil.Factory) *cobra.Command {
	opts := &SyncOptions{
		Remote:         "origin",
		UpstreamRemote: "upstream",
		Strategy:       "ff-only",
	}

	cmd := &cobra.Command{
		Use:   "sync [<upstream-owner>/]<repo>",
		Short: "Sync a fork branch with its upstream repository",
		Long: `Sync the current fork branch with an upstream AtomGit repository.

The command is intended for contributors who work from forks. It can add or
update the upstream remote, fetch the selected branch, apply upstream changes
with a safe strategy, and optionally push the synchronized branch to the fork.`,
		Example: `  # Preview updates from upstream/master
  ag repo sync openeuler/ag-cli --dry-run

  # Fast-forward the current branch from upstream/master
  ag repo sync openeuler/ag-cli

  # Sync develop with rebase and push to origin
  ag repo sync openeuler/ag-cli --branch develop --strategy rebase --push

  # Use an existing remote named source
  ag repo sync --upstream-remote source --branch master`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			upstreamRepo := ""
			if len(args) > 0 {
				upstreamRepo = args[0]
			}
			return runSync(shellRunner{}, opts, upstreamRepo)
		},
	}

	cmd.Flags().StringVarP(&opts.Branch, "branch", "b", "", "Branch to sync; defaults to the current branch")
	cmd.Flags().StringVar(&opts.Remote, "remote", "origin", "Fork remote to push to")
	cmd.Flags().StringVar(&opts.UpstreamRemote, "upstream-remote", "upstream", "Remote name for the upstream repository")
	cmd.Flags().StringVar(&opts.Strategy, "strategy", "ff-only", "Apply strategy: ff-only, merge, or rebase")
	cmd.Flags().BoolVar(&opts.Push, "push", false, "Push the synchronized branch to the fork remote")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show the sync plan without changing the worktree")
	cmd.Flags().BoolVar(&opts.AllowDirty, "allow-dirty", false, "Allow sync when the worktree has uncommitted changes")
	cmd.Flags().BoolVar(&opts.SetUpstream, "set-upstream", false, "Set the fork branch upstream when pushing")

	return cmd
}

func runSync(r commandRunner, opts *SyncOptions, upstreamRepo string) error {
	if err := validateSyncOptions(opts); err != nil {
		return err
	}

	if _, err := r.Run("git", "rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("not a git repository")
	}

	branch := opts.Branch
	if branch == "" {
		current, err := r.Run("git", "branch", "--show-current")
		if err != nil {
			return fmt.Errorf("failed to detect current branch: %w", err)
		}
		if current == "" {
			return fmt.Errorf("cannot sync from detached HEAD; pass --branch explicitly")
		}
		branch = current
	}

	if !opts.AllowDirty {
		dirty, err := r.Run("git", "status", "--porcelain")
		if err != nil {
			return fmt.Errorf("failed to inspect worktree: %w", err)
		}
		if dirty != "" {
			return fmt.Errorf("worktree has uncommitted changes; commit, stash, or pass --allow-dirty")
		}
	}

	upstreamURL := ""
	if upstreamRepo != "" {
		var err error
		upstreamURL, err = buildAtomGitCloneURL(upstreamRepo)
		if err != nil {
			return err
		}
	}

	if err := ensureUpstreamRemote(r, opts.UpstreamRemote, upstreamURL, opts.DryRun); err != nil {
		return err
	}

	source := fmt.Sprintf("%s/%s", opts.UpstreamRemote, branch)
	fmt.Printf("Repository sync plan\n")
	fmt.Printf("  upstream: %s\n", opts.UpstreamRemote)
	if upstreamURL != "" {
		fmt.Printf("  upstream URL: %s\n", upstreamURL)
	}
	fmt.Printf("  branch: %s\n", branch)
	fmt.Printf("  strategy: %s\n", opts.Strategy)
	fmt.Printf("  push: %v\n", opts.Push)

	if opts.DryRun {
		fmt.Println("\nCommands:")
		if upstreamURL != "" {
			fmt.Printf("  git remote add %s %s  # if remote does not exist\n", opts.UpstreamRemote, upstreamURL)
			fmt.Printf("  git remote set-url %s %s  # if remote exists with a different URL\n", opts.UpstreamRemote, upstreamURL)
		}
		fmt.Printf("  git fetch %s %s\n", opts.UpstreamRemote, branch)
		fmt.Printf("  git %s\n", syncStrategyCommand(opts.Strategy, source))
		if opts.Push {
			fmt.Printf("  git %s\n", pushCommand(opts.Remote, branch, opts.SetUpstream))
		}
		printIncomingCommits(r, source)
		return nil
	}

	fmt.Printf("\nFetching %s %s...\n", opts.UpstreamRemote, branch)
	if err := r.RunInteractive("git", "fetch", opts.UpstreamRemote, branch); err != nil {
		return err
	}

	fmt.Printf("Applying %s with %s...\n", source, opts.Strategy)
	if err := runSyncStrategy(r, opts.Strategy, source); err != nil {
		return err
	}

	if opts.Push {
		fmt.Printf("Pushing %s to %s...\n", branch, opts.Remote)
		if err := runPush(r, opts.Remote, branch, opts.SetUpstream); err != nil {
			return err
		}
	}

	fmt.Printf("Synced %s from %s\n", branch, source)
	return nil
}

func validateSyncOptions(opts *SyncOptions) error {
	if opts.Remote == "" {
		return fmt.Errorf("remote cannot be empty")
	}
	if opts.UpstreamRemote == "" {
		return fmt.Errorf("upstream remote cannot be empty")
	}
	switch opts.Strategy {
	case "ff-only", "merge", "rebase":
		return nil
	default:
		return fmt.Errorf("unsupported sync strategy %q (expected ff-only, merge, or rebase)", opts.Strategy)
	}
}

func buildAtomGitCloneURL(repo string) (string, error) {
	if strings.HasPrefix(repo, "http://") || strings.HasPrefix(repo, "https://") || strings.HasPrefix(repo, "git@") {
		if strings.HasSuffix(repo, ".git") {
			return repo, nil
		}
		return repo + ".git", nil
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid upstream repository format: %s (expected owner/repo)", repo)
	}
	return fmt.Sprintf("https://atomgit.com/%s/%s.git", parts[0], parts[1]), nil
}

func ensureUpstreamRemote(r commandRunner, remote, url string, dryRun bool) error {
	currentURL, err := r.Run("git", "remote", "get-url", remote)
	if err != nil {
		if url == "" {
			return fmt.Errorf("remote %q does not exist; pass upstream repository as owner/repo", remote)
		}
		if dryRun {
			return nil
		}
		fmt.Printf("Adding remote %s -> %s\n", remote, url)
		return r.RunInteractive("git", "remote", "add", remote, url)
	}

	if url == "" || currentURL == url {
		return nil
	}
	if dryRun {
		return nil
	}

	fmt.Printf("Updating remote %s -> %s\n", remote, url)
	return r.RunInteractive("git", "remote", "set-url", remote, url)
}

func syncStrategyCommand(strategy, source string) string {
	switch strategy {
	case "merge":
		return fmt.Sprintf("merge --no-edit %s", source)
	case "rebase":
		return fmt.Sprintf("rebase %s", source)
	default:
		return fmt.Sprintf("merge --ff-only %s", source)
	}
}

func runSyncStrategy(r commandRunner, strategy, source string) error {
	switch strategy {
	case "merge":
		return r.RunInteractive("git", "merge", "--no-edit", source)
	case "rebase":
		return r.RunInteractive("git", "rebase", source)
	default:
		return r.RunInteractive("git", "merge", "--ff-only", source)
	}
}

func pushCommand(remote, branch string, setUpstream bool) string {
	if setUpstream {
		return fmt.Sprintf("push --set-upstream %s %s", remote, branch)
	}
	return fmt.Sprintf("push %s %s", remote, branch)
}

func runPush(r commandRunner, remote, branch string, setUpstream bool) error {
	if setUpstream {
		return r.RunInteractive("git", "push", "--set-upstream", remote, branch)
	}
	return r.RunInteractive("git", "push", remote, branch)
}

func printIncomingCommits(r commandRunner, source string) {
	log, err := r.Run("git", "log", "--oneline", "--decorate", "--max-count=20", fmt.Sprintf("HEAD..%s", source))
	if err != nil || log == "" {
		return
	}
	fmt.Println("\nIncoming commits:")
	for _, line := range strings.Split(log, "\n") {
		fmt.Printf("  %s\n", line)
	}
}
