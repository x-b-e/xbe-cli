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

type inventoryCapacitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type inventoryCapacityDetails struct {
	ID               string `json:"id"`
	MaterialSiteID   string `json:"material_site_id,omitempty"`
	MaterialSiteName string `json:"material_site_name,omitempty"`
	MaterialTypeID   string `json:"material_type_id,omitempty"`
	MaterialTypeName string `json:"material_type_name,omitempty"`
	MaxCapacityTons  string `json:"max_capacity_tons,omitempty"`
	MinCapacityTons  string `json:"min_capacity_tons,omitempty"`
	ThresholdTons    string `json:"threshold_tons,omitempty"`
}

func newInventoryCapacitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show inventory capacity details",
		Long: `Show the full details of a specific inventory capacity.

Output Fields:
  ID             Inventory capacity identifier
  Material Site  Material site name and ID
  Material Type  Material type name and ID
  Min Tons       Minimum capacity in tons
  Max Tons       Maximum capacity in tons
  Threshold      Alert threshold in tons

Arguments:
  <id>  The inventory capacity ID (required). You can find IDs using the list command.`,
		Example: `  # Show inventory capacity details
  xbe view inventory-capacities show 123

  # Output as JSON
  xbe view inventory-capacities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInventoryCapacitiesShow,
	}
	initInventoryCapacitiesShowFlags(cmd)
	return cmd
}

func init() {
	inventoryCapacitiesCmd.AddCommand(newInventoryCapacitiesShowCmd())
}

func initInventoryCapacitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryCapacitiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInventoryCapacitiesShowOptions(cmd)
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
		return fmt.Errorf("inventory capacity id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[inventory-capacities]", "max-capacity-tons,min-capacity-tons,threshold-tons,material-site,material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("include", "material-site,material-type")

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-capacities/"+id, query)
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

	details := buildInventoryCapacityDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInventoryCapacityDetails(cmd, details)
}

func parseInventoryCapacitiesShowOptions(cmd *cobra.Command) (inventoryCapacitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryCapacitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInventoryCapacityDetails(resp jsonAPISingleResponse) inventoryCapacityDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := inventoryCapacityDetails{
		ID:              resource.ID,
		MaxCapacityTons: strings.TrimSpace(stringAttr(attrs, "max-capacity-tons")),
		MinCapacityTons: strings.TrimSpace(stringAttr(attrs, "min-capacity-tons")),
		ThresholdTons:   strings.TrimSpace(stringAttr(attrs, "threshold-tons")),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSiteName = strings.TrimSpace(stringAttr(site.Attributes, "name"))
		}
	}
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = strings.TrimSpace(stringAttr(mt.Attributes, "name"))
		}
	}

	return details
}

func renderInventoryCapacityDetails(cmd *cobra.Command, details inventoryCapacityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialSiteName != "" || details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site: %s\n", formatInventoryCapacityLabel(details.MaterialSiteName, details.MaterialSiteID))
	}
	if details.MaterialTypeName != "" || details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type: %s\n", formatInventoryCapacityLabel(details.MaterialTypeName, details.MaterialTypeID))
	}
	if details.MinCapacityTons != "" {
		fmt.Fprintf(out, "Min Capacity (tons): %s\n", details.MinCapacityTons)
	}
	if details.MaxCapacityTons != "" {
		fmt.Fprintf(out, "Max Capacity (tons): %s\n", details.MaxCapacityTons)
	}
	if details.ThresholdTons != "" {
		fmt.Fprintf(out, "Threshold (tons): %s\n", details.ThresholdTons)
	}

	return nil
}
