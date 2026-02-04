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

type timeSheetCostCodeAllocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetCostCodeAllocationCostCode struct {
	ID          string `json:"id"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

type timeSheetCostCodeAllocationDetails struct {
	ID                            string                                `json:"id"`
	TimeSheetID                   string                                `json:"time_sheet_id,omitempty"`
	Details                       []timeSheetCostCodeAllocationDetail   `json:"details,omitempty"`
	CostCodes                     []timeSheetCostCodeAllocationCostCode `json:"cost_codes,omitempty"`
	ProjectPhaseCostItemActualIDs []string                              `json:"project_phase_cost_item_actual_ids,omitempty"`
	CreatedAt                     string                                `json:"created_at,omitempty"`
	UpdatedAt                     string                                `json:"updated_at,omitempty"`
}

func newTimeSheetCostCodeAllocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet cost code allocation details",
		Long: `Show the full details of a time sheet cost code allocation.

Output Fields:
  ID             Allocation identifier
  Time Sheet     Time sheet ID
  Details        Cost code allocation details
  Cost Codes     Related cost codes (if included)
  Created At     Creation timestamp
  Updated At     Last update timestamp

Arguments:
  <id>    The time sheet cost code allocation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet cost code allocation
  xbe view time-sheet-cost-code-allocations show 123

  # Output as JSON
  xbe view time-sheet-cost-code-allocations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetCostCodeAllocationsShow,
	}
	initTimeSheetCostCodeAllocationsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetCostCodeAllocationsCmd.AddCommand(newTimeSheetCostCodeAllocationsShowCmd())
}

func initTimeSheetCostCodeAllocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetCostCodeAllocationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTimeSheetCostCodeAllocationsShowOptions(cmd)
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
		return fmt.Errorf("time sheet cost code allocation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-cost-code-allocations]", "details,created-at,updated-at,time-sheet,cost-codes")
	query.Set("fields[cost-codes]", "code,description")
	query.Set("include", "cost-codes")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-cost-code-allocations/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildTimeSheetCostCodeAllocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetCostCodeAllocationDetails(cmd, details)
}

func parseTimeSheetCostCodeAllocationsShowOptions(cmd *cobra.Command) (timeSheetCostCodeAllocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetCostCodeAllocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetCostCodeAllocationDetails(resp jsonAPISingleResponse) timeSheetCostCodeAllocationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := timeSheetCostCodeAllocationDetails{
		ID:        resource.ID,
		Details:   parseTimeSheetCostCodeAllocationDetails(attrs),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["cost-codes"]; ok && rel.raw != nil {
		for _, ref := range relationshipIDs(rel) {
			costCode := timeSheetCostCodeAllocationCostCode{ID: ref.ID}
			if inc, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
				costCode.Code = stringAttr(inc.Attributes, "code")
				costCode.Description = stringAttr(inc.Attributes, "description")
			}
			details.CostCodes = append(details.CostCodes, costCode)
		}
	}

	if rel, ok := resource.Relationships["project-phase-cost-item-actuals"]; ok && rel.raw != nil {
		details.ProjectPhaseCostItemActualIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderTimeSheetCostCodeAllocationDetails(cmd *cobra.Command, details timeSheetCostCodeAllocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet: %s\n", details.TimeSheetID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if len(details.Details) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Allocations:")
		for _, detail := range details.Details {
			line := fmt.Sprintf("- Cost Code: %s", detail.CostCodeID)
			if detail.Percentage != "" {
				line = fmt.Sprintf("%s | Percentage: %s", line, detail.Percentage)
			}
			if detail.ProjectCostClassificationID != "" {
				line = fmt.Sprintf("%s | Project Cost Classification: %s", line, detail.ProjectCostClassificationID)
			}
			fmt.Fprintln(out, line)
		}
	}

	if len(details.CostCodes) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Cost Codes:")
		for _, costCode := range details.CostCodes {
			label := costCode.ID
			if costCode.Code != "" {
				label = fmt.Sprintf("%s (%s)", costCode.Code, costCode.ID)
			}
			if costCode.Description != "" {
				label = fmt.Sprintf("%s - %s", label, costCode.Description)
			}
			fmt.Fprintf(out, "- %s\n", label)
		}
	}

	if len(details.ProjectPhaseCostItemActualIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Project Phase Cost Item Actuals: %s\n", strings.Join(details.ProjectPhaseCostItemActualIDs, ", "))
	}

	return nil
}
