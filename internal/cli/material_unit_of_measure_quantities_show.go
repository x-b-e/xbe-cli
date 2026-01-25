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

type materialUnitOfMeasureQuantitiesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialUnitOfMeasureQuantityDetails struct {
	ID                    string `json:"id"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
	UnitOfMeasure         string `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID       string `json:"unit_of_measure_id,omitempty"`
	Quantity              string `json:"quantity,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newMaterialUnitOfMeasureQuantitiesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material unit of measure quantity details",
		Long: `Show the full details of a material unit of measure quantity.

Output Fields:
  ID                     Quantity identifier
  Material Transaction   Material transaction ID
  Quantity               Quantity recorded
  Unit Of Measure        Unit of measure
  Created At             Created timestamp
  Updated At             Updated timestamp

Arguments:
  <id>    The material unit of measure quantity ID (required). You can find IDs using the list command.`,
		Example: `  # Show a material unit of measure quantity
  xbe view material-unit-of-measure-quantities show 123

  # Get JSON output
  xbe view material-unit-of-measure-quantities show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialUnitOfMeasureQuantitiesShow,
	}
	initMaterialUnitOfMeasureQuantitiesShowFlags(cmd)
	return cmd
}

func init() {
	materialUnitOfMeasureQuantitiesCmd.AddCommand(newMaterialUnitOfMeasureQuantitiesShowCmd())
}

func initMaterialUnitOfMeasureQuantitiesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialUnitOfMeasureQuantitiesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialUnitOfMeasureQuantitiesShowOptions(cmd)
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
		return fmt.Errorf("material unit of measure quantity id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-unit-of-measure-quantities]", "quantity,created-at,updated-at,material-transaction,unit-of-measure")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/material-unit-of-measure-quantities/"+id, query)
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

	details := buildMaterialUnitOfMeasureQuantityShowDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialUnitOfMeasureQuantityDetails(cmd, details)
}

func parseMaterialUnitOfMeasureQuantitiesShowOptions(cmd *cobra.Command) (materialUnitOfMeasureQuantitiesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialUnitOfMeasureQuantitiesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialUnitOfMeasureQuantityShowDetails(resp jsonAPISingleResponse) materialUnitOfMeasureQuantityDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialUnitOfMeasureQuantityDetails{
		ID:        resource.ID,
		Quantity:  strings.TrimSpace(stringAttr(attrs, "quantity")),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = unitOfMeasureLabel(uom.Attributes)
		}
	}

	return details
}

func renderMaterialUnitOfMeasureQuantityDetails(cmd *cobra.Command, details materialUnitOfMeasureQuantityDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction: %s\n", details.MaterialTransactionID)
	}
	if details.Quantity != "" {
		if details.UnitOfMeasure != "" {
			fmt.Fprintf(out, "Quantity: %s %s\n", details.Quantity, details.UnitOfMeasure)
		} else {
			fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
		}
	}
	if details.UnitOfMeasureID != "" {
		if details.UnitOfMeasure != "" {
			fmt.Fprintf(out, "Unit Of Measure: %s (%s)\n", details.UnitOfMeasure, details.UnitOfMeasureID)
		} else {
			fmt.Fprintf(out, "Unit Of Measure: %s\n", details.UnitOfMeasureID)
		}
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
