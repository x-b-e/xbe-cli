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

type doChangeRequestsUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Requests string
}

func newDoChangeRequestsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a change request",
		Long: `Update a change request.

Optional flags:
  --requests   Request items as JSON array (use [] to clear)

Arguments:
  <id>         The change request ID (required).`,
		Example: `  # Update requests
  xbe do change-requests update 123 \
    --requests '[{"field":"status","from":"approved","to":"rejected"}]'

  # Output as JSON
  xbe do change-requests update 123 \
    --requests '[{"field":"status","from":"approved","to":"rejected"}]' \
    --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoChangeRequestsUpdate,
	}
	initDoChangeRequestsUpdateFlags(cmd)
	return cmd
}

func init() {
	doChangeRequestsCmd.AddCommand(newDoChangeRequestsUpdateCmd())
}

func initDoChangeRequestsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("requests", "", "Request items as JSON array (use [] to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoChangeRequestsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoChangeRequestsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("requests") {
		requests, err := parseChangeRequestRequests(opts.Requests)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["requests"] = requests
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "change-requests",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/change-requests/"+opts.ID, jsonBody)
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

	row := buildChangeRequestRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated change request %s\n", row.ID)
	return nil
}

func parseDoChangeRequestsUpdateOptions(cmd *cobra.Command, args []string) (doChangeRequestsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	requests, _ := cmd.Flags().GetString("requests")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doChangeRequestsUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Requests: requests,
	}, nil
}
