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

type inventoryEstimatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type inventoryEstimateDetails struct {
	ID                          string `json:"id"`
	EstimatedAt                 string `json:"estimated_at,omitempty"`
	AmountTons                  string `json:"amount_tons,omitempty"`
	Description                 string `json:"description,omitempty"`
	MaterialSiteID              string `json:"material_site_id,omitempty"`
	MaterialTypeID              string `json:"material_type_id,omitempty"`
	MaterialSupplierID          string `json:"material_supplier_id,omitempty"`
	BrokerID                    string `json:"broker_id,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
	MostRecentInventoryChangeID string `json:"most_recent_inventory_change_id,omitempty"`
}

func newInventoryEstimatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show inventory estimate details",
		Long: `Show the full details of a specific inventory estimate.

Output Fields:
  ID            Inventory estimate identifier
  Estimated At  Estimated timestamp
  Amount Tons   Estimated amount (tons)
  Description   Description
  Material Site Material site ID
  Material Type Material type ID
  Supplier      Material supplier ID
  Broker        Broker ID
  Created By    Created-by user ID
  Most Recent Inventory Change Most recent inventory change ID

Arguments:
  <id>    The inventory estimate ID (required). You can find IDs using the list command.`,
		Example: `  # Show an inventory estimate
  xbe view inventory-estimates show 123

  # Get JSON output
  xbe view inventory-estimates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInventoryEstimatesShow,
	}
	initInventoryEstimatesShowFlags(cmd)
	return cmd
}

func init() {
	inventoryEstimatesCmd.AddCommand(newInventoryEstimatesShowCmd())
}

func initInventoryEstimatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryEstimatesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseInventoryEstimatesShowOptions(cmd)
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
		return fmt.Errorf("inventory estimate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[inventory-estimates]", "estimated-at,amount-tons,description,material-site,material-type,material-supplier,broker,created-by,most-recent-inventory-change")

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-estimates/"+id, query)
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

	details := buildInventoryEstimateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInventoryEstimateDetails(cmd, details)
}

func parseInventoryEstimatesShowOptions(cmd *cobra.Command) (inventoryEstimatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryEstimatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInventoryEstimateDetails(resp jsonAPISingleResponse) inventoryEstimateDetails {
	attrs := resp.Data.Attributes

	details := inventoryEstimateDetails{
		ID:          resp.Data.ID,
		EstimatedAt: formatDateTime(stringAttr(attrs, "estimated-at")),
		AmountTons:  stringAttr(attrs, "amount-tons"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["most-recent-inventory-change"]; ok && rel.Data != nil {
		details.MostRecentInventoryChangeID = rel.Data.ID
	}

	return details
}

func renderInventoryEstimateDetails(cmd *cobra.Command, details inventoryEstimateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EstimatedAt != "" {
		fmt.Fprintf(out, "Estimated At: %s\n", details.EstimatedAt)
	}
	if details.AmountTons != "" {
		fmt.Fprintf(out, "Amount Tons: %s\n", details.AmountTons)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSiteID)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type: %s\n", details.MaterialTypeID)
	}
	if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "Material Supplier: %s\n", details.MaterialSupplierID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.MostRecentInventoryChangeID != "" {
		fmt.Fprintf(out, "Most Recent Inventory Change: %s\n", details.MostRecentInventoryChangeID)
	}

	return nil
}
