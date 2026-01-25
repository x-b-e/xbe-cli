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

type maintenanceRequirementPartsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type maintenanceRequirementPartDetails struct {
	ID                        string `json:"id"`
	PartNumber                string `json:"part_number,omitempty"`
	Name                      string `json:"name,omitempty"`
	Description               string `json:"description,omitempty"`
	Notes                     string `json:"notes,omitempty"`
	IsTemplate                bool   `json:"is_template"`
	Make                      string `json:"make,omitempty"`
	Model                     string `json:"model,omitempty"`
	Year                      string `json:"year,omitempty"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
}

func newMaintenanceRequirementPartsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement part details",
		Long: `Show the full details of a maintenance requirement part.

Output Fields:
  ID           Maintenance requirement part identifier
  Name         Part name
  Part Number  Part number
  Make         Part make
  Model        Part model
  Year         Part year
  Template     Whether the part is a template
  Broker       Broker ID (template parts)
  Equip Class  Equipment classification ID
  Description  Part description
  Notes        Additional notes

Arguments:
  <id>         The maintenance requirement part ID (required).`,
		Example: `  # Show a maintenance requirement part
  xbe view maintenance-requirement-parts show 123

  # Output as JSON
  xbe view maintenance-requirement-parts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementPartsShow,
	}
	initMaintenanceRequirementPartsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementPartsCmd.AddCommand(newMaintenanceRequirementPartsShowCmd())
}

func initMaintenanceRequirementPartsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementPartsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementPartsShowOptions(cmd)
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
		return fmt.Errorf("maintenance requirement part id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[maintenance-requirement-parts]", "part-number,name,description,notes,is-template,make,model,year,broker,equipment-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-parts/"+id, query)
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

	details := buildMaintenanceRequirementPartDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaintenanceRequirementPartDetails(cmd, details)
}

func parseMaintenanceRequirementPartsShowOptions(cmd *cobra.Command) (maintenanceRequirementPartsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return maintenanceRequirementPartsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return maintenanceRequirementPartsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return maintenanceRequirementPartsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return maintenanceRequirementPartsShowOptions{}, err
	}

	return maintenanceRequirementPartsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaintenanceRequirementPartDetails(resp jsonAPISingleResponse) maintenanceRequirementPartDetails {
	attrs := resp.Data.Attributes

	details := maintenanceRequirementPartDetails{
		ID:          resp.Data.ID,
		PartNumber:  stringAttr(attrs, "part-number"),
		Name:        stringAttr(attrs, "name"),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		Notes:       strings.TrimSpace(stringAttr(attrs, "notes")),
		IsTemplate:  boolAttr(attrs, "is-template"),
		Make:        stringAttr(attrs, "make"),
		Model:       stringAttr(attrs, "model"),
		Year:        stringAttr(attrs, "year"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		details.EquipmentClassificationID = rel.Data.ID
	}

	return details
}

func renderMaintenanceRequirementPartDetails(cmd *cobra.Command, details maintenanceRequirementPartDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.PartNumber != "" {
		fmt.Fprintf(out, "Part Number: %s\n", details.PartNumber)
	}
	if details.Make != "" {
		fmt.Fprintf(out, "Make: %s\n", details.Make)
	}
	if details.Model != "" {
		fmt.Fprintf(out, "Model: %s\n", details.Model)
	}
	if details.Year != "" {
		fmt.Fprintf(out, "Year: %s\n", details.Year)
	}
	fmt.Fprintf(out, "Template: %s\n", formatYesNo(details.IsTemplate))
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.EquipmentClassificationID != "" {
		fmt.Fprintf(out, "Equip Class: %s\n", details.EquipmentClassificationID)
	}
	if details.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Description)
	}
	if details.Notes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Notes:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Notes)
	}

	return nil
}
