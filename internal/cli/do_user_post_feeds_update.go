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

type doUserPostFeedsUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	EnableVectorIndexing string
}

func newDoUserPostFeedsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user post feed",
		Long: `Update a user post feed.

Optional flags:
  --enable-vector-indexing  Enable vector indexing (true/false)

Notes:
  Vector indexing can only be updated once per hour.
`,
		Example: `  # Enable vector indexing on a user post feed
  xbe do user-post-feeds update 123 --enable-vector-indexing true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUserPostFeedsUpdate,
	}
	initDoUserPostFeedsUpdateFlags(cmd)
	return cmd
}

func init() {
	doUserPostFeedsCmd.AddCommand(newDoUserPostFeedsUpdateCmd())
}

func initDoUserPostFeedsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("enable-vector-indexing", "", "Enable vector indexing (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserPostFeedsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserPostFeedsUpdateOptions(cmd, args)
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
	if opts.EnableVectorIndexing != "" {
		attributes["enable-vector-indexing"] = opts.EnableVectorIndexing == "true"
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "user-post-feeds",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/user-post-feeds/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user post feed %s\n", row.ID)
	return nil
}

func parseDoUserPostFeedsUpdateOptions(cmd *cobra.Command, args []string) (doUserPostFeedsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	enableVectorIndexing, _ := cmd.Flags().GetString("enable-vector-indexing")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserPostFeedsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		EnableVectorIndexing: enableVectorIndexing,
	}, nil
}
