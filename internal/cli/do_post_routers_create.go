package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPostRoutersCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	PostID  string
}

func newDoPostRoutersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post router",
		Long: `Create a post router for a post.

Post routers analyze the post and enqueue routing jobs.

Required flags:
  --post   Post ID to route

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a post router for a post
  xbe do post-routers create --post 123

  # JSON output
  xbe do post-routers create --post 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPostRoutersCreate,
	}
	initDoPostRoutersCreateFlags(cmd)
	return cmd
}

func init() {
	doPostRoutersCmd.AddCommand(newDoPostRoutersCreateCmd())
}

func initDoPostRoutersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("post", "", "Post ID to route (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("post")
}

func runDoPostRoutersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPostRoutersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	postID := strings.TrimSpace(opts.PostID)
	if postID == "" {
		err := fmt.Errorf("--post is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"post": map[string]any{
			"data": map[string]any{
				"type": "posts",
				"id":   postID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "post-routers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/post-routers", jsonBody)
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

	row := buildPostRouterDetails(resp)
	if row.PostID == "" {
		row.PostID = postID
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderPostRouterCreate(cmd, row)
}

func parseDoPostRoutersCreateOptions(cmd *cobra.Command) (doPostRoutersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	postID, _ := cmd.Flags().GetString("post")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPostRoutersCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		PostID:  postID,
	}, nil
}

func renderPostRouterCreate(cmd *cobra.Command, details postRouterDetails) error {
	out := cmd.OutOrStdout()

	if details.ID != "" {
		fmt.Fprintf(out, "Created post router %s\n", details.ID)
	} else {
		fmt.Fprintln(out, "Created post router")
	}

	if details.PostID != "" {
		fmt.Fprintf(out, "Post ID: %s\n", details.PostID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
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
