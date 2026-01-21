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

type maintenanceSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type setDetails struct {
	ID        string `json:"id"`
	Status    string `json:"status,omitempty"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	// Equipment
	EquipmentID     string `json:"equipment_id,omitempty"`
	EquipmentName   string `json:"equipment_name,omitempty"`
	EquipmentNumber string `json:"equipment_number,omitempty"`

	// Requirements
	Requirements []setRequirement `json:"requirements,omitempty"`

	// Stats
	TotalCount     int `json:"total_count"`
	CompletedCount int `json:"completed_count"`
}

type setRequirement struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
}

func newMaintenanceSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show maintenance requirement set details",
		Long: `Show the full details of a maintenance requirement set.

Retrieves and displays comprehensive information about a set including
equipment association, requirements list, and completion status.

Output Sections (table format):
  Core Info       ID, status, type, name
  Equipment       Associated equipment
  Requirements    List of requirements in the set
  Stats           Completion progress

Arguments:
  <id>          The set ID (required). Find IDs using the list command.`,
		Example: `  # View a set by ID
  xbe view maintenance sets show 123

  # Get set as JSON
  xbe view maintenance sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaintenanceSetsShow,
	}
	initMaintenanceSetsShowFlags(cmd)
	return cmd
}

func init() {
	maintenanceSetsCmd.AddCommand(newMaintenanceSetsShowCmd())
}

func initMaintenanceSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaintenanceSetsShowOptions(cmd)
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
		return fmt.Errorf("set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "maintenance-requirements,equipments")

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-sets/"+id, query)
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

	details := buildSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSetDetails(cmd, details)
}

func parseMaintenanceSetsShowOptions(cmd *cobra.Command) (maintenanceSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSetDetails(resp jsonAPISingleResponse) setDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := setDetails{
		ID:        resp.Data.ID,
		Status:    stringAttr(attrs, "status"),
		Type:      stringAttr(attrs, "maintenance-type"),
		Name:      strings.TrimSpace(stringAttr(attrs, "template-name")),
		CreatedAt: formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDate(stringAttr(attrs, "updated-at")),
	}

	// Equipment (may be array)
	if rel, ok := resp.Data.Relationships["equipments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil && len(refs) > 0 {
			details.EquipmentID = refs[0].ID
			key := resourceKey(refs[0].Type, refs[0].ID)
			if inc, ok := included[key]; ok {
				details.EquipmentName = stringAttr(inc.Attributes, "name")
				details.EquipmentNumber = stringAttr(inc.Attributes, "equipment-number")
			}
		}
	} else if rel, ok := resp.Data.Relationships["equipment"]; ok && rel.Data != nil {
		details.EquipmentID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.EquipmentName = stringAttr(inc.Attributes, "name")
			details.EquipmentNumber = stringAttr(inc.Attributes, "equipment-number")
		}
	}

	// Requirements
	if rel, ok := resp.Data.Relationships["maintenance-requirements"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			details.TotalCount = len(refs)
			for _, ref := range refs {
				key := resourceKey(ref.Type, ref.ID)
				if inc, ok := included[key]; ok {
					req := setRequirement{
						ID:          inc.ID,
						Status:      stringAttr(inc.Attributes, "status"),
						Description: strings.TrimSpace(stringAttr(inc.Attributes, "description")),
					}
					details.Requirements = append(details.Requirements, req)
					if req.Status == "completed" {
						details.CompletedCount++
					}
				}
			}
		}
	}

	return details
}

func renderSetDetails(cmd *cobra.Command, d setDetails) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Status: %s\n", d.Status)
	if d.Type != "" {
		fmt.Fprintf(out, "Type: %s\n", d.Type)
	}
	if d.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
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

	// Stats
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Progress:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Completed: %d / %d\n", d.CompletedCount, d.TotalCount)
	if d.TotalCount > 0 {
		pct := float64(d.CompletedCount) / float64(d.TotalCount) * 100
		fmt.Fprintf(out, "  Percentage: %.0f%%\n", pct)
	}

	// Requirements
	if len(d.Requirements) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Requirements (%d):\n", len(d.Requirements))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for i, req := range d.Requirements {
			desc := req.Description
			if desc == "" {
				desc = "(no description)"
			}
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Fprintf(out, "  %d. [%s] %s (ID: %s)\n", i+1, req.Status, desc, req.ID)
		}
	}

	return nil
}
