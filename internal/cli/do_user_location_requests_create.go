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

type doUserLocationRequestsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	UserID  string
}

func newDoUserLocationRequestsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user location request",
		Long: `Create a user location request.

Required flags:
  --user    User ID to request location from

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Request a user's location
  xbe do user-location-requests create --user 123

  # Output as JSON
  xbe do user-location-requests create --user 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoUserLocationRequestsCreate,
	}
	initDoUserLocationRequestsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserLocationRequestsCmd.AddCommand(newDoUserLocationRequestsCreateCmd())
}

func initDoUserLocationRequestsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID to request location from (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("user")
}

func runDoUserLocationRequestsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUserLocationRequestsCreateOptions(cmd)
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

	userID := strings.TrimSpace(opts.UserID)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-location-requests",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-location-requests", jsonBody)
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

	details := buildUserLocationRequestDetails(resp)
	if details.UserID == "" {
		details.UserID = userID
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user location request %s\n", details.ID)
	return nil
}

func parseDoUserLocationRequestsCreateOptions(cmd *cobra.Command) (doUserLocationRequestsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserLocationRequestsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		UserID:  userID,
	}, nil
}
