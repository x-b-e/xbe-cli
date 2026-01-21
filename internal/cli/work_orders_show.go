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

type workOrdersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type workOrderDetails struct {
	ID              string `json:"id"`
	Status          string `json:"status,omitempty"`
	Priority        string `json:"priority,omitempty"`
	DueDate         string `json:"due_date,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	Notes           string `json:"notes,omitempty"`
	SafetyTagStatus string `json:"safety_tag_status,omitempty"`

	// Responsible party (business unit)
	BusinessUnitID string `json:"business_unit_id,omitempty"`
	BusinessUnit   string `json:"business_unit,omitempty"`

	// Service site
	ServiceSiteID string `json:"service_site_id,omitempty"`
	ServiceSite   string `json:"service_site,omitempty"`

	// Related sets
	RequirementSets []workOrderSetSummary `json:"requirement_sets,omitempty"`

	// Assigned users
	AssignedUsers []string `json:"assigned_users,omitempty"`
}

type workOrderSetSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Type   string `json:"type,omitempty"`
}

func newWorkOrdersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show work order details",
		Long: `Show the full details of a work order.

Retrieves and displays comprehensive information about a work order including
status, priority, responsible party, service site, and related requirement sets.

Output Sections (table format):
  Core Info       ID, status, priority, due date, dates
  Location        Responsible business unit, service site
  Requirement     Related requirement sets with status
  Sets
  Assigned        Users assigned to this work order
  Users
  Notes           Any notes or comments

Arguments:
  <id>          The work order ID (required). Find IDs using the list command.`,
		Example: `  # View a work order by ID
  xbe view work-orders show 123

  # Get work order as JSON
  xbe view work-orders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runWorkOrdersShow,
	}
	initWorkOrdersShowFlags(cmd)
	return cmd
}

func init() {
	workOrdersCmd.AddCommand(newWorkOrdersShowCmd())
}

func initWorkOrdersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrdersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseWorkOrdersShowOptions(cmd)
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
		return fmt.Errorf("work order id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "responsible-party,service-site,maintenance-requirement-sets,assigned-users")

	body, _, err := client.Get(cmd.Context(), "/v1/work-orders/"+id, query)
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

	details := buildWorkOrderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderWorkOrderDetails(cmd, details)
}

func parseWorkOrdersShowOptions(cmd *cobra.Command) (workOrdersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrdersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildWorkOrderDetails(resp jsonAPISingleResponse) workOrderDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := workOrderDetails{
		ID:              resp.Data.ID,
		Status:          stringAttr(attrs, "status"),
		Priority:        stringAttr(attrs, "priority"),
		DueDate:         formatDate(stringAttr(attrs, "due-date")),
		CreatedAt:       formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt:       formatDate(stringAttr(attrs, "updated-at")),
		CompletedAt:     formatDate(stringAttr(attrs, "completed-at")),
		Notes:           strings.TrimSpace(stringAttr(attrs, "notes")),
		SafetyTagStatus: stringAttr(attrs, "safety-tag-status"),
	}

	// Responsible party (business unit)
	if rel, ok := resp.Data.Relationships["responsible-party"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.BusinessUnit = firstNonEmpty(
				stringAttr(inc.Attributes, "company-name"),
				stringAttr(inc.Attributes, "name"),
				rel.Data.ID,
			)
		}
	}

	// Service site
	if rel, ok := resp.Data.Relationships["service-site"]; ok && rel.Data != nil {
		details.ServiceSiteID = rel.Data.ID
		key := resourceKey(rel.Data.Type, rel.Data.ID)
		if inc, ok := included[key]; ok {
			details.ServiceSite = firstNonEmpty(
				stringAttr(inc.Attributes, "name"),
				stringAttr(inc.Attributes, "address"),
				rel.Data.ID,
			)
		}
	}

	// Requirement sets
	if rel, ok := resp.Data.Relationships["maintenance-requirement-sets"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				key := resourceKey(ref.Type, ref.ID)
				summary := workOrderSetSummary{ID: ref.ID}
				if inc, ok := included[key]; ok {
					summary.Name = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
					summary.Status = stringAttr(inc.Attributes, "status")
					summary.Type = stringAttr(inc.Attributes, "set-type")
				}
				details.RequirementSets = append(details.RequirementSets, summary)
			}
		}
	}

	// Assigned users
	if rel, ok := resp.Data.Relationships["assigned-users"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				key := resourceKey(ref.Type, ref.ID)
				if inc, ok := included[key]; ok {
					name := stringAttr(inc.Attributes, "name")
					if name != "" {
						details.AssignedUsers = append(details.AssignedUsers, name)
					} else {
						details.AssignedUsers = append(details.AssignedUsers, ref.ID)
					}
				} else {
					details.AssignedUsers = append(details.AssignedUsers, ref.ID)
				}
			}
		}
	}

	return details
}

func renderWorkOrderDetails(cmd *cobra.Command, d workOrderDetails) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", d.Status)
	}
	if d.Priority != "" {
		fmt.Fprintf(out, "Priority: %s\n", d.Priority)
	}
	if d.DueDate != "" {
		fmt.Fprintf(out, "Due Date: %s\n", d.DueDate)
	}
	if d.SafetyTagStatus != "" {
		fmt.Fprintf(out, "Safety Tag: %s\n", d.SafetyTagStatus)
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}
	if d.CompletedAt != "" {
		fmt.Fprintf(out, "Completed: %s\n", d.CompletedAt)
	}

	// Location
	if d.BusinessUnit != "" || d.ServiceSite != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Location:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.BusinessUnit != "" {
			fmt.Fprintf(out, "  Responsible Party: %s (ID: %s)\n", d.BusinessUnit, d.BusinessUnitID)
		}
		if d.ServiceSite != "" {
			fmt.Fprintf(out, "  Service Site: %s (ID: %s)\n", d.ServiceSite, d.ServiceSiteID)
		}
	}

	// Requirement sets
	if len(d.RequirementSets) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Requirement Sets (%d):\n", len(d.RequirementSets))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, s := range d.RequirementSets {
			name := s.Name
			if name == "" {
				name = s.ID
			}
			status := s.Status
			if status == "" {
				status = "-"
			}
			setType := s.Type
			if setType == "" {
				setType = "-"
			}
			fmt.Fprintf(out, "  - %s [%s] (%s)\n", name, status, setType)
		}
	}

	// Assigned users
	if len(d.AssignedUsers) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Assigned Users (%d):\n", len(d.AssignedUsers))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, u := range d.AssignedUsers {
			fmt.Fprintf(out, "  - %s\n", u)
		}
	}

	// Notes
	if d.Notes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Notes:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Notes)
	}

	return nil
}
