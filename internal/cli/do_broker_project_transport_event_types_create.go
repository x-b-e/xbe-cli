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

type doBrokerProjectTransportEventTypesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	Code                      string
	Broker                    string
	ProjectTransportEventType string
}

func newDoBrokerProjectTransportEventTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker project transport event type",
		Long: `Create a broker project transport event type.

Required flags:
  --code                       Broker-specific event type code (required)
  --broker                     Broker ID (required)
  --project-transport-event-type Project transport event type ID (required)`,
		Example: `  # Create a broker project transport event type
  xbe do broker-project-transport-event-types create \\
    --broker 123 \\
    --project-transport-event-type 456 \\
    --code "PICK"

  # JSON output
  xbe do broker-project-transport-event-types create \\
    --broker 123 \\
    --project-transport-event-type 456 \\
    --code "PICK" \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerProjectTransportEventTypesCreate,
	}
	initDoBrokerProjectTransportEventTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerProjectTransportEventTypesCmd.AddCommand(newDoBrokerProjectTransportEventTypesCreateCmd())
}

func initDoBrokerProjectTransportEventTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("code", "", "Broker-specific event type code (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("project-transport-event-type", "", "Project transport event type ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerProjectTransportEventTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerProjectTransportEventTypesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Code) == "" {
		err := fmt.Errorf("--code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Broker) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ProjectTransportEventType) == "" {
		err := fmt.Errorf("--project-transport-event-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"code": opts.Code,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"project-transport-event-type": map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.ProjectTransportEventType,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-project-transport-event-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-project-transport-event-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker project transport event type %s (%s)\n", row.ID, row.Code)
	return nil
}

func parseDoBrokerProjectTransportEventTypesCreateOptions(cmd *cobra.Command) (doBrokerProjectTransportEventTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	code, _ := cmd.Flags().GetString("code")
	broker, _ := cmd.Flags().GetString("broker")
	eventType, _ := cmd.Flags().GetString("project-transport-event-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerProjectTransportEventTypesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		Code:                      code,
		Broker:                    broker,
		ProjectTransportEventType: eventType,
	}, nil
}

func buildBrokerProjectTransportEventTypeRowFromSingle(resp jsonAPISingleResponse) brokerProjectTransportEventTypeRow {
	row := brokerProjectTransportEventTypeRow{
		ID:   resp.Data.ID,
		Code: stringAttr(resp.Data.Attributes, "code"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		row.ProjectTransportEventTypeID = rel.Data.ID
	}

	return row
}
