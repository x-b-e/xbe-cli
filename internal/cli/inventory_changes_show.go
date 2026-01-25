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

type inventoryChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type inventoryChangeDetails struct {
	ID                                     string `json:"id"`
	MaterialSiteID                         string `json:"material_site_id,omitempty"`
	MaterialSite                           string `json:"material_site,omitempty"`
	MaterialTypeID                         string `json:"material_type_id,omitempty"`
	MaterialType                           string `json:"material_type,omitempty"`
	MostRecentInventoryEstimateID          string `json:"most_recent_inventory_estimate_id,omitempty"`
	MostRecentInventoryEstimateAt          string `json:"most_recent_inventory_estimate_at,omitempty"`
	MostRecentInventoryEstimateAmountTons  string `json:"most_recent_inventory_estimate_amount_tons,omitempty"`
	MostRecentInventoryEstimateDescription string `json:"most_recent_inventory_estimate_description,omitempty"`
	EstimateAt                             string `json:"estimate_at,omitempty"`
	ForecastStartAt                        string `json:"forecast_start_at,omitempty"`
	CalculatedAt                           string `json:"calculated_at,omitempty"`
	StartingAmountTons                     string `json:"starting_amount_tons,omitempty"`
	EndingAmountTons                       string `json:"ending_amount_tons,omitempty"`
	Details                                any    `json:"details,omitempty"`
}

func newInventoryChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show inventory change details",
		Long: `Show the full details of a specific inventory change.

Includes calculation metadata, inventory amounts, and detailed change entries.

Arguments:
  <id>  Inventory change ID (required). Use the list command to find IDs.`,
		Example: `  # Show an inventory change
  xbe view inventory-changes show 123

  # JSON output
  xbe view inventory-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInventoryChangesShow,
	}
	initInventoryChangesShowFlags(cmd)
	return cmd
}

func init() {
	inventoryChangesCmd.AddCommand(newInventoryChangesShowCmd())
}

func initInventoryChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInventoryChangesShowOptions(cmd)
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
		return fmt.Errorf("inventory change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[inventory-changes]", "estimate-at,forecast-start-at,calculated-at,details,starting-amount-tons,ending-amount-tons,material-site,material-type,most-recent-inventory-estimate")
	query.Set("fields[material-sites]", "name,material-supplier")
	query.Set("fields[material-types]", "name")
	query.Set("fields[inventory-estimates]", "estimated-at,amount-tons,description")
	query.Set("include", "material-site,material-type,most-recent-inventory-estimate")

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-changes/"+id, query)
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

	details := buildInventoryChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInventoryChangeDetails(cmd, details)
}

func parseInventoryChangesShowOptions(cmd *cobra.Command) (inventoryChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInventoryChangeDetails(resp jsonAPISingleResponse) inventoryChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := inventoryChangeDetails{
		ID:                 resource.ID,
		EstimateAt:         formatDateTime(stringAttr(attrs, "estimate-at")),
		ForecastStartAt:    formatDateTime(stringAttr(attrs, "forecast-start-at")),
		CalculatedAt:       formatDateTime(stringAttr(attrs, "calculated-at")),
		StartingAmountTons: stringAttr(attrs, "starting-amount-tons"),
		EndingAmountTons:   stringAttr(attrs, "ending-amount-tons"),
		Details:            attrs["details"],
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(ms.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = stringAttr(mt.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["most-recent-inventory-estimate"]; ok && rel.Data != nil {
		details.MostRecentInventoryEstimateID = rel.Data.ID
		if estimate, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MostRecentInventoryEstimateAt = formatDateTime(stringAttr(estimate.Attributes, "estimated-at"))
			details.MostRecentInventoryEstimateAmountTons = stringAttr(estimate.Attributes, "amount-tons")
			details.MostRecentInventoryEstimateDescription = stringAttr(estimate.Attributes, "description")
		}
	}

	return details
}

func renderInventoryChangeDetails(cmd *cobra.Command, details inventoryChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialSite != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSite)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}
	if details.MaterialType != "" {
		fmt.Fprintf(out, "Material Type: %s\n", details.MaterialType)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
	}
	if details.MostRecentInventoryEstimateID != "" {
		fmt.Fprintf(out, "Most Recent Inventory Estimate ID: %s\n", details.MostRecentInventoryEstimateID)
	}
	if details.MostRecentInventoryEstimateAt != "" {
		fmt.Fprintf(out, "Most Recent Inventory Estimate At: %s\n", details.MostRecentInventoryEstimateAt)
	}
	if details.MostRecentInventoryEstimateAmountTons != "" {
		fmt.Fprintf(out, "Most Recent Inventory Estimate Amount: %s tons\n", details.MostRecentInventoryEstimateAmountTons)
	}
	if details.MostRecentInventoryEstimateDescription != "" {
		fmt.Fprintf(out, "Most Recent Inventory Estimate Description: %s\n", details.MostRecentInventoryEstimateDescription)
	}
	if details.EstimateAt != "" {
		fmt.Fprintf(out, "Estimate At: %s\n", details.EstimateAt)
	}
	if details.ForecastStartAt != "" {
		fmt.Fprintf(out, "Forecast Start At: %s\n", details.ForecastStartAt)
	}
	if details.CalculatedAt != "" {
		fmt.Fprintf(out, "Calculated At: %s\n", details.CalculatedAt)
	}
	if details.StartingAmountTons != "" {
		fmt.Fprintf(out, "Starting Amount: %s tons\n", details.StartingAmountTons)
	}
	if details.EndingAmountTons != "" {
		fmt.Fprintf(out, "Ending Amount: %s tons\n", details.EndingAmountTons)
	}

	if details.Details != nil {
		prettyDetails := formatJSONValue(details.Details)
		if prettyDetails != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, prettyDetails)
		}
	}

	return nil
}
