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

type doEquipmentUtilizationReadingsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Equipment     string
	BusinessUnit  string
	User          string
	ReportedAt    string
	Odometer      string
	Hourmeter     string
	OtherReadings string
}

func newDoEquipmentUtilizationReadingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment utilization reading",
		Long: `Update an equipment utilization reading.

Optional flags:
  --equipment       Equipment ID
  --business-unit   Business unit ID
  --user            User ID
  --reported-at     Reported timestamp (ISO 8601)
  --odometer        Odometer reading
  --hourmeter       Hourmeter reading
  --other-readings  Other readings payload (JSON string)`,
		Example: `  # Update readings
  xbe do equipment-utilization-readings update 123 --odometer 120

  # Update reported-at and hourmeter
  xbe do equipment-utilization-readings update 123 --reported-at 2025-01-03T08:00:00Z --hourmeter 14

  # Update other readings
  xbe do equipment-utilization-readings update 123 --other-readings '{"source":"manual"}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentUtilizationReadingsUpdate,
	}
	initDoEquipmentUtilizationReadingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentUtilizationReadingsCmd.AddCommand(newDoEquipmentUtilizationReadingsUpdateCmd())
}

func initDoEquipmentUtilizationReadingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("reported-at", "", "Reported timestamp (ISO 8601)")
	cmd.Flags().String("odometer", "", "Odometer reading")
	cmd.Flags().String("hourmeter", "", "Hourmeter reading")
	cmd.Flags().String("other-readings", "", "Other readings payload (JSON string)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentUtilizationReadingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentUtilizationReadingsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("reported-at") {
		attributes["reported-at"] = opts.ReportedAt
	}
	if cmd.Flags().Changed("odometer") {
		attributes["odometer"] = opts.Odometer
	}
	if cmd.Flags().Changed("hourmeter") {
		attributes["hourmeter"] = opts.Hourmeter
	}
	if cmd.Flags().Changed("other-readings") {
		if strings.TrimSpace(opts.OtherReadings) == "" {
			attributes["other-readings"] = nil
		} else {
			var other any
			if err := json.Unmarshal([]byte(opts.OtherReadings), &other); err != nil {
				err := fmt.Errorf("invalid other-readings JSON: %w", err)
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["other-readings"] = other
		}
	}

	if cmd.Flags().Changed("equipment") {
		if opts.Equipment == "" {
			relationships["equipment"] = map[string]any{"data": nil}
		} else {
			relationships["equipment"] = map[string]any{
				"data": map[string]any{
					"type": "equipment",
					"id":   opts.Equipment,
				},
			}
		}
	}
	if cmd.Flags().Changed("business-unit") {
		if opts.BusinessUnit == "" {
			relationships["business-unit"] = map[string]any{"data": nil}
		} else {
			relationships["business-unit"] = map[string]any{
				"data": map[string]any{
					"type": "business-units",
					"id":   opts.BusinessUnit,
				},
			}
		}
	}
	if cmd.Flags().Changed("user") {
		if opts.User == "" {
			relationships["user"] = map[string]any{"data": nil}
		} else {
			relationships["user"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.User,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment-utilization-readings",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-utilization-readings/"+opts.ID, jsonBody)
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

	row := buildEquipmentUtilizationReadingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment utilization reading %s\n", row.ID)
	return nil
}

func parseDoEquipmentUtilizationReadingsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentUtilizationReadingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipment, _ := cmd.Flags().GetString("equipment")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	user, _ := cmd.Flags().GetString("user")
	reportedAt, _ := cmd.Flags().GetString("reported-at")
	odometer, _ := cmd.Flags().GetString("odometer")
	hourmeter, _ := cmd.Flags().GetString("hourmeter")
	otherReadings, _ := cmd.Flags().GetString("other-readings")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentUtilizationReadingsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Equipment:     equipment,
		BusinessUnit:  businessUnit,
		User:          user,
		ReportedAt:    reportedAt,
		Odometer:      odometer,
		Hourmeter:     hourmeter,
		OtherReadings: otherReadings,
	}, nil
}
