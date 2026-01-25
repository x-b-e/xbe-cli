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

type materialTransactionInspectionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionInspectionDetails struct {
	ID                    string                                   `json:"id"`
	Status                string                                   `json:"status,omitempty"`
	Strategy              string                                   `json:"strategy,omitempty"`
	Note                  string                                   `json:"note,omitempty"`
	ChangedByName         string                                   `json:"changed_by_name,omitempty"`
	ChangedByID           string                                   `json:"changed_by_id,omitempty"`
	CreatedByName         string                                   `json:"created_by_name,omitempty"`
	CreatedByID           string                                   `json:"created_by_id,omitempty"`
	MaterialTransactionID string                                   `json:"material_transaction_id,omitempty"`
	TripID                string                                   `json:"trip_id,omitempty"`
	TenderJobScheduleID   string                                   `json:"tender_job_schedule_shift_id,omitempty"`
	Customer              string                                   `json:"customer,omitempty"`
	CustomerID            string                                   `json:"customer_id,omitempty"`
	Broker                string                                   `json:"broker,omitempty"`
	BrokerID              string                                   `json:"broker_id,omitempty"`
	MaterialSupplier      string                                   `json:"material_supplier,omitempty"`
	MaterialSupplierID    string                                   `json:"material_supplier_id,omitempty"`
	JobProductionPlan     string                                   `json:"job_production_plan,omitempty"`
	JobProductionPlanID   string                                   `json:"job_production_plan_id,omitempty"`
	DeliverySite          string                                   `json:"delivery_site,omitempty"`
	DeliverySiteID        string                                   `json:"delivery_site_id,omitempty"`
	DeliverySiteType      string                                   `json:"delivery_site_type,omitempty"`
	TonsAccepted          float64                                  `json:"tons_accepted,omitempty"`
	TonsRejected          float64                                  `json:"tons_rejected,omitempty"`
	Rejections            []materialTransactionInspectionRejection `json:"rejections,omitempty"`
}

type materialTransactionInspectionRejection struct {
	ID              string  `json:"id"`
	Quantity        float64 `json:"quantity,omitempty"`
	UnitOfMeasure   string  `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID string  `json:"unit_of_measure_id,omitempty"`
	Note            string  `json:"note,omitempty"`
	RejectedByName  string  `json:"rejected_by_name,omitempty"`
}

func newMaterialTransactionInspectionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction inspection details",
		Long: `Show the full details of a material transaction inspection.

Includes inspection status, strategy, related transaction, and any rejection
details tied to the inspection.

Output Fields:
  ID                  Inspection identifier
  Status              Inspection status
  Strategy            Inspection strategy
  Note                Inspection note
  Material Transaction ID  Related material transaction
  Delivery Site       Delivery site name and type
  Changed By          User who last updated the inspection
  Created By          User who created the inspection
  Tons Accepted       Calculated accepted tonnage
  Tons Rejected       Calculated rejected tonnage
  Rejections          List of inspection rejections (if any)

Arguments:
  <id>    The inspection ID (required). Use the list command to find IDs.`,
		Example: `  # Show an inspection
  xbe view material-transaction-inspections show 123

  # Get JSON output
  xbe view material-transaction-inspections show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionInspectionsShow,
	}
	initMaterialTransactionInspectionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionInspectionsCmd.AddCommand(newMaterialTransactionInspectionsShowCmd())
}

func initMaterialTransactionInspectionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionInspectionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionInspectionsShowOptions(cmd)
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
		return fmt.Errorf("material transaction inspection id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-inspections]", "note,status,strategy,changed-by-name,tons-accepted,tons-rejected,material-transaction-id,material-transaction,delivery-site,changed-by,created-by,trip,tender-job-schedule-shift,customer,broker,material-supplier,job-production-plan,material-transaction-inspection-rejections")
	query.Set("fields[material-transaction-inspection-rejections]", "quantity,note,rejected-by-name,unit-of-measure")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[users]", "name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name,company-name")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("include", "material-transaction,delivery-site,changed-by,created-by,trip,tender-job-schedule-shift,customer,broker,material-supplier,job-production-plan,material-transaction-inspection-rejections,material-transaction-inspection-rejections.unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-inspections/"+id, query)
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

	details := buildMaterialTransactionInspectionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionInspectionDetails(cmd, details)
}

func parseMaterialTransactionInspectionsShowOptions(cmd *cobra.Command) (materialTransactionInspectionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionInspectionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionInspectionDetails(resp jsonAPISingleResponse) materialTransactionInspectionDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := materialTransactionInspectionDetails{
		ID:                    resp.Data.ID,
		Status:                stringAttr(attrs, "status"),
		Strategy:              stringAttr(attrs, "strategy"),
		Note:                  stringAttr(attrs, "note"),
		ChangedByName:         stringAttr(attrs, "changed-by-name"),
		TonsAccepted:          floatAttr(attrs, "tons-accepted"),
		TonsRejected:          floatAttr(attrs, "tons-rejected"),
		MaterialTransactionID: stringAttr(attrs, "material-transaction-id"),
	}

	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
		if details.ChangedByName == "" {
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				details.ChangedByName = stringAttr(user.Attributes, "name")
			}
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["trip"]; ok && rel.Data != nil {
		details.TripID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Customer = firstNonEmpty(
				stringAttr(customer.Attributes, "company-name"),
				stringAttr(customer.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Broker = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSupplier = firstNonEmpty(
				stringAttr(supplier.Attributes, "name"),
				stringAttr(supplier.Attributes, "company-name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			jobNumber := stringAttr(plan.Attributes, "job-number")
			jobName := stringAttr(plan.Attributes, "job-name")
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["delivery-site"]; ok && rel.Data != nil {
		details.DeliverySiteType = rel.Data.Type
		details.DeliverySiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DeliverySite = stringAttr(site.Attributes, "name")
		}
	}

	for _, inc := range resp.Included {
		if inc.Type != "material-transaction-inspection-rejections" {
			continue
		}

		rejection := materialTransactionInspectionRejection{
			ID:             inc.ID,
			Quantity:       floatAttr(inc.Attributes, "quantity"),
			Note:           stringAttr(inc.Attributes, "note"),
			RejectedByName: stringAttr(inc.Attributes, "rejected-by-name"),
		}

		if rel, ok := inc.Relationships["unit-of-measure"]; ok && rel.Data != nil {
			rejection.UnitOfMeasureID = rel.Data.ID
			if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				name := stringAttr(uom.Attributes, "name")
				abbr := stringAttr(uom.Attributes, "abbreviation")
				if name != "" && abbr != "" {
					rejection.UnitOfMeasure = fmt.Sprintf("%s (%s)", name, abbr)
				} else {
					rejection.UnitOfMeasure = firstNonEmpty(name, abbr)
				}
			}
		}

		details.Rejections = append(details.Rejections, rejection)
	}

	return details
}

func renderMaterialTransactionInspectionDetails(cmd *cobra.Command, details materialTransactionInspectionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Strategy != "" {
		fmt.Fprintf(out, "Strategy: %s\n", details.Strategy)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction ID: %s\n", details.MaterialTransactionID)
	}

	deliverySite := formatInspectionDetailDeliverySite(details)
	writeLabelWithID(out, "Delivery Site", deliverySite, details.DeliverySiteID)
	writeLabelWithID(out, "Customer", details.Customer, details.CustomerID)
	writeLabelWithID(out, "Broker", details.Broker, details.BrokerID)
	writeLabelWithID(out, "Material Supplier", details.MaterialSupplier, details.MaterialSupplierID)
	writeLabelWithID(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)
	writeLabelWithID(out, "Trip", "", details.TripID)
	writeLabelWithID(out, "Tender Job Schedule Shift", "", details.TenderJobScheduleID)
	writeLabelWithID(out, "Changed By", details.ChangedByName, details.ChangedByID)
	writeLabelWithID(out, "Created By", details.CreatedByName, details.CreatedByID)

	if details.TonsAccepted != 0 || details.TonsRejected != 0 {
		fmt.Fprintf(out, "Tons Accepted: %.2f\n", details.TonsAccepted)
		fmt.Fprintf(out, "Tons Rejected: %.2f\n", details.TonsRejected)
	}

	if len(details.Rejections) > 0 {
		fmt.Fprintf(out, "Rejections (%d):\n", len(details.Rejections))
		for _, rejection := range details.Rejections {
			line := fmt.Sprintf("  - %s: %.2f", rejection.ID, rejection.Quantity)
			if rejection.UnitOfMeasure != "" {
				line = fmt.Sprintf("%s %s", line, rejection.UnitOfMeasure)
			} else if rejection.UnitOfMeasureID != "" {
				line = fmt.Sprintf("%s (unit %s)", line, rejection.UnitOfMeasureID)
			}
			if rejection.Note != "" {
				line = fmt.Sprintf("%s | %s", line, rejection.Note)
			}
			if rejection.RejectedByName != "" {
				line = fmt.Sprintf("%s | by %s", line, rejection.RejectedByName)
			}
			fmt.Fprintln(out, line)
		}
	}

	return nil
}

func formatInspectionDetailDeliverySite(details materialTransactionInspectionDetails) string {
	if details.DeliverySite != "" && details.DeliverySiteType != "" {
		return fmt.Sprintf("%s (%s)", details.DeliverySite, details.DeliverySiteType)
	}
	if details.DeliverySite != "" {
		return details.DeliverySite
	}
	if details.DeliverySiteType != "" && details.DeliverySiteID != "" {
		return fmt.Sprintf("%s:%s", details.DeliverySiteType, details.DeliverySiteID)
	}
	return details.DeliverySiteID
}
