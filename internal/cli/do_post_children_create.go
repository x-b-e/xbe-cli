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

type doPostChildrenCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ParentPost string
	ChildPost  string
}

func newDoPostChildrenCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post child link",
		Long: `Create a post child link.

Required flags:
  --parent-post  Parent post ID (required)
  --child-post   Child post ID (required)

Child posts must be in scope for the parent post creator and cannot be the
same as the parent post.`,
		Example: `  # Link a child post to a parent post
  xbe do post-children create --parent-post 123 --child-post 456

  # Output as JSON
  xbe do post-children create --parent-post 123 --child-post 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPostChildrenCreate,
	}
	initDoPostChildrenCreateFlags(cmd)
	return cmd
}

func init() {
	doPostChildrenCmd.AddCommand(newDoPostChildrenCreateCmd())
}

func initDoPostChildrenCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("parent-post", "", "Parent post ID (required)")
	cmd.Flags().String("child-post", "", "Child post ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPostChildrenCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPostChildrenCreateOptions(cmd)
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

	if opts.ParentPost == "" {
		err := fmt.Errorf("--parent-post is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ChildPost == "" {
		err := fmt.Errorf("--child-post is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"parent-post": map[string]any{
			"data": map[string]any{
				"type": "posts",
				"id":   opts.ParentPost,
			},
		},
		"child-post": map[string]any{
			"data": map[string]any{
				"type": "posts",
				"id":   opts.ChildPost,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "post-children",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/post-children", jsonBody)
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

	row := postChildRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created post child %s\n", row.ID)
	return nil
}

func parseDoPostChildrenCreateOptions(cmd *cobra.Command) (doPostChildrenCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	parentPost, _ := cmd.Flags().GetString("parent-post")
	childPost, _ := cmd.Flags().GetString("child-post")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPostChildrenCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ParentPost: parentPost,
		ChildPost:  childPost,
	}, nil
}
