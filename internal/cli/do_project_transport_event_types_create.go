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

type doProjectTransportEventTypesCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	Code                   string
	Name                   string
	DwellMinutesMinDefault string
	TransportOrderStopRole string
	Broker                 string
}

func newDoProjectTransportEventTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project transport event type",
		Long: `Create a new project transport event type.

Required flags:
  --name    The event type name (required)
  --broker  The broker ID (required)

Optional flags:
  --code                      Event type code
  --dwell-minutes-min-default Default minimum dwell minutes
  --transport-order-stop-role Transport order stop role`,
		Example: `  # Create a project transport event type
  xbe do project-transport-event-types create --name "Pickup" --broker 123

  # Create with all options
  xbe do project-transport-event-types create --name "Pickup" --code "PU" --dwell-minutes-min-default 15 --broker 123`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportEventTypesCreate,
	}
	initDoProjectTransportEventTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportEventTypesCmd.AddCommand(newDoProjectTransportEventTypesCreateCmd())
}

func initDoProjectTransportEventTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Event type code")
	cmd.Flags().String("name", "", "Event type name (required)")
	cmd.Flags().String("dwell-minutes-min-default", "", "Default minimum dwell minutes")
	cmd.Flags().String("transport-order-stop-role", "", "Transport order stop role")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportEventTypesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportEventTypesCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.Code != "" {
		attributes["code"] = opts.Code
	}
	if opts.DwellMinutesMinDefault != "" {
		attributes["dwell-minutes-min-default"] = opts.DwellMinutesMinDefault
	}
	if opts.TransportOrderStopRole != "" {
		attributes["transport-order-stop-role"] = opts.TransportOrderStopRole
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-event-types",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-event-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport event type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectTransportEventTypesCreateOptions(cmd *cobra.Command) (doProjectTransportEventTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	name, _ := cmd.Flags().GetString("name")
	dwellMinutesMinDefault, _ := cmd.Flags().GetString("dwell-minutes-min-default")
	transportOrderStopRole, _ := cmd.Flags().GetString("transport-order-stop-role")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportEventTypesCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		Code:                   code,
		Name:                   name,
		DwellMinutesMinDefault: dwellMinutesMinDefault,
		TransportOrderStopRole: transportOrderStopRole,
		Broker:                 broker,
	}, nil
}

func buildProjectTransportEventTypeRowFromSingle(resp jsonAPISingleResponse) projectTransportEventTypeRow {
	attrs := resp.Data.Attributes

	row := projectTransportEventTypeRow{
		ID:                     resp.Data.ID,
		Code:                   stringAttr(attrs, "code"),
		Name:                   stringAttr(attrs, "name"),
		DwellMinutesMinDefault: attrs["dwell-minutes-min-default"],
		TransportOrderStopRole: stringAttr(attrs, "transport-order-stop-role"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
