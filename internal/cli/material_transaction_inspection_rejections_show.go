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

type materialTransactionInspectionRejectionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionInspectionRejectionDetails struct {
	ID                              string `json:"id"`
	MaterialTransactionInspectionID string `json:"material_transaction_inspection_id,omitempty"`
	UnitOfMeasure                   string `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID                 string `json:"unit_of_measure_id,omitempty"`
	Quantity                        string `json:"quantity,omitempty"`
	Note                            string `json:"note,omitempty"`
	RejectedByName                  string `json:"rejected_by_name,omitempty"`
	CreatedAt                       string `json:"created_at,omitempty"`
	UpdatedAt                       string `json:"updated_at,omitempty"`
}

func newMaterialTransactionInspectionRejectionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction inspection rejection details",
		Long: `Show the full details of a material transaction inspection rejection.

Output Fields:
  ID                              Rejection identifier
  Material Transaction Inspection Material transaction inspection ID
  Quantity                        Rejected quantity
  Unit Of Measure                 Unit of measure
  Note                            Rejection note
  Rejected By                     Rejected by name (when available)
  Created At                      Created timestamp
  Updated At                      Updated timestamp

Arguments:
  <id>    The material transaction inspection rejection ID (required). You can find IDs using the list command.`,
		Example: `  # Show a rejection
  xbe view material-transaction-inspection-rejections show 123

  # Get JSON output
  xbe view material-transaction-inspection-rejections show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionInspectionRejectionsShow,
	}
	initMaterialTransactionInspectionRejectionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionInspectionRejectionsCmd.AddCommand(newMaterialTransactionInspectionRejectionsShowCmd())
}

func initMaterialTransactionInspectionRejectionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionInspectionRejectionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionInspectionRejectionsShowOptions(cmd)
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
		return fmt.Errorf("material transaction inspection rejection id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-inspection-rejections]", "quantity,note,rejected-by-name,created-at,updated-at,material-transaction-inspection,unit-of-measure")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-inspection-rejections/"+id, query)
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

	details := buildMaterialTransactionInspectionRejectionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionInspectionRejectionDetails(cmd, details)
}

func parseMaterialTransactionInspectionRejectionsShowOptions(cmd *cobra.Command) (materialTransactionInspectionRejectionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionInspectionRejectionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionInspectionRejectionDetails(resp jsonAPISingleResponse) materialTransactionInspectionRejectionDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialTransactionInspectionRejectionDetails{
		ID:             resource.ID,
		Quantity:       strings.TrimSpace(stringAttr(attrs, "quantity")),
		Note:           stringAttr(attrs, "note"),
		RejectedByName: stringAttr(attrs, "rejected-by-name"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["material-transaction-inspection"]; ok && rel.Data != nil {
		details.MaterialTransactionInspectionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = unitOfMeasureLabel(uom.Attributes)
		}
	}

	return details
}

func renderMaterialTransactionInspectionRejectionDetails(cmd *cobra.Command, details materialTransactionInspectionRejectionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTransactionInspectionID != "" {
		fmt.Fprintf(out, "Material Transaction Inspection: %s\n", details.MaterialTransactionInspectionID)
	}
	if details.Quantity != "" {
		if details.UnitOfMeasure != "" {
			fmt.Fprintf(out, "Quantity: %s %s\n", details.Quantity, details.UnitOfMeasure)
		} else {
			fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
		}
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.RejectedByName != "" {
		fmt.Fprintf(out, "Rejected By: %s\n", details.RejectedByName)
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
