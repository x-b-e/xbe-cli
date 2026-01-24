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

type doUserCreatorFeedsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	User    string
}

func newDoUserCreatorFeedsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user creator feed",
		Long: `Create a user creator feed.

Optional flags:
  --user  User ID (defaults to the authenticated user, if supported by the API)`,
		Example: `  # Create a user creator feed for the current user
  xbe do user-creator-feeds create

  # Create for a specific user
  xbe do user-creator-feeds create --user 123`,
		Args: cobra.NoArgs,
		RunE: runDoUserCreatorFeedsCreate,
	}
	initDoUserCreatorFeedsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserCreatorFeedsCmd.AddCommand(newDoUserCreatorFeedsCreateCmd())
}

func initDoUserCreatorFeedsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (optional)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserCreatorFeedsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserCreatorFeedsCreateOptions(cmd)
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "user-creator-feeds",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-creator-feeds", jsonBody)
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

	row := buildUserCreatorFeedRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user creator feed %s\n", row.ID)
	return nil
}

func parseDoUserCreatorFeedsCreateOptions(cmd *cobra.Command) (doUserCreatorFeedsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserCreatorFeedsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		User:    user,
	}, nil
}
