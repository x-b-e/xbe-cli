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

type projectEstimateSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectEstimateSetDetails struct {
	ID                                         string   `json:"id"`
	Name                                       string   `json:"name,omitempty"`
	IsBid                                      bool     `json:"is_bid,omitempty"`
	IsActual                                   bool     `json:"is_actual,omitempty"`
	IsPossible                                 bool     `json:"is_possible,omitempty"`
	ProjectID                                  string   `json:"project_id,omitempty"`
	CreatedByID                                string   `json:"created_by_id,omitempty"`
	BackupEstimateSetID                        string   `json:"backup_estimate_set_id,omitempty"`
	ProjectRevenueItemPriceEstimateIDs         []string `json:"project_revenue_item_price_estimate_ids,omitempty"`
	ProjectRevenueItemQuantityEstimateIDs      []string `json:"project_revenue_item_quantity_estimate_ids,omitempty"`
	ProjectPhaseCostItemPriceEstimateIDs       []string `json:"project_phase_cost_item_price_estimate_ids,omitempty"`
	ProjectPhaseCostItemQuantityEstimateIDs    []string `json:"project_phase_cost_item_quantity_estimate_ids,omitempty"`
	ProjectPhaseRevenueItemQuantityEstimateIDs []string `json:"project_phase_revenue_item_quantity_estimate_ids,omitempty"`
	ProjectPhaseDatesEstimateIDs               []string `json:"project_phase_dates_estimate_ids,omitempty"`
	EffectiveEstimates                         any      `json:"effective_estimates,omitempty"`
}

func newProjectEstimateSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project estimate set details",
		Long: `Show the full details of a project estimate set.

Output Fields:
  ID                   Project estimate set identifier
  NAME                 Estimate set name
  PROJECT              Project ID
  CREATED BY           Created-by user ID
  BACKUP ESTIMATE SET  Backup estimate set ID
  BID                  Bid estimate set flag
  ACTUAL               Actual estimate set flag
  POSSIBLE             Possible estimate set flag
  REVENUE/COST ESTIMATES Related estimate IDs
  EFFECTIVE ESTIMATES  Effective estimates payload

Arguments:
  <id>  Project estimate set ID (required). Find IDs using the list command.`,
		Example: `  # View a project estimate set
  xbe view project-estimate-sets show 123

  # Get JSON output
  xbe view project-estimate-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectEstimateSetsShow,
	}
	initProjectEstimateSetsShowFlags(cmd)
	return cmd
}

func init() {
	projectEstimateSetsCmd.AddCommand(newProjectEstimateSetsShowCmd())
}

func initProjectEstimateSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectEstimateSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectEstimateSetsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project estimate set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-estimate-sets]", strings.Join([]string{
		"name",
		"is-bid",
		"is-actual",
		"is-possible",
		"effective-estimates",
		"project",
		"created-by",
		"backup-estimate-set",
		"project-revenue-item-price-estimates",
		"project-revenue-item-quantity-estimates",
		"project-phase-cost-item-price-estimates",
		"project-phase-cost-item-quantity-estimates",
		"project-phase-revenue-item-quantity-estimates",
		"project-phase-dates-estimates",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/project-estimate-sets/"+id, query)
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

	details := buildProjectEstimateSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectEstimateSetDetails(cmd, details)
}

func parseProjectEstimateSetsShowOptions(cmd *cobra.Command) (projectEstimateSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectEstimateSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectEstimateSetDetails(resp jsonAPISingleResponse) projectEstimateSetDetails {
	attrs := resp.Data.Attributes

	details := projectEstimateSetDetails{
		ID:                 resp.Data.ID,
		Name:               stringAttr(attrs, "name"),
		IsBid:              boolAttr(attrs, "is-bid"),
		IsActual:           boolAttr(attrs, "is-actual"),
		IsPossible:         boolAttr(attrs, "is-possible"),
		EffectiveEstimates: anyAttr(attrs, "effective-estimates"),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["backup-estimate-set"]; ok && rel.Data != nil {
		details.BackupEstimateSetID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-revenue-item-price-estimates"]; ok {
		details.ProjectRevenueItemPriceEstimateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-revenue-item-quantity-estimates"]; ok {
		details.ProjectRevenueItemQuantityEstimateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-phase-cost-item-price-estimates"]; ok {
		details.ProjectPhaseCostItemPriceEstimateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-phase-cost-item-quantity-estimates"]; ok {
		details.ProjectPhaseCostItemQuantityEstimateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-phase-revenue-item-quantity-estimates"]; ok {
		details.ProjectPhaseRevenueItemQuantityEstimateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-phase-dates-estimates"]; ok {
		details.ProjectPhaseDatesEstimateIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectEstimateSetDetails(cmd *cobra.Command, details projectEstimateSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project: %s\n", details.ProjectID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.BackupEstimateSetID != "" {
		fmt.Fprintf(out, "Backup Estimate Set: %s\n", details.BackupEstimateSetID)
	}
	fmt.Fprintf(out, "Bid: %t\n", details.IsBid)
	fmt.Fprintf(out, "Actual: %t\n", details.IsActual)
	fmt.Fprintf(out, "Possible: %t\n", details.IsPossible)

	if len(details.ProjectRevenueItemPriceEstimateIDs) > 0 {
		fmt.Fprintf(out, "Revenue Item Price Estimates: %s\n", strings.Join(details.ProjectRevenueItemPriceEstimateIDs, ", "))
	}
	if len(details.ProjectRevenueItemQuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Revenue Item Quantity Estimates: %s\n", strings.Join(details.ProjectRevenueItemQuantityEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseCostItemPriceEstimateIDs) > 0 {
		fmt.Fprintf(out, "Phase Cost Item Price Estimates: %s\n", strings.Join(details.ProjectPhaseCostItemPriceEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseCostItemQuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Phase Cost Item Quantity Estimates: %s\n", strings.Join(details.ProjectPhaseCostItemQuantityEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseRevenueItemQuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Phase Revenue Item Quantity Estimates: %s\n", strings.Join(details.ProjectPhaseRevenueItemQuantityEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseDatesEstimateIDs) > 0 {
		fmt.Fprintf(out, "Phase Dates Estimates: %s\n", strings.Join(details.ProjectPhaseDatesEstimateIDs, ", "))
	}

	if details.EffectiveEstimates != nil {
		if formatted := formatAnyJSON(details.EffectiveEstimates); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Effective Estimates:")
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
