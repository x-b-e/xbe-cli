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

type postRouterJobsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postRouterJobDetails struct {
	ID                  string `json:"id"`
	PostRouterID        string `json:"post_router_id,omitempty"`
	PostID              string `json:"post_id,omitempty"`
	PostWorkerClassName string `json:"post_worker_class_name,omitempty"`
	PostWorkerJID       string `json:"post_worker_jid,omitempty"`
	IsNotWorthPosting   bool   `json:"is_not_worth_posting,omitempty"`
}

func newPostRouterJobsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post router job details",
		Long: `Show the full details of a post router job.

Output Fields:
  ID
  Post Router ID
  Post ID
  Post Worker Class Name
  Post Worker JID
  Is Not Worth Posting

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The post router job ID (required). You can find IDs using the list command.`,
		Example: `  # Show a post router job
  xbe view post-router-jobs show 123

  # Get JSON output
  xbe view post-router-jobs show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPostRouterJobsShow,
	}
	initPostRouterJobsShowFlags(cmd)
	return cmd
}

func init() {
	postRouterJobsCmd.AddCommand(newPostRouterJobsShowCmd())
}

func initPostRouterJobsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostRouterJobsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostRouterJobsShowOptions(cmd)
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
		return fmt.Errorf("post router job id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-router-jobs]", "post-worker-class-name,post-worker-jid,is-not-worth-posting,post-router,post")

	body, _, err := client.Get(cmd.Context(), "/v1/post-router-jobs/"+id, query)
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

	details := buildPostRouterJobDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostRouterJobDetails(cmd, details)
}

func parsePostRouterJobsShowOptions(cmd *cobra.Command) (postRouterJobsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postRouterJobsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostRouterJobDetails(resp jsonAPISingleResponse) postRouterJobDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := postRouterJobDetails{
		ID:                  resource.ID,
		PostWorkerClassName: stringAttr(attrs, "post-worker-class-name"),
		PostWorkerJID:       stringAttr(attrs, "post-worker-jid"),
		IsNotWorthPosting:   boolAttr(attrs, "is-not-worth-posting"),
		PostRouterID:        relationshipIDFromMap(resource.Relationships, "post-router"),
		PostID:              relationshipIDFromMap(resource.Relationships, "post"),
	}

	return details
}

func renderPostRouterJobDetails(cmd *cobra.Command, details postRouterJobDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PostRouterID != "" {
		fmt.Fprintf(out, "Post Router ID: %s\n", details.PostRouterID)
	}
	if details.PostID != "" {
		fmt.Fprintf(out, "Post ID: %s\n", details.PostID)
	}
	if details.PostWorkerClassName != "" {
		fmt.Fprintf(out, "Post Worker Class Name: %s\n", details.PostWorkerClassName)
	}
	if details.PostWorkerJID != "" {
		fmt.Fprintf(out, "Post Worker JID: %s\n", details.PostWorkerJID)
	}
	fmt.Fprintf(out, "Is Not Worth Posting: %s\n", formatBool(details.IsNotWorthPosting))

	return nil
}
