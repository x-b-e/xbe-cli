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

type maintenanceRequirementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type requirementDetails struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	DueOn       string `json:"due_on,omitempty"`
	IsTemplate  bool   `json:"is_template"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`

	// Equipment
	EquipmentID     string `json:"equipment_id,omitempty"`
	EquipmentName   string `json:"equipment_name,omitempty"`
	EquipmentNumber string `json:"equipment_number,omitempty"`

	// Set
	SetID   string `json:"set_id,omitempty"`
	SetName string `json:"set_name,omitempty"`

	// Additional fields
	Notes            string `json:"notes,omitempty"`
	EstimatedMinutes int    `json:"estimated_minutes,omitempty"`
	ActualMinutes    int    `json:"actual_minutes,omitempty"`
	CompletedAt      string `json:"completed_at,omitempty"`
}

func newMaintenanceRequirementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement details",
		Long: `Show the full details of a maintenance requirement.

Retrieves and displays comprehensive information about a requirement including
equipment association, set membership, and completion status.

Output Sections (table format):
  Core Info       ID, status, description, due date
  Equipment       Associated equipment name and number
  Set             Parent requirement set
  Details         Notes, estimated/actual time, completion

Arguments:
  <id>          The requirement ID (required). Find IDs using the list command.`,
		Example: `  # View a requirement by ID
  xbe view maintenance requirements show 123

  # Get requirement as JSON
  xbe view maintenance requirements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceRequirementsShow,
	}
	initMaintenanceRequirementsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementsCmd.AddCommand(newMaintenanceRequirementsShowCmd())
}

func initMaintenanceRequirementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceRequirementsShowOptions(cmd)
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
		return fmt.Errorf("requirement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "equipment,maintenance-requirement-sets")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirements/"+id, query)
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

	details := buildRequirementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRequirementDetails(cmd, details)
}

func parseMaintenanceRequirementsShowOptions(cmd *cobra.Command) (maintenanceRequirementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRequirementDetails(resp jsonAPISingleResponse) requirementDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := requirementDetails{
		ID:               resp.Data.ID,
		Status:           stringAttr(attrs, "status"),
		Description:      strings.TrimSpace(stringAttr(attrs, "description")),
		DueOn:            formatDate(stringAttr(attrs, "due-on")),
		IsTemplate:       boolAttr(attrs, "is-template"),
		CreatedAt:        formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt:        formatDate(stringAttr(attrs, "updated-at")),
		Notes:            strings.TrimSpace(stringAttr(attrs, "notes")),
		EstimatedMinutes: intAttr(attrs, "estimated-minutes"),
		ActualMinutes:    intAttr(attrs, "actual-minutes"),
		CompletedAt:      formatDate(stringAttr(attrs, "completed-at")),
	}

	// Equipment
	if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.EquipmentName = stringAttr(inc.Attributes, "name")
			details.EquipmentNumber = stringAttr(inc.Attributes, "equipment-number")
		}
	}

	// Set (may be array)
	if rel, ok := resp.Data.Relationships["maintenance-requirement-sets"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil && len(refs) > 0 {
			details.SetID = refs[0].ID
			key := resourceKey(refs[0].Type, refs[0].ID)
			if inc, ok := included[key]; ok {
				details.SetName = stringAttr(inc.Attributes, "template-name")
			}
		}
	} else if rel, ok := resp.Data.Relationships["maintenance-requirement-sets"]; ok && rel.Data != nil {
		details.SetID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.SetName = stringAttr(inc.Attributes, "template-name")
		}
	}

	return details
}

func renderRequirementDetails(cmd *cobra.Command, d requirementDetails) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Status: %s\n", d.Status)
	if d.IsTemplate {
		fmt.Fprintln(out, "Is Template: true")
	}
	if d.DueOn != "" {
		fmt.Fprintf(out, "Due On: %s\n", d.DueOn)
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}

	// Equipment
	if d.EquipmentID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Equipment:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  ID: %s\n", d.EquipmentID)
		if d.EquipmentName != "" {
			fmt.Fprintf(out, "  Name: %s\n", d.EquipmentName)
		}
		if d.EquipmentNumber != "" {
			fmt.Fprintf(out, "  Number: %s\n", d.EquipmentNumber)
		}
	}

	// Set
	if d.SetID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Requirement Set:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  ID: %s\n", d.SetID)
		if d.SetName != "" {
			fmt.Fprintf(out, "  Name: %s\n", d.SetName)
		}
	}

	// Time tracking
	if d.EstimatedMinutes > 0 || d.ActualMinutes > 0 || d.CompletedAt != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Time Tracking:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.EstimatedMinutes > 0 {
			fmt.Fprintf(out, "  Estimated: %d minutes\n", d.EstimatedMinutes)
		}
		if d.ActualMinutes > 0 {
			fmt.Fprintf(out, "  Actual: %d minutes\n", d.ActualMinutes)
		}
		if d.CompletedAt != "" {
			fmt.Fprintf(out, "  Completed At: %s\n", d.CompletedAt)
		}
	}

	// Description and notes
	if d.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Description)
	}

	if d.Notes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Notes:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Notes)
	}

	return nil
}
