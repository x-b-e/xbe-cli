package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type maintenanceRequirementMaintenanceRequirementPartsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementMaintenanceRequirementPartDetails struct {
	ID                           string `json:"id"`
	MaintenanceRequirementID     string `json:"maintenance_requirement_id,omitempty"`
	MaintenanceRequirement       string `json:"maintenance_requirement,omitempty"`
	MaintenanceRequirementStatus string `json:"maintenance_requirement_status,omitempty"`
	MaintenanceRequirementPartID string `json:"maintenance_requirement_part_id,omitempty"`
	Quantity                     string `json:"quantity,omitempty"`
	UnitCost                     string `json:"unit_cost,omitempty"`
	TotalCost                    string `json:"total_cost,omitempty"`
	Source                       string `json:"source,omitempty"`
	PartName                     string `json:"part_name,omitempty"`
	PartNumber                   string `json:"part_part_number,omitempty"`
	PartDescription              string `json:"part_description,omitempty"`
	PartMake                     string `json:"part_make,omitempty"`
	PartModel                    string `json:"part_model,omitempty"`
	PartYear                     string `json:"part_year,omitempty"`
	PartNotes                    string `json:"part_notes,omitempty"`
}

func newMaintenanceRequirementMaintenanceRequirementPartsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement part details",
		Long: `Show the full details of a maintenance requirement part link.

Output Fields:
  ID                           Link identifier
  Maintenance Requirement      Requirement description or template name
  Maintenance Requirement ID   Requirement identifier
  Maintenance Requirement Status Requirement status
  Maintenance Requirement Part ID Part identifier
  Quantity                     Required quantity
  Unit Cost                    Unit cost
  Total Cost                   Total cost
  Source                       Part source
  Part Name                    Part name
  Part Part Number             Part number
  Part Description             Part description
  Part Make                    Part make
  Part Model                   Part model
  Part Year                    Part year
  Part Notes                   Part notes

Arguments:
  <id>    Maintenance requirement part link ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a maintenance requirement part link
  xbe view maintenance-requirement-maintenance-requirement-parts show 123

  # JSON output
  xbe view maintenance-requirement-maintenance-requirement-parts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementMaintenanceRequirementPartsShow,
	}
	initMaintenanceRequirementMaintenanceRequirementPartsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementMaintenanceRequirementPartsCmd.AddCommand(newMaintenanceRequirementMaintenanceRequirementPartsShowCmd())
}

func initMaintenanceRequirementMaintenanceRequirementPartsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementMaintenanceRequirementPartsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementMaintenanceRequirementPartsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("maintenance requirement part link id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[maintenance-requirement-maintenance-requirement-parts]", "quantity,unit-cost,source,total-cost,part-name,part-part-number,part-description,part-make,part-model,part-year,part-notes,maintenance-requirement,maintenance-requirement-part")
	query.Set("fields[maintenance-requirements]", "template-name,description,status")
	query.Set("include", "maintenance-requirement")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-maintenance-requirement-parts/"+id, query)
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

	details := buildMaintenanceRequirementMaintenanceRequirementPartDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementMaintenanceRequirementPartDetails(cmd, details)
}

func parseMaintenanceRequirementMaintenanceRequirementPartsShowOptions(cmd *cobra.Command) (maintenanceRequirementMaintenanceRequirementPartsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementMaintenanceRequirementPartsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementMaintenanceRequirementPartDetails(resp jsonAPISingleResponse) maintenanceRequirementMaintenanceRequirementPartDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := maintenanceRequirementMaintenanceRequirementPartDetails{
		ID:              resp.Data.ID,
		Quantity:        stringAttr(resp.Data.Attributes, "quantity"),
		UnitCost:        stringAttr(resp.Data.Attributes, "unit-cost"),
		TotalCost:       stringAttr(resp.Data.Attributes, "total-cost"),
		Source:          stringAttr(resp.Data.Attributes, "source"),
		PartName:        stringAttr(resp.Data.Attributes, "part-name"),
		PartNumber:      stringAttr(resp.Data.Attributes, "part-part-number"),
		PartDescription: stringAttr(resp.Data.Attributes, "part-description"),
		PartMake:        stringAttr(resp.Data.Attributes, "part-make"),
		PartModel:       stringAttr(resp.Data.Attributes, "part-model"),
		PartYear:        stringAttr(resp.Data.Attributes, "part-year"),
		PartNotes:       stringAttr(resp.Data.Attributes, "part-notes"),
	}

	if rel, ok := resp.Data.Relationships["maintenance-requirement"]; ok && rel.Data != nil {
		details.MaintenanceRequirementID = rel.Data.ID
		if req, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			templateName := stringAttr(req.Attributes, "template-name")
			description := stringAttr(req.Attributes, "description")
			details.MaintenanceRequirement = firstNonEmpty(templateName, description)
			details.MaintenanceRequirementStatus = stringAttr(req.Attributes, "status")
		}
	}
	if rel, ok := resp.Data.Relationships["maintenance-requirement-part"]; ok && rel.Data != nil {
		details.MaintenanceRequirementPartID = rel.Data.ID
	}

	return details
}

func renderMaintenanceRequirementMaintenanceRequirementPartDetails(cmd *cobra.Command, details maintenanceRequirementMaintenanceRequirementPartDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaintenanceRequirement != "" {
		fmt.Fprintf(out, "Maintenance Requirement: %s\n", details.MaintenanceRequirement)
	}
	if details.MaintenanceRequirementID != "" {
		fmt.Fprintf(out, "Maintenance Requirement ID: %s\n", details.MaintenanceRequirementID)
	}
	if details.MaintenanceRequirementStatus != "" {
		fmt.Fprintf(out, "Maintenance Requirement Status: %s\n", details.MaintenanceRequirementStatus)
	}
	if details.MaintenanceRequirementPartID != "" {
		fmt.Fprintf(out, "Maintenance Requirement Part ID: %s\n", details.MaintenanceRequirementPartID)
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.UnitCost != "" {
		fmt.Fprintf(out, "Unit Cost: %s\n", details.UnitCost)
	}
	if details.TotalCost != "" {
		fmt.Fprintf(out, "Total Cost: %s\n", details.TotalCost)
	}
	if details.Source != "" {
		fmt.Fprintf(out, "Source: %s\n", details.Source)
	}
	if details.PartName != "" {
		fmt.Fprintf(out, "Part Name: %s\n", details.PartName)
	}
	if details.PartNumber != "" {
		fmt.Fprintf(out, "Part Number: %s\n", details.PartNumber)
	}
	if details.PartDescription != "" {
		fmt.Fprintf(out, "Part Description: %s\n", details.PartDescription)
	}
	if details.PartMake != "" {
		fmt.Fprintf(out, "Part Make: %s\n", details.PartMake)
	}
	if details.PartModel != "" {
		fmt.Fprintf(out, "Part Model: %s\n", details.PartModel)
	}
	if details.PartYear != "" {
		fmt.Fprintf(out, "Part Year: %s\n", details.PartYear)
	}
	if details.PartNotes != "" {
		fmt.Fprintf(out, "Part Notes: %s\n", details.PartNotes)
	}

	return nil
}
