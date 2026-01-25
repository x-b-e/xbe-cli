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

type postRoutersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postRouterDetails struct {
	ID               string   `json:"id"`
	Status           string   `json:"status,omitempty"`
	PostID           string   `json:"post_id,omitempty"`
	PostRouterJobIDs []string `json:"post_router_job_ids,omitempty"`
	Analysis         any      `json:"analysis,omitempty"`
}

func newPostRoutersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post router details",
		Long: `Show the full details of a post router.

Output Fields:
  ID
  Status
  Post ID
  Post Router Job IDs
  Analysis

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The post router ID (required). You can find IDs using the list command.`,
		Example: `  # Show a post router
  xbe view post-routers show 123

  # Get JSON output
  xbe view post-routers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPostRoutersShow,
	}
	initPostRoutersShowFlags(cmd)
	return cmd
}

func init() {
	postRoutersCmd.AddCommand(newPostRoutersShowCmd())
}

func initPostRoutersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostRoutersShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostRoutersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("post router id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-routers]", "status,analysis,post,post-router-jobs")

	body, _, err := client.Get(cmd.Context(), "/v1/post-routers/"+id, query)
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

	details := buildPostRouterDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostRouterDetails(cmd, details)
}

func parsePostRoutersShowOptions(cmd *cobra.Command) (postRoutersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postRoutersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostRouterDetails(resp jsonAPISingleResponse) postRouterDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return postRouterDetails{
		ID:               resource.ID,
		Status:           stringAttr(attrs, "status"),
		PostID:           relationshipIDFromMap(resource.Relationships, "post"),
		PostRouterJobIDs: relationshipIDsFromMap(resource.Relationships, "post-router-jobs"),
		Analysis:         attrs["analysis"],
	}
}

func renderPostRouterDetails(cmd *cobra.Command, details postRouterDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.PostID != "" {
		fmt.Fprintf(out, "Post ID: %s\n", details.PostID)
	}
	if len(details.PostRouterJobIDs) > 0 {
		fmt.Fprintf(out, "Post Router Job IDs: %s\n", strings.Join(details.PostRouterJobIDs, ", "))
	}
	if details.Analysis != nil {
		fmt.Fprintln(out, "Analysis:")
		fmt.Fprintln(out, formatJSONBlock(details.Analysis, "  "))
	}

	return nil
}
