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

type materialPurchaseOrderReleasesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialPurchaseOrderReleaseDetails struct {
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
	RedemptionTicketNumber string  `json:"redemption_ticket_number,omitempty"`
	CreatedAt              string  `json:"created_at,omitempty"`
	UpdatedAt              string  `json:"updated_at,omitempty"`
}

func newMaterialPurchaseOrderReleasesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material purchase order release details",
		Long: `Show the full details of a material purchase order release.

Output Fields:
  ID
  Release Number
  Status
  Quantity
  Token
  Purchase Order ID/Number
  Trucker
  Tender Job Schedule Shift ID
  Job Schedule Shift ID
  Redemption ID
  Redemption Ticket Number
  Created At
  Updated At

Arguments:
  <id>    The release ID (required). You can find IDs using the list command.`,
		Example: `  # Show a release
  xbe view material-purchase-order-releases show 123

  # Get JSON output
  xbe view material-purchase-order-releases show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialPurchaseOrderReleasesShow,
	}
	initMaterialPurchaseOrderReleasesShowFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrderReleasesCmd.AddCommand(newMaterialPurchaseOrderReleasesShowCmd())
}

func initMaterialPurchaseOrderReleasesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrderReleasesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialPurchaseOrderReleasesShowOptions(cmd)
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
		return fmt.Errorf("material purchase order release id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-purchase-order-releases]", "status,quantity,token,purchase-order,trucker,tender-job-schedule-shift,job-schedule-shift,redemption,created-at,updated-at")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-purchase-order-release-redemptions]", "ticket-number")
	query.Set("include", "purchase-order,trucker,tender-job-schedule-shift,job-schedule-shift,redemption")

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-order-releases/"+id, query)
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

	details := buildMaterialPurchaseOrderReleaseDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialPurchaseOrderReleaseDetails(cmd, details)
}

func parseMaterialPurchaseOrderReleasesShowOptions(cmd *cobra.Command) (materialPurchaseOrderReleasesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialPurchaseOrderReleasesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialPurchaseOrderReleaseDetails(resp jsonAPISingleResponse) materialPurchaseOrderReleaseDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := materialPurchaseOrderReleaseDetails{
		ID:            resource.ID,
		ReleaseNumber: stringAttr(attrs, "release-number"),
		Status:        stringAttr(attrs, "status"),
		Quantity:      floatAttr(attrs, "quantity"),
		Token:         stringAttr(attrs, "token"),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["purchase-order"]; ok && rel.Data != nil {
		details.PurchaseOrderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}

	details.TenderJobScheduleShift = relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift")
	details.JobScheduleShift = relationshipIDFromMap(resource.Relationships, "job-schedule-shift")
	details.RedemptionID = relationshipIDFromMap(resource.Relationships, "redemption")

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.PurchaseOrderID != "" {
		if purchaseOrder, ok := included[resourceKey("material-purchase-orders", details.PurchaseOrderID)]; ok {
			details.PurchaseOrderNumber = stringAttr(purchaseOrder.Attributes, "purchase-order-id")
			if details.PurchaseOrderNumber == "" {
				details.PurchaseOrderNumber = stringAttr(purchaseOrder.Attributes, "sales-order-id")
			}
		}
	}

	if details.TruckerID != "" {
		if trucker, ok := included[resourceKey("truckers", details.TruckerID)]; ok {
			details.TruckerName = stringAttr(trucker.Attributes, "company-name")
			if details.TruckerName == "" {
				details.TruckerName = stringAttr(trucker.Attributes, "name")
			}
		}
	}

	if details.RedemptionID != "" {
		if redemption, ok := included[resourceKey("material-purchase-order-release-redemptions", details.RedemptionID)]; ok {
			details.RedemptionTicketNumber = stringAttr(redemption.Attributes, "ticket-number")
		}
	}

	return details
}

func renderMaterialPurchaseOrderReleaseDetails(cmd *cobra.Command, details materialPurchaseOrderReleaseDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ReleaseNumber != "" {
		fmt.Fprintf(out, "Release Number: %s\n", details.ReleaseNumber)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Quantity != 0 {
		fmt.Fprintf(out, "Quantity: %.2f\n", details.Quantity)
	}
	if details.Token != "" {
		fmt.Fprintf(out, "Token: %s\n", details.Token)
	}
	if details.PurchaseOrderID != "" {
		purchaseOrder := details.PurchaseOrderID
		if details.PurchaseOrderNumber != "" {
			purchaseOrder = fmt.Sprintf("%s (%s)", details.PurchaseOrderNumber, details.PurchaseOrderID)
		}
		fmt.Fprintf(out, "Purchase Order: %s\n", purchaseOrder)
	}
	if details.TruckerID != "" {
		trucker := details.TruckerID
		if details.TruckerName != "" {
			trucker = fmt.Sprintf("%s (%s)", details.TruckerName, details.TruckerID)
		}
		fmt.Fprintf(out, "Trucker: %s\n", trucker)
	}
	if details.TenderJobScheduleShift != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShift)
	}
	if details.JobScheduleShift != "" {
		fmt.Fprintf(out, "Job Schedule Shift ID: %s\n", details.JobScheduleShift)
	}
	if details.RedemptionID != "" {
		fmt.Fprintf(out, "Redemption ID: %s\n", details.RedemptionID)
	}
	if details.RedemptionTicketNumber != "" {
		fmt.Fprintf(out, "Redemption Ticket Number: %s\n", details.RedemptionTicketNumber)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
