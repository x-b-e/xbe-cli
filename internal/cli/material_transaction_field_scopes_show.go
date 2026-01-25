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

type materialTransactionFieldScopesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionFieldScopeRelation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type materialUnitOfMeasureQuantityDetail struct {
	ID              string `json:"id"`
	Quantity        string `json:"quantity,omitempty"`
	UnitOfMeasureID string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure   string `json:"unit_of_measure,omitempty"`
}

type materialTransactionFieldScopeDetails struct {
	ID                              string                                  `json:"id"`
	TicketNumber                    string                                  `json:"ticket_number,omitempty"`
	TransactionAt                   string                                  `json:"transaction_at,omitempty"`
	SourceNotes                     string                                  `json:"source_notes,omitempty"`
	TenderJobScheduleShiftIDs       []string                                `json:"tender_job_schedule_shift_ids,omitempty"`
	JobProductionPlans              []materialTransactionFieldScopeRelation `json:"job_production_plans,omitempty"`
	Origins                         []materialTransactionFieldScopeRelation `json:"origins,omitempty"`
	Destinations                    []materialTransactionFieldScopeRelation `json:"destinations,omitempty"`
	JobSites                        []materialTransactionFieldScopeRelation `json:"job_sites,omitempty"`
	MaterialSites                   []materialTransactionFieldScopeRelation `json:"material_sites,omitempty"`
	MaterialTypes                   []materialTransactionFieldScopeRelation `json:"material_types,omitempty"`
	UnitOfMeasures                  []materialTransactionFieldScopeRelation `json:"unit_of_measures,omitempty"`
	MaterialUnitOfMeasureQuantities []materialUnitOfMeasureQuantityDetail   `json:"material_unit_of_measure_quantities,omitempty"`
}

func newMaterialTransactionFieldScopesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <material-transaction-id>",
		Short: "Show material transaction field scope details",
		Long: `Show details for a material transaction field scope.

Field scopes provide the matching context for a material transaction, including
related job/material selections and derived ticket data.

Output Fields:
  ID                            Material transaction ID
  Ticket Number                 Ticket number
  Transaction At                Transaction timestamp
  Source Notes                  Source data notes
  Tender Job Schedule Shifts    Related tender job schedule shift IDs
  Job Production Plans          Related job production plans
  Origins/Destinations          Matched origin/destination records
  Job Sites                     Related job sites
  Material Sites                Related material sites
  Material Types                Related material types
  Unit Of Measures              Related unit of measures
  Material UOM Quantities       Quantities by unit of measure

Arguments:
  <material-transaction-id>  The material transaction ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show field scope details for a material transaction
  xbe view material-transaction-field-scopes show 123

  # Output as JSON
  xbe view material-transaction-field-scopes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionFieldScopesShow,
	}
	initMaterialTransactionFieldScopesShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionFieldScopesCmd.AddCommand(newMaterialTransactionFieldScopesShowCmd())
}

func initMaterialTransactionFieldScopesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionFieldScopesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionFieldScopesShowOptions(cmd)
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
		return fmt.Errorf("material transaction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-field-scopes]", "ticket-number,transaction-at,source-notes,tender-job-schedule-shifts,job-production-plans,origins,destinations,job-sites,material-sites,material-types,unit-of-measures,material-unit-of-measure-quantities")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[material-unit-of-measure-quantities]", "quantity,unit-of-measure")
	query.Set("include", "tender-job-schedule-shifts,job-production-plans,origins,destinations,job-sites,material-sites,material-types,unit-of-measures,material-unit-of-measure-quantities.unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-field-scopes/"+id, query)
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

	details := buildMaterialTransactionFieldScopeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionFieldScopeDetails(cmd, details)
}

func parseMaterialTransactionFieldScopesShowOptions(cmd *cobra.Command) (materialTransactionFieldScopesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionFieldScopesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionFieldScopeDetails(resp jsonAPISingleResponse) materialTransactionFieldScopeDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialTransactionFieldScopeDetails{
		ID:            resource.ID,
		TicketNumber:  stringAttr(attrs, "ticket-number"),
		TransactionAt: formatDateTime(stringAttr(attrs, "transaction-at")),
		SourceNotes:   stringAttr(attrs, "source-notes"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["job-production-plans"]; ok {
		details.JobProductionPlans = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return firstNonEmpty(stringAttr(res.Attributes, "job-number"), stringAttr(res.Attributes, "job-name"), stringAttr(res.Attributes, "name"))
		})
	}
	if rel, ok := resource.Relationships["origins"]; ok {
		details.Origins = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return firstNonEmpty(stringAttr(res.Attributes, "name"), stringAttr(res.Attributes, "title"))
		})
	}
	if rel, ok := resource.Relationships["destinations"]; ok {
		details.Destinations = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return firstNonEmpty(stringAttr(res.Attributes, "name"), stringAttr(res.Attributes, "title"))
		})
	}
	if rel, ok := resource.Relationships["job-sites"]; ok {
		details.JobSites = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return stringAttr(res.Attributes, "name")
		})
	}
	if rel, ok := resource.Relationships["material-sites"]; ok {
		details.MaterialSites = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return stringAttr(res.Attributes, "name")
		})
	}
	if rel, ok := resource.Relationships["material-types"]; ok {
		details.MaterialTypes = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return stringAttr(res.Attributes, "name")
		})
	}
	if rel, ok := resource.Relationships["unit-of-measures"]; ok {
		details.UnitOfMeasures = buildMaterialTransactionFieldScopeRelations(rel, included, func(res jsonAPIResource) string {
			return firstNonEmpty(stringAttr(res.Attributes, "abbreviation"), stringAttr(res.Attributes, "name"))
		})
	}
	if rel, ok := resource.Relationships["material-unit-of-measure-quantities"]; ok {
		details.MaterialUnitOfMeasureQuantities = buildMaterialUnitOfMeasureQuantityDetails(rel, included)
	}

	return details
}

func buildMaterialTransactionFieldScopeRelations(rel jsonAPIRelationship, included map[string]jsonAPIResource, nameFn func(jsonAPIResource) string) []materialTransactionFieldScopeRelation {
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	out := make([]materialTransactionFieldScopeRelation, 0, len(ids))
	for _, id := range ids {
		relation := materialTransactionFieldScopeRelation{ID: id.ID, Type: id.Type}
		if res, ok := included[resourceKey(id.Type, id.ID)]; ok && nameFn != nil {
			if name := nameFn(res); name != "" {
				relation.Name = name
			}
		}
		out = append(out, relation)
	}
	return out
}

func buildMaterialUnitOfMeasureQuantityDetails(rel jsonAPIRelationship, included map[string]jsonAPIResource) []materialUnitOfMeasureQuantityDetail {
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	out := make([]materialUnitOfMeasureQuantityDetail, 0, len(ids))
	for _, id := range ids {
		quantityDetail := materialUnitOfMeasureQuantityDetail{ID: id.ID}
		if res, ok := included[resourceKey(id.Type, id.ID)]; ok {
			quantityDetail.Quantity = stringAttr(res.Attributes, "quantity")
			if uomRel, ok := res.Relationships["unit-of-measure"]; ok && uomRel.Data != nil {
				quantityDetail.UnitOfMeasureID = uomRel.Data.ID
				if uomRes, ok := included[resourceKey(uomRel.Data.Type, uomRel.Data.ID)]; ok {
					quantityDetail.UnitOfMeasure = firstNonEmpty(stringAttr(uomRes.Attributes, "abbreviation"), stringAttr(uomRes.Attributes, "name"))
				}
			}
		}
		out = append(out, quantityDetail)
	}
	return out
}

func renderMaterialTransactionFieldScopeDetails(cmd *cobra.Command, details materialTransactionFieldScopeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TicketNumber != "" {
		fmt.Fprintf(out, "Ticket Number: %s\n", details.TicketNumber)
	}
	if details.TransactionAt != "" {
		fmt.Fprintf(out, "Transaction At: %s\n", details.TransactionAt)
	}
	if details.SourceNotes != "" {
		fmt.Fprintf(out, "Source Notes: %s\n", details.SourceNotes)
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shifts: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}

	writeScopeRelations(out, "Job Production Plans", details.JobProductionPlans)
	writeScopeRelations(out, "Origins", details.Origins)
	writeScopeRelations(out, "Destinations", details.Destinations)
	writeScopeRelations(out, "Job Sites", details.JobSites)
	writeScopeRelations(out, "Material Sites", details.MaterialSites)
	writeScopeRelations(out, "Material Types", details.MaterialTypes)
	writeScopeRelations(out, "Unit Of Measures", details.UnitOfMeasures)

	if len(details.MaterialUnitOfMeasureQuantities) > 0 {
		fmt.Fprintln(out, "Material UOM Quantities:")
		for _, quantity := range details.MaterialUnitOfMeasureQuantities {
			label := quantity.ID
			if quantity.UnitOfMeasure != "" {
				label = fmt.Sprintf("%s (%s)", quantity.UnitOfMeasure, quantity.ID)
			}
			if quantity.Quantity != "" {
				fmt.Fprintf(out, "  %s: %s\n", label, quantity.Quantity)
			} else {
				fmt.Fprintf(out, "  %s\n", label)
			}
		}
	}

	return nil
}

func writeScopeRelations(outWriter interface{ Write([]byte) (int, error) }, label string, items []materialTransactionFieldScopeRelation) {
	if len(items) == 0 {
		return
	}
	formatted := make([]string, 0, len(items))
	for _, item := range items {
		formatted = append(formatted, formatScopeRelation(item))
	}
	fmt.Fprintf(outWriter, "%s: %s\n", label, strings.Join(formatted, ", "))
}

func formatScopeRelation(item materialTransactionFieldScopeRelation) string {
	if item.Name == "" {
		return fmt.Sprintf("%s/%s", item.Type, item.ID)
	}
	return fmt.Sprintf("%s (%s/%s)", item.Name, item.Type, item.ID)
}
