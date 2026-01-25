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

type doMaintenanceRequirementMaintenanceRequirementPartsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	MaintenanceRequirement     string
	MaintenanceRequirementPart string
	Quantity                   string
	UnitCost                   string
	Source                     string
}

func newDoMaintenanceRequirementMaintenanceRequirementPartsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance requirement part link",
		Long: `Create a maintenance requirement part link.

Required flags:
  --maintenance-requirement       Maintenance requirement ID (required)
  --maintenance-requirement-part  Maintenance requirement part ID (required)

Optional flags:
  --quantity   Required quantity
  --unit-cost  Unit cost
  --source     Part source (stock or purchase)`,
		Example: `  # Create a maintenance requirement part link
  xbe do maintenance-requirement-maintenance-requirement-parts create \
    --maintenance-requirement 123 \
    --maintenance-requirement-part 456 \
    --quantity 2 \
    --unit-cost 15.50 \
    --source purchase

  # JSON output
  xbe do maintenance-requirement-maintenance-requirement-parts create \
    --maintenance-requirement 123 \
    --maintenance-requirement-part 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaintenanceRequirementMaintenanceRequirementPartsCreate,
	}
	initDoMaintenanceRequirementMaintenanceRequirementPartsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaintenanceRequirementMaintenanceRequirementPartsCmd.AddCommand(newDoMaintenanceRequirementMaintenanceRequirementPartsCreateCmd())
}

func initDoMaintenanceRequirementMaintenanceRequirementPartsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("maintenance-requirement", "", "Maintenance requirement ID (required)")
	cmd.Flags().String("maintenance-requirement-part", "", "Maintenance requirement part ID (required)")
	cmd.Flags().String("quantity", "", "Required quantity")
	cmd.Flags().String("unit-cost", "", "Unit cost")
	cmd.Flags().String("source", "", "Part source (stock or purchase)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaintenanceRequirementMaintenanceRequirementPartsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaintenanceRequirementMaintenanceRequirementPartsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaintenanceRequirement) == "" {
		err := fmt.Errorf("--maintenance-requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.MaintenanceRequirementPart) == "" {
		err := fmt.Errorf("--maintenance-requirement-part is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("unit-cost") {
		attributes["unit-cost"] = opts.UnitCost
	}
	if cmd.Flags().Changed("source") {
		attributes["source"] = opts.Source
	}

	relationships := map[string]any{
		"maintenance-requirement": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirements",
				"id":   opts.MaintenanceRequirement,
			},
		},
		"maintenance-requirement-part": map[string]any{
			"data": map[string]any{
				"type": "maintenance-requirement-parts",
				"id":   opts.MaintenanceRequirementPart,
			},
		},
	}

	data := map[string]any{
		"type":          "maintenance-requirement-maintenance-requirement-parts",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/maintenance-requirement-maintenance-requirement-parts", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created maintenance requirement part link %s\n", row.ID)
	return nil
}

func parseDoMaintenanceRequirementMaintenanceRequirementPartsCreateOptions(cmd *cobra.Command) (doMaintenanceRequirementMaintenanceRequirementPartsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maintenanceRequirement, _ := cmd.Flags().GetString("maintenance-requirement")
	maintenanceRequirementPart, _ := cmd.Flags().GetString("maintenance-requirement-part")
	quantity, _ := cmd.Flags().GetString("quantity")
	unitCost, _ := cmd.Flags().GetString("unit-cost")
	source, _ := cmd.Flags().GetString("source")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaintenanceRequirementMaintenanceRequirementPartsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		MaintenanceRequirement:     maintenanceRequirement,
		MaintenanceRequirementPart: maintenanceRequirementPart,
		Quantity:                   quantity,
		UnitCost:                   unitCost,
		Source:                     source,
	}, nil
}
