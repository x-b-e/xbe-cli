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

type materialTransactionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionDetails struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	TransactionAt   string `json:"transaction_at"`
	TransactionDate string `json:"transaction_date"`
	TicketNumber    string `json:"ticket_number"`
	TicketBOLNumber string `json:"ticket_bol_number,omitempty"`
	IsVoided        bool   `json:"is_voided"`

	// Weights
	NetWeightLbs    float64 `json:"net_weight_lbs"`
	GrossWeightLbs  float64 `json:"gross_weight_lbs"`
	TareWeightLbs   float64 `json:"tare_weight_lbs"`
	MaxGVMWeightLbs float64 `json:"max_gvm_weight_lbs,omitempty"`
	Tons            float64 `json:"tons"`

	// Timing
	CycleMinutes         float64 `json:"cycle_minutes,omitempty"`
	PickupDwellMinutes   float64 `json:"pickup_dwell_minutes,omitempty"`
	LoadedDrivingMinutes float64 `json:"loaded_driving_minutes,omitempty"`
	DeliveryDwellMinutes float64 `json:"delivery_dwell_minutes,omitempty"`

	// Relationships
	MaterialType        string `json:"material_type,omitempty"`
	MaterialTypeID      string `json:"material_type_id,omitempty"`
	MaterialSite        string `json:"material_site,omitempty"`
	MaterialSiteID      string `json:"material_site_id,omitempty"`
	MaterialSupplier    string `json:"material_supplier,omitempty"`
	MaterialSupplierID  string `json:"material_supplier_id,omitempty"`
	Origin              string `json:"origin,omitempty"`
	OriginID            string `json:"origin_id,omitempty"`
	OriginType          string `json:"origin_type,omitempty"`
	Destination         string `json:"destination,omitempty"`
	DestinationID       string `json:"destination_id,omitempty"`
	DestinationType     string `json:"destination_type,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`

	// Status changes
	StatusChanges []statusChange `json:"status_changes,omitempty"`
}

type statusChange struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Comment   string `json:"comment,omitempty"`
	ChangedAt string `json:"changed_at"`
	ChangedBy string `json:"changed_by,omitempty"`
}

func newMaterialTransactionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction details",
		Long: `Show the full details of a specific material transaction.

Retrieves and displays comprehensive information about a material transaction
including weights, timing, locations, and status history.

Output Fields (table format):
  ID              Unique transaction identifier
  Status          Current workflow state
  Transaction At  Full timestamp of the transaction
  Ticket Number   Associated ticket identifier
  Weights         Gross, Tare, Net weights in lbs, plus calculated tons
  Cycle Time      Total cycle time in minutes
  Material Type   What was transported
  Origin          Where the material came from
  Destination     Where the material was delivered
  Status History  Chronological list of status changes

Arguments:
  <id>          The transaction ID (required). You can find IDs using the list command.`,
		Example: `  # View a transaction by ID
  xbe view material-transactions show 123

  # Get transaction as JSON
  xbe view material-transactions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionsShow,
	}
	initMaterialTransactionsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionsCmd.AddCommand(newMaterialTransactionsShowCmd())
}

func initMaterialTransactionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialTransactionsShowOptions(cmd)
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
		return fmt.Errorf("material transaction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transactions]", "status,transaction-at,ticket-number,ticket-bol-number,is-voided,net-weight-lbs,gross-weight-lbs,tare-weight-lbs,max-gvm-weight-lbs,cycle-minutes,pickup-dwell-minutes,loaded-driving-minutes,delivery-dwell-minutes,material-type,material-site,trip,job-production-plan,material-transaction-status-changes")
	query.Set("fields[material-types]", "name")
	query.Set("fields[material-sites]", "name,material-supplier")
	query.Set("fields[material-suppliers]", "company-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[trips]", "origin,destination")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[material-transaction-status-changes]", "status,comment,created-at")
	query.Set("fields[users]", "name")
	query.Set("include", "material-type,material-site.material-supplier,trip.origin,trip.destination,job-production-plan,material-transaction-status-changes")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transactions/"+id, query)
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

	details := buildMaterialTransactionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionDetails(cmd, details)
}

func parseMaterialTransactionsShowOptions(cmd *cobra.Command) (materialTransactionsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return materialTransactionsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return materialTransactionsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return materialTransactionsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return materialTransactionsShowOptions{}, err
	}

	return materialTransactionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionDetails(resp jsonAPISingleResponse) materialTransactionDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	netWeightLbs := floatAttr(attrs, "net-weight-lbs")

	details := materialTransactionDetails{
		ID:                   resp.Data.ID,
		Status:               stringAttr(attrs, "status"),
		TransactionAt:        stringAttr(attrs, "transaction-at"),
		TransactionDate:      formatDate(stringAttr(attrs, "transaction-at")),
		TicketNumber:         stringAttr(attrs, "ticket-number"),
		TicketBOLNumber:      stringAttr(attrs, "ticket-bol-number"),
		IsVoided:             boolAttr(attrs, "is-voided"),
		NetWeightLbs:         netWeightLbs,
		GrossWeightLbs:       floatAttr(attrs, "gross-weight-lbs"),
		TareWeightLbs:        floatAttr(attrs, "tare-weight-lbs"),
		MaxGVMWeightLbs:      floatAttr(attrs, "max-gvm-weight-lbs"),
		Tons:                 netWeightLbs / 2000.0,
		CycleMinutes:         floatAttr(attrs, "cycle-minutes"),
		PickupDwellMinutes:   floatAttr(attrs, "pickup-dwell-minutes"),
		LoadedDrivingMinutes: floatAttr(attrs, "loaded-driving-minutes"),
		DeliveryDwellMinutes: floatAttr(attrs, "delivery-dwell-minutes"),
	}

	// Resolve material type
	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = stringAttr(mt.Attributes, "name")
		}
	}

	// Resolve material site and supplier
	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(ms.Attributes, "name")
			// Get supplier from material site
			if supplierRel, ok := ms.Relationships["material-supplier"]; ok && supplierRel.Data != nil {
				details.MaterialSupplierID = supplierRel.Data.ID
				if supplier, ok := included[resourceKey(supplierRel.Data.Type, supplierRel.Data.ID)]; ok {
					details.MaterialSupplier = stringAttr(supplier.Attributes, "company-name")
				}
			}
		}
	}

	// Resolve job production plan
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if jpp, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			jobNumber := stringAttr(jpp.Attributes, "job-number")
			jobName := stringAttr(jpp.Attributes, "job-name")
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	// Resolve trip origin/destination
	if rel, ok := resp.Data.Relationships["trip"]; ok && rel.Data != nil {
		if trip, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			// Origin
			if originRel, ok := trip.Relationships["origin"]; ok && originRel.Data != nil {
				details.OriginID = originRel.Data.ID
				details.OriginType = originRel.Data.Type
				if origin, ok := included[resourceKey(originRel.Data.Type, originRel.Data.ID)]; ok {
					details.Origin = stringAttr(origin.Attributes, "name")
				}
			}
			// Destination
			if destRel, ok := trip.Relationships["destination"]; ok && destRel.Data != nil {
				details.DestinationID = destRel.Data.ID
				details.DestinationType = destRel.Data.Type
				if dest, ok := included[resourceKey(destRel.Data.Type, destRel.Data.ID)]; ok {
					details.Destination = stringAttr(dest.Attributes, "name")
				}
			}
		}
	}

	// Resolve status changes
	if rel, ok := resp.Data.Relationships["material-transaction-status-changes"]; ok && rel.raw != nil {
		var statusChangeRefs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &statusChangeRefs); err == nil {
			for _, ref := range statusChangeRefs {
				if sc, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					change := statusChange{
						ID:        sc.ID,
						Status:    stringAttr(sc.Attributes, "status"),
						Comment:   stringAttr(sc.Attributes, "comment"),
						ChangedAt: formatDateTime(stringAttr(sc.Attributes, "created-at")),
					}
					// Get changed by user
					if userRel, ok := sc.Relationships["created-by"]; ok && userRel.Data != nil {
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							change.ChangedBy = stringAttr(user.Attributes, "name")
						}
					}
					details.StatusChanges = append(details.StatusChanges, change)
				}
			}
		}
	}

	return details
}

func formatDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	// Return as-is for now, could parse and reformat
	return value
}

func renderMaterialTransactionDetails(cmd *cobra.Command, details materialTransactionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Status: %s\n", details.Status)
	if details.IsVoided {
		fmt.Fprintln(out, "Voided: yes")
	}
	fmt.Fprintf(out, "Transaction At: %s\n", details.TransactionAt)
	fmt.Fprintf(out, "Ticket Number: %s\n", details.TicketNumber)
	if details.TicketBOLNumber != "" {
		fmt.Fprintf(out, "BOL Number: %s\n", details.TicketBOLNumber)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Weights:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Gross: %.0f lbs\n", details.GrossWeightLbs)
	fmt.Fprintf(out, "  Tare:  %.0f lbs\n", details.TareWeightLbs)
	fmt.Fprintf(out, "  Net:   %.0f lbs (%.2f tons)\n", details.NetWeightLbs, details.Tons)
	if details.MaxGVMWeightLbs > 0 {
		fmt.Fprintf(out, "  Max GVM: %.0f lbs\n", details.MaxGVMWeightLbs)
	}

	if details.CycleMinutes > 0 || details.PickupDwellMinutes > 0 || details.LoadedDrivingMinutes > 0 || details.DeliveryDwellMinutes > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Timing:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if details.CycleMinutes > 0 {
			fmt.Fprintf(out, "  Cycle Time: %.0f min\n", details.CycleMinutes)
		}
		if details.PickupDwellMinutes > 0 {
			fmt.Fprintf(out, "  Pickup Dwell: %.0f min\n", details.PickupDwellMinutes)
		}
		if details.LoadedDrivingMinutes > 0 {
			fmt.Fprintf(out, "  Loaded Driving: %.0f min\n", details.LoadedDrivingMinutes)
		}
		if details.DeliveryDwellMinutes > 0 {
			fmt.Fprintf(out, "  Delivery Dwell: %.0f min\n", details.DeliveryDwellMinutes)
		}
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Material:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	if details.MaterialType != "" {
		fmt.Fprintf(out, "  Type: %s (ID: %s)\n", details.MaterialType, details.MaterialTypeID)
	}
	if details.MaterialSite != "" {
		fmt.Fprintf(out, "  Site: %s (ID: %s)\n", details.MaterialSite, details.MaterialSiteID)
	}
	if details.MaterialSupplier != "" {
		fmt.Fprintf(out, "  Supplier: %s (ID: %s)\n", details.MaterialSupplier, details.MaterialSupplierID)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Route:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	if details.Origin != "" {
		fmt.Fprintf(out, "  Origin: %s (ID: %s, Type: %s)\n", details.Origin, details.OriginID, details.OriginType)
	}
	if details.Destination != "" {
		fmt.Fprintf(out, "  Destination: %s (ID: %s, Type: %s)\n", details.Destination, details.DestinationID, details.DestinationType)
	}

	if details.JobProductionPlan != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Assignment:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  Job Production Plan: %s (ID: %s)\n", details.JobProductionPlan, details.JobProductionPlanID)
	}

	if len(details.StatusChanges) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Status History:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, sc := range details.StatusChanges {
			if sc.ChangedBy != "" {
				fmt.Fprintf(out, "  [%s] %s by %s\n", sc.Status, sc.ChangedAt, sc.ChangedBy)
			} else {
				fmt.Fprintf(out, "  [%s] %s\n", sc.Status, sc.ChangedAt)
			}
			if sc.Comment != "" {
				fmt.Fprintf(out, "    Comment: %s\n", sc.Comment)
			}
		}
	}

	return nil
}
