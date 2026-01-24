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

type doBrokerProjectTransportEventTypesUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Code    string
}

func newDoBrokerProjectTransportEventTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker project transport event type",
		Long: `Update a broker project transport event type.

Provide the broker project transport event type ID as an argument, then use flags
to specify which fields to update. Only specified fields will be modified.

Updatable fields:
  --code    Broker-specific event type code`,
		Example: `  # Update the broker-specific code
  xbe do broker-project-transport-event-types update 123 --code "DROP"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerProjectTransportEventTypesUpdate,
	}
	initDoBrokerProjectTransportEventTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerProjectTransportEventTypesCmd.AddCommand(newDoBrokerProjectTransportEventTypesUpdateCmd())
}

func initDoBrokerProjectTransportEventTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Broker-specific event type code")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerProjectTransportEventTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerProjectTransportEventTypesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("code") {
		attributes["code"] = opts.Code
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --code")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "broker-project-transport-event-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-project-transport-event-types/"+opts.ID, jsonBody)
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

	row := buildBrokerProjectTransportEventTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker project transport event type %s (%s)\n", row.ID, row.Code)
	return nil
}

func parseDoBrokerProjectTransportEventTypesUpdateOptions(cmd *cobra.Command, args []string) (doBrokerProjectTransportEventTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerProjectTransportEventTypesUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Code:    code,
	}, nil
}
