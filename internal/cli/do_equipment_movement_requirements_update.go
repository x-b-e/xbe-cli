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

type doEquipmentMovementRequirementsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	Equipment           string
	CustomerExplicit    string
	InboundRequirement  string
	OutboundRequirement string
	Origin              string
	Destination         string
	OriginAtMin         string
	DestinationAtMax    string
	Note                string
}

func newDoEquipmentMovementRequirementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement requirement",
		Long: `Update an equipment movement requirement.

Optional flags:
  --equipment            Equipment ID
  --customer-explicit    Customer ID (explicit override, empty to clear)
  --inbound-requirement  Inbound equipment requirement ID (empty to clear)
  --outbound-requirement Outbound equipment requirement ID (empty to clear)
  --origin               Origin location ID (empty to clear)
  --destination          Destination location ID (empty to clear)
  --origin-at-min         Earliest origin time (ISO 8601)
  --destination-at-max    Latest destination time (ISO 8601)
  --note                  Note`,
		Example: `  # Update note
  xbe do equipment-movement-requirements update 123 --note "Updated note"

  # Update dates
  xbe do equipment-movement-requirements update 123 --origin-at-min "2025-01-02T08:00:00Z" --destination-at-max "2025-01-02T17:00:00Z"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementRequirementsUpdate,
	}
	initDoEquipmentMovementRequirementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementRequirementsCmd.AddCommand(newDoEquipmentMovementRequirementsUpdateCmd())
}

func initDoEquipmentMovementRequirementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("customer-explicit", "", "Customer ID (explicit override, empty to clear)")
	cmd.Flags().String("inbound-requirement", "", "Inbound equipment requirement ID (empty to clear)")
	cmd.Flags().String("outbound-requirement", "", "Outbound equipment requirement ID (empty to clear)")
	cmd.Flags().String("origin", "", "Origin location ID (empty to clear)")
	cmd.Flags().String("destination", "", "Destination location ID (empty to clear)")
	cmd.Flags().String("origin-at-min", "", "Earliest origin time (ISO 8601)")
	cmd.Flags().String("destination-at-max", "", "Latest destination time (ISO 8601)")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementRequirementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementRequirementsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("origin-at-min") {
		attributes["origin-at-min"] = opts.OriginAtMin
	}
	if cmd.Flags().Changed("destination-at-max") {
		attributes["destination-at-max"] = opts.DestinationAtMax
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
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
	if cmd.Flags().Changed("customer-explicit") {
		if opts.CustomerExplicit == "" {
			relationships["customer-explicit"] = map[string]any{"data": nil}
		} else {
			relationships["customer-explicit"] = map[string]any{
				"data": map[string]any{
					"type": "customers",
					"id":   opts.CustomerExplicit,
				},
			}
		}
	}
	if cmd.Flags().Changed("inbound-requirement") {
		if opts.InboundRequirement == "" {
			relationships["inbound-requirement"] = map[string]any{"data": nil}
		} else {
			relationships["inbound-requirement"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-requirements",
					"id":   opts.InboundRequirement,
				},
			}
		}
	}
	if cmd.Flags().Changed("outbound-requirement") {
		if opts.OutboundRequirement == "" {
			relationships["outbound-requirement"] = map[string]any{"data": nil}
		} else {
			relationships["outbound-requirement"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-requirements",
					"id":   opts.OutboundRequirement,
				},
			}
		}
	}
	if cmd.Flags().Changed("origin") {
		if opts.Origin == "" {
			relationships["origin"] = map[string]any{"data": nil}
		} else {
			relationships["origin"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-movement-requirement-locations",
					"id":   opts.Origin,
				},
			}
		}
	}
	if cmd.Flags().Changed("destination") {
		if opts.Destination == "" {
			relationships["destination"] = map[string]any{"data": nil}
		} else {
			relationships["destination"] = map[string]any{
				"data": map[string]any{
					"type": "equipment-movement-requirement-locations",
					"id":   opts.Destination,
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
		"type": "equipment-movement-requirements",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-requirements/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := buildEquipmentMovementRequirementRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement requirement %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentMovementRequirementsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementRequirementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipment, _ := cmd.Flags().GetString("equipment")
	customerExplicit, _ := cmd.Flags().GetString("customer-explicit")
	inboundRequirement, _ := cmd.Flags().GetString("inbound-requirement")
	outboundRequirement, _ := cmd.Flags().GetString("outbound-requirement")
	origin, _ := cmd.Flags().GetString("origin")
	destination, _ := cmd.Flags().GetString("destination")
	originAtMin, _ := cmd.Flags().GetString("origin-at-min")
	destinationAtMax, _ := cmd.Flags().GetString("destination-at-max")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementRequirementsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		Equipment:           equipment,
		CustomerExplicit:    customerExplicit,
		InboundRequirement:  inboundRequirement,
		OutboundRequirement: outboundRequirement,
		Origin:              origin,
		Destination:         destination,
		OriginAtMin:         originAtMin,
		DestinationAtMax:    destinationAtMax,
		Note:                note,
	}, nil
}
