package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type materialPurchaseOrderReleasesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Status           string
	TokenFilter      string
	PurchaseOrder    string
	Trucker          string
	TenderJobShift   string
	JobShift         string
	Quantity         string
	Broker           string
	Customer         string
	ValidForCustomer string
	IsAssigned       string
	MaterialSupplier string
	Active           string
	NotActive        string
}

type materialPurchaseOrderReleaseRow struct {
	ID                     string  `json:"id"`
	ReleaseNumber          string  `json:"release_number,omitempty"`
	Status                 string  `json:"status,omitempty"`
	Quantity               float64 `json:"quantity,omitempty"`
	Token                  string  `json:"token,omitempty"`
	PurchaseOrderID        string  `json:"purchase_order_id,omitempty"`
	PurchaseOrderNumber    string  `json:"purchase_order_number,omitempty"`
	TruckerID              string  `json:"trucker_id,omitempty"`
	TruckerName            string  `json:"trucker,omitempty"`
	TenderJobScheduleShift string  `json:"tender_job_schedule_shift_id,omitempty"`
	JobScheduleShift       string  `json:"job_schedule_shift_id,omitempty"`
	RedemptionID           string  `json:"redemption_id,omitempty"`
}

func newMaterialPurchaseOrderReleasesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material purchase order releases",
		Long: `List material purchase order releases with filtering and pagination.

Output Columns:
  ID            Release identifier
  STATUS        Release status
  TOKEN         Release token
  QTY           Released quantity
  PURCHASE ORD  Purchase order ID or number
  SHIFT         Assigned shift (tender or job)
  TRUCKER       Trucker name or ID

Filters:
  --status                  Filter by status (editing,approved,redeemed,closed)
  --release-token           Filter by release token
  --purchase-order          Filter by purchase order ID
  --trucker                 Filter by trucker ID
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --job-schedule-shift       Filter by job schedule shift ID
  --quantity                Filter by quantity
  --broker                  Filter by broker ID
  --customer                Filter by customer ID
  --valid-for-customer      Filter by customer eligibility
  --is-assigned             Filter by assignment status (true/false)
  --material-supplier       Filter by material supplier ID
  --active                  Filter by active status (true/false)
  --not-active              Filter by not active status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List releases
  xbe view material-purchase-order-releases list

  # Filter by purchase order
  xbe view material-purchase-order-releases list --purchase-order 123

  # Filter by status
  xbe view material-purchase-order-releases list --status approved

  # Show only assigned releases
  xbe view material-purchase-order-releases list --is-assigned true

  # Output as JSON
  xbe view material-purchase-order-releases list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialPurchaseOrderReleasesList,
	}
	initMaterialPurchaseOrderReleasesListFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrderReleasesCmd.AddCommand(newMaterialPurchaseOrderReleasesListCmd())
}

func initMaterialPurchaseOrderReleasesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (editing,approved,redeemed,closed)")
	cmd.Flags().String("release-token", "", "Filter by release token")
	cmd.Flags().String("purchase-order", "", "Filter by purchase order ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID (comma-separated for multiple)")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID (comma-separated for multiple)")
	cmd.Flags().String("quantity", "", "Filter by quantity")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("valid-for-customer", "", "Filter by customer ID for eligibility")
	cmd.Flags().String("is-assigned", "", "Filter by assignment status (true/false)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("active", "", "Filter by active status (true/false)")
	cmd.Flags().String("not-active", "", "Filter by not active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrderReleasesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialPurchaseOrderReleasesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-purchase-order-releases]", "status,quantity,token,purchase-order,trucker,tender-job-schedule-shift,job-schedule-shift,redemption")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[tender-job-schedule-shifts]", "start-at")
	query.Set("fields[job-schedule-shifts]", "start-at")
	query.Set("fields[material-purchase-order-release-redemptions]", "ticket-number")
	query.Set("include", "purchase-order,trucker,tender-job-schedule-shift,job-schedule-shift,redemption")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[token]", opts.TokenFilter)
	setFilterIfPresent(query, "filter[purchase-order]", opts.PurchaseOrder)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobShift)
	setFilterIfPresent(query, "filter[job-schedule-shift]", opts.JobShift)
	setFilterIfPresent(query, "filter[quantity]", opts.Quantity)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[valid-for-customer]", opts.ValidForCustomer)
	setFilterIfPresent(query, "filter[is-assigned]", opts.IsAssigned)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[active]", opts.Active)
	setFilterIfPresent(query, "filter[not-active]", opts.NotActive)

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-order-releases", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildMaterialPurchaseOrderReleaseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialPurchaseOrderReleasesTable(cmd, rows)
}

func parseMaterialPurchaseOrderReleasesListOptions(cmd *cobra.Command) (materialPurchaseOrderReleasesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	tokenFilter, _ := cmd.Flags().GetString("release-token")
	purchaseOrder, _ := cmd.Flags().GetString("purchase-order")
	trucker, _ := cmd.Flags().GetString("trucker")
	tenderJobShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobShift, _ := cmd.Flags().GetString("job-schedule-shift")
	quantity, _ := cmd.Flags().GetString("quantity")
	broker, _ := cmd.Flags().GetString("broker")
	customer, _ := cmd.Flags().GetString("customer")
	validForCustomer, _ := cmd.Flags().GetString("valid-for-customer")
	isAssigned, _ := cmd.Flags().GetString("is-assigned")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	active, _ := cmd.Flags().GetString("active")
	notActive, _ := cmd.Flags().GetString("not-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialPurchaseOrderReleasesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Status:           status,
		TokenFilter:      tokenFilter,
		PurchaseOrder:    purchaseOrder,
		Trucker:          trucker,
		TenderJobShift:   tenderJobShift,
		JobShift:         jobShift,
		Quantity:         quantity,
		Broker:           broker,
		Customer:         customer,
		ValidForCustomer: validForCustomer,
		IsAssigned:       isAssigned,
		MaterialSupplier: materialSupplier,
		Active:           active,
		NotActive:        notActive,
	}, nil
}

func buildMaterialPurchaseOrderReleaseRows(resp jsonAPIResponse) []materialPurchaseOrderReleaseRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialPurchaseOrderReleaseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialPurchaseOrderReleaseRow{
			ID:            resource.ID,
			ReleaseNumber: stringAttr(resource.Attributes, "release-number"),
			Status:        stringAttr(resource.Attributes, "status"),
			Token:         stringAttr(resource.Attributes, "token"),
			Quantity:      floatAttr(resource.Attributes, "quantity"),
		}

		if rel, ok := resource.Relationships["purchase-order"]; ok && rel.Data != nil {
			row.PurchaseOrderID = rel.Data.ID
			if purchaseOrder, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.PurchaseOrderNumber = stringAttr(purchaseOrder.Attributes, "purchase-order-id")
				if row.PurchaseOrderNumber == "" {
					row.PurchaseOrderNumber = stringAttr(purchaseOrder.Attributes, "sales-order-id")
				}
			}
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
			if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.TruckerName = stringAttr(trucker.Attributes, "company-name")
				if row.TruckerName == "" {
					row.TruckerName = stringAttr(trucker.Attributes, "name")
				}
			}
		}

		row.TenderJobScheduleShift = relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift")
		row.JobScheduleShift = relationshipIDFromMap(resource.Relationships, "job-schedule-shift")
		row.RedemptionID = relationshipIDFromMap(resource.Relationships, "redemption")

		rows = append(rows, row)
	}
	return rows
}

func renderMaterialPurchaseOrderReleasesTable(cmd *cobra.Command, rows []materialPurchaseOrderReleaseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material purchase order releases found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTOKEN\tQTY\tPURCHASE ORD\tSHIFT\tTRUCKER")
	for _, row := range rows {
		shift := ""
		if row.TenderJobScheduleShift != "" {
			shift = "T:" + row.TenderJobScheduleShift
		} else if row.JobScheduleShift != "" {
			shift = "J:" + row.JobScheduleShift
		}

		purchaseOrder := row.PurchaseOrderNumber
		if purchaseOrder == "" {
			purchaseOrder = row.PurchaseOrderID
		}

		trucker := row.TruckerName
		if trucker == "" {
			trucker = row.TruckerID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%.2f\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Token,
			row.Quantity,
			truncateString(purchaseOrder, 20),
			truncateString(shift, 18),
			truncateString(trucker, 20),
		)
	}
	return writer.Flush()
}
