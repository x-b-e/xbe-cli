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

type doKeepTruckinUsersUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	UserID  string
}

func newDoKeepTruckinUsersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a KeepTruckin user assignment",
		Long: `Update an existing KeepTruckin user assignment.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Arguments:
  <id>    The KeepTruckin user ID (required)

Flags:
  --user   User ID to link (use empty string to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Assign a KeepTruckin user to an existing user
  xbe do keep-truckin-users update 123 --user 456

  # Clear the linked user
  xbe do keep-truckin-users update 123 --user ""

  # Get JSON output
  xbe do keep-truckin-users update 123 --user 456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoKeepTruckinUsersUpdate,
	}
	initDoKeepTruckinUsersUpdateFlags(cmd)
	return cmd
}

func init() {
	doKeepTruckinUsersCmd.AddCommand(newDoKeepTruckinUsersUpdateCmd())
}

func initDoKeepTruckinUsersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID to link (use empty string to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoKeepTruckinUsersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoKeepTruckinUsersUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("user") {
		if strings.TrimSpace(opts.UserID) == "" {
			relationships["user"] = map[string]any{"data": nil}
		} else {
			relationships["user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.UserID,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "keep-truckin-users",
		"id":            opts.ID,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)
	body, _, err := client.Patch(cmd.Context(), "/v1/keep-truckin-users/"+opts.ID, jsonBody)
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

	details := buildKeepTruckinUserDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated KeepTruckin user %s\n", details.ID)
	return nil
}

func parseDoKeepTruckinUsersUpdateOptions(cmd *cobra.Command, args []string) (doKeepTruckinUsersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doKeepTruckinUsersUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		UserID:  userID,
	}, nil
}
