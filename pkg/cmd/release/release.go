package release

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"atomgit.com/openeuler/ag-cli/internal/api"
	"atomgit.com/openeuler/ag-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdRelease(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Manage releases",
		Long:  `List, create, view, and download releases.`,
	}

	cmd.AddCommand(newCmdReleaseList(f))
	cmd.AddCommand(newCmdReleaseView(f))
	cmd.AddCommand(newCmdReleaseCreate(f))
	cmd.AddCommand(newCmdReleaseDownload(f))

	return cmd
}

func buildReleaseListPath(owner, repo string) string {
	return fmt.Sprintf("/repos/%s/%s/releases", owner, repo)
}

func buildReleaseTagPath(owner, repo, tag string) string {
	return fmt.Sprintf("/repos/%s/%s/releases/%s", owner, repo, tag)
}

func downloadFile(url, token string, w io.Writer) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", res.Status)
	}

	_, err = io.Copy(w, res.Body)
	return err
}

func validateReleaseCreateOptions(tag, name, body string) error {
	if tag == "" {
		return fmt.Errorf("tag is required")
	}
	return nil
}

func parseOwnerRepo(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("repository required")
	}
	parts := strings.Split(args[0], "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
	}
	return parts[0], parts[1], nil
}

func renderReleaseList(releases []api.Release, w io.Writer) {
	for _, r := range releases {
		preRelease := ""
		if r.Prerelease {
			preRelease = " (pre-release)"
		}
		fmt.Fprintf(w, "%s - %s%s\n", r.TagName, r.Name, preRelease)
	}
}

func newCmdReleaseList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <owner/repo>",
		Short: "List releases",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := f.Config.GetToken()
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			owner, repo, err := parseOwnerRepo(args)
			if err != nil {
				return err
			}

			client := api.NewClient(token)

			var releases []api.Release
			if err := client.Get(buildReleaseListPath(owner, repo), &releases); err != nil {
				return err
			}

			if len(releases) == 0 {
				fmt.Println("No releases found")
				return nil
			}

			renderReleaseList(releases, os.Stdout)
			return nil
		},
	}

	return cmd
}

func newCmdReleaseView(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <owner/repo> <tag>",
		Short: "View a release",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := f.Config.GetToken()
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			owner, repo, err := parseOwnerRepo(args)
			if err != nil {
				return err
			}
			tag := args[1]

			client := api.NewClient(token)

			var release api.Release
			if err := client.Get(buildReleaseTagPath(owner, repo, tag), &release); err != nil {
				return err
			}

			fmt.Printf("Tag: %s\n", release.TagName)
			fmt.Printf("Name: %s\n", release.Name)
			fmt.Printf("URL: %s\n", release.HTMLURL)
			fmt.Printf("Created: %s\n", release.CreatedAt)
			if release.Prerelease {
				fmt.Println("Pre-release: yes")
			}
			if release.Body != "" {
				fmt.Printf("\n%s\n", release.Body)
			}
			if len(release.Assets) > 0 {
				fmt.Println("\nAssets:")
				for _, a := range release.Assets {
					fmt.Printf("  - %s (%d downloads)\n", a.Name, a.DownloadCount)
				}
			}

			return nil
		},
	}

	return cmd
}

func newCmdReleaseCreate(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		Tag    string
		Name   string
		Body   string
		Target string
		Draft  bool
		Pre    bool
	}

	cmd := &cobra.Command{
		Use:   "create <owner/repo>",
		Short: "Create a release",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := f.Config.GetToken()
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			owner, repo, err := parseOwnerRepo(args)
			if err != nil {
				return err
			}

			if err := validateReleaseCreateOptions(opts.Tag, opts.Name, opts.Body); err != nil {
				return err
			}

			client := api.NewClient(token)

			body := api.CreateReleaseRequest{
				TagName:         opts.Tag,
				Name:            opts.Name,
				Body:            opts.Body,
				TargetCommitish: opts.Target,
				Draft:           opts.Draft,
				Prerelease:      opts.Pre,
			}

			var release api.Release
			if err := client.Post(buildReleaseListPath(owner, repo), body, &release); err != nil {
				return err
			}

			fmt.Printf("Created release %s\n", release.TagName)
			if release.HTMLURL != "" {
				fmt.Printf("%s\n", release.HTMLURL)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Tag, "tag", "", "Tag name (required)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Release title")
	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Release description")
	cmd.Flags().StringVar(&opts.Target, "target", "", "Target commit SHA or branch name")
	cmd.Flags().BoolVar(&opts.Draft, "draft", false, "Create as draft")
	cmd.Flags().BoolVar(&opts.Pre, "prerelease", false, "Mark as pre-release")

	return cmd
}

func newCmdReleaseDownload(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		Output string
	}

	cmd := &cobra.Command{
		Use:   "download <owner/repo> <tag> [file]",
		Short: "Download release assets",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := f.Config.GetToken()
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			owner, repo, err := parseOwnerRepo(args)
			if err != nil {
				return err
			}
			tag := args[1]

			client := api.NewClient(token)

			var release api.Release
			if err := client.Get(buildReleaseTagPath(owner, repo, tag), &release); err != nil {
				return err
			}

			if len(release.Assets) == 0 {
				fmt.Println("No assets to download")
				return nil
			}

			fileName := ""
			if len(args) == 3 {
				fileName = args[2]
			} else if len(release.Assets) == 1 {
				fileName = release.Assets[0].Name
			} else {
				fmt.Println("Multiple assets found, specify one:")
				for _, a := range release.Assets {
					fmt.Printf("  - %s\n", a.Name)
				}
				return fmt.Errorf("file name required")
			}

			for _, a := range release.Assets {
				if a.Name != fileName {
					continue
				}
				outPath := opts.Output
				if outPath == "" {
					outPath = a.Name
				}
				out, err := os.Create(outPath)
				if err != nil {
					return err
				}
				defer out.Close()
				if err := downloadFile(a.BrowserDownloadURL, token, out); err != nil {
					return err
				}
				fmt.Printf("Downloaded to %s\n", outPath)
				return nil
			}

			return fmt.Errorf("asset %q not found in release %s", fileName, tag)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path")

	return cmd
}
