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

type doTransportOrderStopsUpdateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	ID                    string
	LocationID            string
	Role                  string
	Status                string
	Position              int
	AtMin                 string
	AtMax                 string
	ExternalTmsStopNumber string
}

func newDoTransportOrderStopsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transport order stop",
		Long: `Update a transport order stop.

Editable fields:
  --location                 Transport location ID
  --role                     Stop role (pickup, delivery)
  --status                   Stop status (planned, started, finished, cancelled)
  --position                 Stop position in the order
  --at-min                   Earliest scheduled time (ISO 8601)
  --at-max                   Latest scheduled time (ISO 8601)
  --external-tms-stop-number External TMS stop number

At least one flag is required.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a stop status
  xbe do transport-order-stops update 123 --status started

  # Update stop timing
  xbe do transport-order-stops update 123 \
    --at-min 2026-01-23T09:00:00Z \
    --at-max 2026-01-23T11:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportOrderStopsUpdate,
	}
	initDoTransportOrderStopsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderStopsCmd.AddCommand(newDoTransportOrderStopsUpdateCmd())
}

func initDoTransportOrderStopsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("location", "", "Transport location ID")
	cmd.Flags().String("role", "", "Stop role (pickup, delivery)")
	cmd.Flags().String("status", "", "Stop status (planned, started, finished, cancelled)")
	cmd.Flags().Int("position", 0, "Stop position in the order")
	cmd.Flags().String("at-min", "", "Earliest scheduled time (ISO 8601)")
	cmd.Flags().String("at-max", "", "Latest scheduled time (ISO 8601)")
	cmd.Flags().String("external-tms-stop-number", "", "External TMS stop number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderStopsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderStopsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("role") {
		attributes["role"] = opts.Role
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}
	if cmd.Flags().Changed("at-min") {
		attributes["at-min"] = opts.AtMin
	}
	if cmd.Flags().Changed("at-max") {
		attributes["at-max"] = opts.AtMax
	}
	if cmd.Flags().Changed("external-tms-stop-number") {
		attributes["external-tms-stop-number"] = opts.ExternalTmsStopNumber
	}

	if cmd.Flags().Changed("location") {
		if opts.LocationID == "" {
			relationships["location"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["location"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-locations",
					"id":   opts.LocationID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "transport-order-stops",
			"id":   opts.ID,
		},
	}

	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/transport-order-stops/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated transport order stop %s\n", row.ID)
	return nil
}

func parseDoTransportOrderStopsUpdateOptions(cmd *cobra.Command, args []string) (doTransportOrderStopsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	locationID, _ := cmd.Flags().GetString("location")
	role, _ := cmd.Flags().GetString("role")
	status, _ := cmd.Flags().GetString("status")
	position, _ := cmd.Flags().GetInt("position")
	atMin, _ := cmd.Flags().GetString("at-min")
	atMax, _ := cmd.Flags().GetString("at-max")
	externalTmsStopNumber, _ := cmd.Flags().GetString("external-tms-stop-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportOrderStopsUpdateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		ID:                    args[0],
		LocationID:            locationID,
		Role:                  role,
		Status:                status,
		Position:              position,
		AtMin:                 atMin,
		AtMax:                 atMax,
		ExternalTmsStopNumber: externalTmsStopNumber,
	}, nil
}
