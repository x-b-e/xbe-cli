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

type doMaintenanceRequirementMaintenanceRequirementPartsUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
	MaintenanceRequirement     string
	MaintenanceRequirementPart string
	Quantity                   string
	UnitCost                   string
	Source                     string
}

func newDoMaintenanceRequirementMaintenanceRequirementPartsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a maintenance requirement part link",
		Long: `Update a maintenance requirement part link.

Optional flags:
  --quantity   Required quantity
  --unit-cost  Unit cost
  --source     Part source (stock or purchase)

Relationships:
  --maintenance-requirement       Maintenance requirement ID
  --maintenance-requirement-part  Maintenance requirement part ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity and source
  xbe do maintenance-requirement-maintenance-requirement-parts update 123 \
    --quantity 3 \
    --source stock

  # Update unit cost
  xbe do maintenance-requirement-maintenance-requirement-parts update 123 --unit-cost 20`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaintenanceRequirementMaintenanceRequirementPartsUpdate,
	}
	initDoMaintenanceRequirementMaintenanceRequirementPartsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementMaintenanceRequirementPartsCmd.AddCommand(newDoMaintenanceRequirementMaintenanceRequirementPartsUpdateCmd())
}

func initDoMaintenanceRequirementMaintenanceRequirementPartsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID")
	cmd.Flags().String("maintenance-requirement-part", "", "Maintenance requirement part ID")
	cmd.Flags().String("quantity", "", "Required quantity")
	cmd.Flags().String("unit-cost", "", "Unit cost")
	cmd.Flags().String("source", "", "Part source (stock or purchase)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementMaintenanceRequirementPartsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaintenanceRequirementMaintenanceRequirementPartsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("unit-cost") {
		attributes["unit-cost"] = opts.UnitCost
	}
	if cmd.Flags().Changed("source") {
		attributes["source"] = opts.Source
	}

	if cmd.Flags().Changed("maintenance-requirement") {
		relationships["maintenance-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   opts.MaintenanceRequirement,
			},
		}
	}
	if cmd.Flags().Changed("maintenance-requirement-part") {
		relationships["maintenance-requirement-part"] = map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-parts",
				"id":   opts.MaintenanceRequirementPart,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "maintenance-requirement-maintenance-requirement-parts",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/maintenance-requirement-maintenance-requirement-parts/"+opts.ID, jsonBody)
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

	row := buildMaintenanceRequirementMaintenanceRequirementPartRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated maintenance requirement part link %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementMaintenanceRequirementPartsUpdateOptions(cmd *cobra.Command, args []string) (doMaintenanceRequirementMaintenanceRequirementPartsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	maintenanceRequirementPart, _ := cmd.Flags().GetString("maintenance-requirement-part")
	quantity, _ := cmd.Flags().GetString("quantity")
	unitCost, _ := cmd.Flags().GetString("unit-cost")
	source, _ := cmd.Flags().GetString("source")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementMaintenanceRequirementPartsUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
		MaintenanceRequirement:     maintenanceRequirement,
		MaintenanceRequirementPart: maintenanceRequirementPart,
		Quantity:                   quantity,
		UnitCost:                   unitCost,
		Source:                     source,
	}, nil
}
