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

type doUserAuthTokenResetsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	UserID  string
}

type userAuthTokenResetRow struct {
	ID      string `json:"id"`
	IsReset bool   `json:"is_reset"`
}

func newDoUserAuthTokenResetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reset a user's auth token",
		Long: `Reset a user's auth token.

Required flags:
  --user-id    User ID to reset (required)`,
		Example: `  # Reset a user's auth token
  xbe do user-auth-token-resets create --user-id 123

  # Output as JSON
  xbe do user-auth-token-resets create --user-id 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoUserAuthTokenResetsCreate,
	}
	initDoUserAuthTokenResetsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserAuthTokenResetsCmd.AddCommand(newDoUserAuthTokenResetsCreateCmd())
}

func initDoUserAuthTokenResetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user-id", "", "User ID to reset (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserAuthTokenResetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserAuthTokenResetsCreateOptions(cmd)
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

	opts.UserID = strings.TrimSpace(opts.UserID)
	if opts.UserID == "" {
		err := fmt.Errorf("--user-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "user-auth-token-resets",
			"attributes": map[string]any{
				"user-id": opts.UserID,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-auth-token-resets", jsonBody)
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

	row := userAuthTokenResetRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.IsReset {
		fmt.Fprintf(cmd.OutOrStdout(), "Reset auth token for user %s\n", row.ID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user auth token reset %s\n", row.ID)
	return nil
}

func userAuthTokenResetRowFromSingle(resp jsonAPISingleResponse) userAuthTokenResetRow {
	attrs := resp.Data.Attributes
	isReset := boolAttr(attrs, "is-reset")
	if !isReset {
		isReset = boolAttr(attrs, "is_reset")
	}
	return userAuthTokenResetRow{
		ID:      resp.Data.ID,
		IsReset: isReset,
	}
}

func parseDoUserAuthTokenResetsCreateOptions(cmd *cobra.Command) (doUserAuthTokenResetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserAuthTokenResetsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		UserID:  userID,
	}, nil
}
