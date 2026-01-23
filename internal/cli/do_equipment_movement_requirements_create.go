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

type doEquipmentMovementRequirementsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Broker              string
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

func newDoEquipmentMovementRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement requirement",
		Long: `Create an equipment movement requirement.

Required (unless inbound/outbound requirements provide broker/equipment):
  --broker      Broker ID
  --equipment   Equipment ID

Optional:
  --customer-explicit    Customer ID (explicit override)
  --inbound-requirement  Inbound equipment requirement ID
  --outbound-requirement Outbound equipment requirement ID
  --origin               Origin location ID
  --destination          Destination location ID
  --origin-at-min         Earliest origin time (ISO 8601)
  --destination-at-max    Latest destination time (ISO 8601)
  --note                  Note`,
		Example: `  # Create with broker + equipment
  xbe do equipment-movement-requirements create --broker 123 --equipment 456

  # Create with dates and note
  xbe do equipment-movement-requirements create --broker 123 --equipment 456 \
    --origin-at-min "2025-01-01T08:00:00Z" --destination-at-max "2025-01-01T17:00:00Z" \
    --note "Move to yard"`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementRequirementsCreate,
	}
	initDoEquipmentMovementRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementRequirementsCmd.AddCommand(newDoEquipmentMovementRequirementsCreateCmd())
}

func initDoEquipmentMovementRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("customer-explicit", "", "Customer ID (explicit override)")
	cmd.Flags().String("inbound-requirement", "", "Inbound equipment requirement ID")
	cmd.Flags().String("outbound-requirement", "", "Outbound equipment requirement ID")
	cmd.Flags().String("origin", "", "Origin location ID")
	cmd.Flags().String("destination", "", "Destination location ID")
	cmd.Flags().String("origin-at-min", "", "Earliest origin time (ISO 8601)")
	cmd.Flags().String("destination-at-max", "", "Latest destination time (ISO 8601)")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementRequirementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementRequirementsCreateOptions(cmd)
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

	if opts.InboundRequirement == "" && opts.OutboundRequirement == "" {
		if opts.Broker == "" {
			err := fmt.Errorf("--broker is required when inbound/outbound requirements are not set")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if opts.Equipment == "" {
			err := fmt.Errorf("--equipment is required when inbound/outbound requirements are not set")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if opts.OriginAtMin != "" {
		attributes["origin-at-min"] = opts.OriginAtMin
	}
	if opts.DestinationAtMax != "" {
		attributes["destination-at-max"] = opts.DestinationAtMax
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{}
	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if opts.CustomerExplicit != "" {
		relationships["customer-explicit"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerExplicit,
			},
		}
	}
	if opts.InboundRequirement != "" {
		relationships["inbound-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-requirements",
				"id":   opts.InboundRequirement,
			},
		}
	}
	if opts.OutboundRequirement != "" {
		relationships["outbound-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-requirements",
				"id":   opts.OutboundRequirement,
			},
		}
	}
	if opts.Origin != "" {
		relationships["origin"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirement-locations",
				"id":   opts.Origin,
			},
		}
	}
	if opts.Destination != "" {
		relationships["destination"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirement-locations",
				"id":   opts.Destination,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-requirements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-requirements", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement requirement %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentMovementRequirementsCreateOptions(cmd *cobra.Command) (doEquipmentMovementRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
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

	return doEquipmentMovementRequirementsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Broker:              broker,
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
