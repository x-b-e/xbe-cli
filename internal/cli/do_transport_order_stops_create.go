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

type doTransportOrderStopsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	TransportOrderID      string
	LocationID            string
	Role                  string
	Status                string
	Position              int
	AtMin                 string
	AtMax                 string
	ExternalTmsStopNumber string
}

func newDoTransportOrderStopsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport order stop",
		Long: `Create a transport order stop.

Required flags:
  --transport-order   Transport order ID (required)
  --location          Transport location ID (required)
  --role              Stop role (pickup, delivery) (required)

Optional flags:
  --status                 Stop status (planned, started, finished, cancelled)
  --position               Stop position in the order
  --at-min                 Earliest scheduled time (ISO 8601)
  --at-max                 Latest scheduled time (ISO 8601)
  --external-tms-stop-number External TMS stop number`,
		Example: `  # Create a pickup stop
  xbe do transport-order-stops create \
    --transport-order 123 \
    --location 456 \
    --role pickup \
    --at-min 2026-01-23T08:00:00Z \
    --at-max 2026-01-23T10:00:00Z

  # Create with position and external TMS stop number
  xbe do transport-order-stops create \
    --transport-order 123 \
    --location 456 \
    --role delivery \
    --position 2 \
    --external-tms-stop-number TMS-STOP-200`,
		Args: cobra.NoArgs,
		RunE: runDoTransportOrderStopsCreate,
	}
	initDoTransportOrderStopsCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderStopsCmd.AddCommand(newDoTransportOrderStopsCreateCmd())
}

func initDoTransportOrderStopsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("transport-order", "", "Transport order ID (required)")
	cmd.Flags().String("location", "", "Transport location ID (required)")
	cmd.Flags().String("role", "", "Stop role (pickup, delivery) (required)")
	cmd.Flags().String("status", "", "Stop status (planned, started, finished, cancelled)")
	cmd.Flags().Int("position", 0, "Stop position in the order")
	cmd.Flags().String("at-min", "", "Earliest scheduled time (ISO 8601)")
	cmd.Flags().String("at-max", "", "Latest scheduled time (ISO 8601)")
	cmd.Flags().String("external-tms-stop-number", "", "External TMS stop number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderStopsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTransportOrderStopsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TransportOrderID) == "" {
		err := fmt.Errorf("--transport-order is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.LocationID) == "" {
		err := fmt.Errorf("--location is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Role) == "" {
		err := fmt.Errorf("--role is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"role": opts.Role,
	}

	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}
	if opts.AtMin != "" {
		attributes["at-min"] = opts.AtMin
	}
	if opts.AtMax != "" {
		attributes["at-max"] = opts.AtMax
	}
	if opts.ExternalTmsStopNumber != "" {
		attributes["external-tms-stop-number"] = opts.ExternalTmsStopNumber
	}

	relationships := map[string]any{
		"transport-order": map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrderID,
			},
		},
		"location": map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.LocationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "transport-order-stops",
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

	body, _, err := client.Post(cmd.Context(), "/v1/transport-order-stops", jsonBody)
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

	row := buildTransportOrderStopRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created transport order stop %s\n", row.ID)
	return nil
}

func parseDoTransportOrderStopsCreateOptions(cmd *cobra.Command) (doTransportOrderStopsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	transportOrderID, _ := cmd.Flags().GetString("transport-order")
	locationID, _ := cmd.Flags().GetString("location")
	role, _ := cmd.Flags().GetString("role")
	status, _ := cmd.Flags().GetString("status")
	position, _ := cmd.Flags().GetInt("position")
	atMin, _ := cmd.Flags().GetString("at-min")
	atMax, _ := cmd.Flags().GetString("at-max")
	externalTmsStopNumber, _ := cmd.Flags().GetString("external-tms-stop-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderStopsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		TransportOrderID:      transportOrderID,
		LocationID:            locationID,
		Role:                  role,
		Status:                status,
		Position:              position,
		AtMin:                 atMin,
		AtMax:                 atMax,
		ExternalTmsStopNumber: externalTmsStopNumber,
	}, nil
}

func buildTransportOrderStopRowFromSingle(resp jsonAPISingleResponse) transportOrderStopRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := transportOrderStopRow{
		ID:                    resource.ID,
		Role:                  stringAttr(attrs, "role"),
		Status:                stringAttr(attrs, "status"),
		Position:              intAttr(attrs, "position"),
		AtMin:                 formatDateTime(stringAttr(attrs, "at-min")),
		AtMax:                 formatDateTime(stringAttr(attrs, "at-max")),
		ExternalTmsStopNumber: stringAttr(attrs, "external-tms-stop-number"),
	}

	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		row.TransportOrderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["location"]; ok && rel.Data != nil {
		row.LocationID = rel.Data.ID
	}

	return row
}
