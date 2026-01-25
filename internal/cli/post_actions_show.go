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

type postActionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postActionDetails struct {
	ID        string `json:"id"`
	Token     string `json:"token,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newPostActionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post action details",
		Long: `Show the full details of a post action.

Output Fields:
  ID         Post action identifier
  Token      Post action token
  Created At Creation timestamp
  Updated At Last update timestamp

Arguments:
  <id>    The post action ID (required).`,
		Example: `  # View a post action by ID
  xbe view post-actions show 123

  # Get post action details as JSON
  xbe view post-actions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPostActionsShow,
	}
	initPostActionsShowFlags(cmd)
	return cmd
}

func init() {
	postActionsCmd.AddCommand(newPostActionsShowCmd())
}

func initPostActionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostActionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostActionsShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("post action id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-actions]", "token,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/post-actions/"+id, query)
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

	details := buildPostActionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostActionDetails(cmd, details)
}

func parsePostActionsShowOptions(cmd *cobra.Command) (postActionsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return postActionsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return postActionsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return postActionsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return postActionsShowOptions{}, err
	}

	return postActionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostActionDetails(resp jsonAPISingleResponse) postActionDetails {
	attrs := resp.Data.Attributes
	return postActionDetails{
		ID:        resp.Data.ID,
		Token:     stringAttr(attrs, "token"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}
}

func renderPostActionDetails(cmd *cobra.Command, details postActionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Token != "" {
		fmt.Fprintf(out, "Token: %s\n", details.Token)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
