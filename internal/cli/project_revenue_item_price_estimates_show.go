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

type projectRevenueItemPriceEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectRevenueItemPriceEstimateDetails struct {
	ID                    string `json:"id"`
	Kind                  string `json:"kind,omitempty"`
	PricePerUnitExplicit  string `json:"price_per_unit_explicit,omitempty"`
	CostMultiplier        string `json:"cost_multiplier,omitempty"`
	PricePerUnit          string `json:"price_per_unit,omitempty"`
	PricePerUnitEffective string `json:"price_per_unit_effective,omitempty"`
	ProjectRevenueItemID  string `json:"project_revenue_item_id,omitempty"`
	ProjectEstimateSetID  string `json:"project_estimate_set_id,omitempty"`
	CreatedByID           string `json:"created_by_id,omitempty"`
}

func newProjectRevenueItemPriceEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project revenue item price estimate details",
		Long: `Show the full details of a project revenue item price estimate.

Output Fields:
  ID                         Price estimate identifier
  Project Revenue Item ID    Related project revenue item
  Project Estimate Set ID    Related project estimate set
  Created By ID              Created-by user ID
  Kind                       Estimate kind (explicit or cost_multiplier)
  Price Per Unit (Explicit)  Explicit price per unit
  Cost Multiplier            Cost multiplier
  Price Per Unit             Resolved price per unit
  Price Per Unit (Effective) Effective price per unit

Arguments:
  <id>    Project revenue item price estimate ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project revenue item price estimate
  xbe view project-revenue-item-price-estimates show 123

  # JSON output
  xbe view project-revenue-item-price-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectRevenueItemPriceEstimatesShow,
	}
	initProjectRevenueItemPriceEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	projectRevenueItemPriceEstimatesCmd.AddCommand(newProjectRevenueItemPriceEstimatesShowCmd())
}

func initProjectRevenueItemPriceEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectRevenueItemPriceEstimatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectRevenueItemPriceEstimatesShowOptions(cmd)
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
		return fmt.Errorf("project revenue item price estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-revenue-item-price-estimates]", "kind,price-per-unit,price-per-unit-explicit,cost-multiplier,price-per-unit-effective,project-revenue-item,project-estimate-set,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-revenue-item-price-estimates/"+id, query)
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

	details := buildProjectRevenueItemPriceEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectRevenueItemPriceEstimateDetails(cmd, details)
}

func parseProjectRevenueItemPriceEstimatesShowOptions(cmd *cobra.Command) (projectRevenueItemPriceEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectRevenueItemPriceEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectRevenueItemPriceEstimateDetails(resp jsonAPISingleResponse) projectRevenueItemPriceEstimateDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := projectRevenueItemPriceEstimateDetails{
		ID:                    resource.ID,
		Kind:                  stringAttr(attrs, "kind"),
		PricePerUnitExplicit:  stringAttr(attrs, "price-per-unit-explicit"),
		CostMultiplier:        stringAttr(attrs, "cost-multiplier"),
		PricePerUnit:          stringAttr(attrs, "price-per-unit"),
		PricePerUnitEffective: stringAttr(attrs, "price-per-unit-effective"),
	}

	if rel, ok := resource.Relationships["project-revenue-item"]; ok && rel.Data != nil {
		details.ProjectRevenueItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-estimate-set"]; ok && rel.Data != nil {
		details.ProjectEstimateSetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderProjectRevenueItemPriceEstimateDetails(cmd *cobra.Command, details projectRevenueItemPriceEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectRevenueItemID != "" {
		fmt.Fprintf(out, "Project Revenue Item ID: %s\n", details.ProjectRevenueItemID)
	}
	if details.ProjectEstimateSetID != "" {
		fmt.Fprintf(out, "Project Estimate Set ID: %s\n", details.ProjectEstimateSetID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.PricePerUnitExplicit != "" {
		fmt.Fprintf(out, "Price Per Unit (Explicit): %s\n", details.PricePerUnitExplicit)
	}
	if details.CostMultiplier != "" {
		fmt.Fprintf(out, "Cost Multiplier: %s\n", details.CostMultiplier)
	}
	if details.PricePerUnit != "" {
		fmt.Fprintf(out, "Price Per Unit: %s\n", details.PricePerUnit)
	}
	if details.PricePerUnitEffective != "" {
		fmt.Fprintf(out, "Price Per Unit (Effective): %s\n", details.PricePerUnitEffective)
	}

	return nil
}
