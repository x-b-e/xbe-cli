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

type doPostViewsCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Post     string
	Viewer   string
	ViewedAt string
}

func newDoPostViewsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Record a post view",
		Long: `Record a post view for a user and post.

Required flags:
  --post       Post ID (required)
  --viewer     Viewer user ID (required)
  --viewed-at  View timestamp (ISO 8601, required)

Note: The viewer must match the authenticated user and must be allowed to view the post.`,
		Example: `  # Record a post view
  xbe do post-views create --post 123 --viewer 456 --viewed-at 2025-01-01T12:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoPostViewsCreate,
	}
	initDoPostViewsCreateFlags(cmd)
	return cmd
}

func init() {
	doPostViewsCmd.AddCommand(newDoPostViewsCreateCmd())
}

func initDoPostViewsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("post", "", "Post ID (required)")
	cmd.Flags().String("viewer", "", "Viewer user ID (required)")
	cmd.Flags().String("viewed-at", "", "View timestamp (ISO 8601, required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPostViewsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPostViewsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Post) == "" {
		err := fmt.Errorf("--post is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Viewer) == "" {
		err := fmt.Errorf("--viewer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ViewedAt) == "" {
		err := fmt.Errorf("--viewed-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"post": map[string]any{
			"data": map[string]any{
				"type": "posts",
				"id":   opts.Post,
			},
		},
		"viewer": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Viewer,
			},
		},
	}

	attributes := map[string]any{
		"viewed-at": opts.ViewedAt,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "post-views",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/post-views", jsonBody)
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

	row := postViewRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created post view %s\n", row.ID)
	return nil
}

func parseDoPostViewsCreateOptions(cmd *cobra.Command) (doPostViewsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	post, _ := cmd.Flags().GetString("post")
	viewer, _ := cmd.Flags().GetString("viewer")
	viewedAt, _ := cmd.Flags().GetString("viewed-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPostViewsCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Post:     post,
		Viewer:   viewer,
		ViewedAt: viewedAt,
	}, nil
}
