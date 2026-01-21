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

type doProjectTransportEventTypesUpdateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	ID                     string
	Code                   string
	Name                   string
	DwellMinutesMinDefault string
	TransportOrderStopRole string
}

func newDoProjectTransportEventTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project transport event type",
		Long: `Update an existing project transport event type.

Provide the event type ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --code                      Event type code
  --name                      Event type name
  --dwell-minutes-min-default Default minimum dwell minutes
  --transport-order-stop-role Transport order stop role`,
		Example: `  # Update name
  xbe do project-transport-event-types update 123 --name "Updated Name"

  # Update multiple fields
  xbe do project-transport-event-types update 123 --name "New Name" --code "NEW"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportEventTypesUpdate,
	}
	initDoProjectTransportEventTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportEventTypesCmd.AddCommand(newDoProjectTransportEventTypesUpdateCmd())
}

func initDoProjectTransportEventTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Event type code")
	cmd.Flags().String("name", "", "Event type name")
	cmd.Flags().String("dwell-minutes-min-default", "", "Default minimum dwell minutes")
	cmd.Flags().String("transport-order-stop-role", "", "Transport order stop role")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportEventTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportEventTypesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("dwell-minutes-min-default") {
		attributes["dwell-minutes-min-default"] = opts.DwellMinutesMinDefault
	}
	if cmd.Flags().Changed("transport-order-stop-role") {
		attributes["transport-order-stop-role"] = opts.TransportOrderStopRole
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --code, --name, --dwell-minutes-min-default, --transport-order-stop-role")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-transport-event-types",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-event-types/"+opts.ID, jsonBody)
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

	row := buildProjectTransportEventTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport event type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectTransportEventTypesUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportEventTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	name, _ := cmd.Flags().GetString("name")
	dwellMinutesMinDefault, _ := cmd.Flags().GetString("dwell-minutes-min-default")
	transportOrderStopRole, _ := cmd.Flags().GetString("transport-order-stop-role")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportEventTypesUpdateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		ID:                     args[0],
		Code:                   code,
		Name:                   name,
		DwellMinutesMinDefault: dwellMinutesMinDefault,
		TransportOrderStopRole: transportOrderStopRole,
	}, nil
}
