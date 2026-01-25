package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type nativeAppReleasesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type nativeAppReleaseDetails struct {
	ID                    string   `json:"id"`
	GitTag                string   `json:"git_tag,omitempty"`
	GitSHA                string   `json:"git_sha,omitempty"`
	BuildNumber           string   `json:"build_number,omitempty"`
	ReleaseChannelDetails any      `json:"release_channel_details,omitempty"`
	Notes                 string   `json:"notes,omitempty"`
	FileAttachmentIDs     []string `json:"file_attachment_ids,omitempty"`
}

func newNativeAppReleasesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show native app release details",
		Long: `Show the full details of a native app release.

Output Fields:
  ID
  Git Tag
  Git SHA
  Build Number
  Notes
  Release Channel Details
  File Attachment IDs

Arguments:
  <id>    The native app release ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a native app release
  xbe view native-app-releases show 123

  # JSON output
  xbe view native-app-releases show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runNativeAppReleasesShow,
	}
	initNativeAppReleasesShowFlags(cmd)
	return cmd
}

func init() {
	nativeAppReleasesCmd.AddCommand(newNativeAppReleasesShowCmd())
}

func initNativeAppReleasesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNativeAppReleasesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseNativeAppReleasesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("native app release id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[native-app-releases]", "git-tag,git-sha,build-number,release-channel-details,notes,file-attachments")
	query.Set("include", "file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/native-app-releases/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildNativeAppReleaseDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderNativeAppReleaseDetails(cmd, details)
}

func parseNativeAppReleasesShowOptions(cmd *cobra.Command) (nativeAppReleasesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return nativeAppReleasesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildNativeAppReleaseDetails(resp jsonAPISingleResponse) nativeAppReleaseDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return nativeAppReleaseDetails{
		ID:                    resource.ID,
		GitTag:                strings.TrimSpace(stringAttr(attrs, "git-tag")),
		GitSHA:                strings.TrimSpace(stringAttr(attrs, "git-sha")),
		BuildNumber:           strings.TrimSpace(stringAttr(attrs, "build-number")),
		ReleaseChannelDetails: attrs["release-channel-details"],
		Notes:                 strings.TrimSpace(stringAttr(attrs, "notes")),
		FileAttachmentIDs:     relationshipIDsFromMap(resource.Relationships, "file-attachments"),
	}
}

func renderNativeAppReleaseDetails(cmd *cobra.Command, details nativeAppReleaseDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.GitTag != "" {
		fmt.Fprintf(out, "Git Tag: %s\n", details.GitTag)
	}
	if details.GitSHA != "" {
		fmt.Fprintf(out, "Git SHA: %s\n", details.GitSHA)
	}
	if details.BuildNumber != "" {
		fmt.Fprintf(out, "Build Number: %s\n", details.BuildNumber)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.ReleaseChannelDetails != nil {
		fmt.Fprintln(out, "\nRelease Channel Details:")
		fmt.Fprintln(out, formatJSONBlock(details.ReleaseChannelDetails, "  "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
