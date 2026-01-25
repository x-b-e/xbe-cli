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

type doUserPostFeedsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	User                 string
	EnableVectorIndexing string
}

func newDoUserPostFeedsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user post feed",
		Long: `Create a user post feed.

Optional flags:
  --user                    User ID (defaults to the authenticated user, if supported by the API)
  --enable-vector-indexing  Enable vector indexing (true/false)`,
		Example: `  # Create a user post feed for the current user
  xbe do user-post-feeds create

  # Create for a specific user
  xbe do user-post-feeds create --user 123

  # Create with vector indexing enabled
  xbe do user-post-feeds create --enable-vector-indexing true`,
		Args: cobra.NoArgs,
		RunE: runDoUserPostFeedsCreate,
	}
	initDoUserPostFeedsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserPostFeedsCmd.AddCommand(newDoUserPostFeedsCreateCmd())
}

func initDoUserPostFeedsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (optional)")
	cmd.Flags().String("enable-vector-indexing", "", "Enable vector indexing (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserPostFeedsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserPostFeedsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.User) != "" {
		attributes["user-id"] = opts.User
	}
	if opts.EnableVectorIndexing != "" {
		attributes["enable-vector-indexing"] = opts.EnableVectorIndexing == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "user-post-feeds",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-post-feeds", jsonBody)
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

	row := buildUserPostFeedRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user post feed %s\n", row.ID)
	return nil
}

func parseDoUserPostFeedsCreateOptions(cmd *cobra.Command) (doUserPostFeedsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	enableVectorIndexing, _ := cmd.Flags().GetString("enable-vector-indexing")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserPostFeedsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		User:                 user,
		EnableVectorIndexing: enableVectorIndexing,
	}, nil
}
