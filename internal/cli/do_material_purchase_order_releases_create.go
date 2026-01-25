package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialPurchaseOrderReleasesCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	Status                 string
	Quantity               string
	PurchaseOrder          string
	Trucker                string
	TenderJobScheduleShift string
	JobScheduleShift       string
}

func newDoMaterialPurchaseOrderReleasesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material purchase order release",
		Long: `Create a material purchase order release.

Required flags:
  --purchase-order   Purchase order ID
  --quantity         Release quantity

Optional flags:
  --status           Release status (editing,approved,redeemed,closed)
  --trucker          Trucker ID
  --tender-job-schedule-shift  Tender job schedule shift ID
  --job-schedule-shift         Job schedule shift ID`,
		Example: `  # Create a release
  xbe do material-purchase-order-releases create \
    --purchase-order 123 \
    --quantity 10

  # Create and assign to a shift
  xbe do material-purchase-order-releases create \
    --purchase-order 123 \
    --quantity 10 \
    --job-schedule-shift 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialPurchaseOrderReleasesCreate,
	}
	initDoMaterialPurchaseOrderReleasesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialPurchaseOrderReleasesCmd.AddCommand(newDoMaterialPurchaseOrderReleasesCreateCmd())
}

func initDoMaterialPurchaseOrderReleasesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Release status (editing,approved,redeemed,closed)")
	cmd.Flags().String("quantity", "", "Release quantity (required)")
	cmd.Flags().String("purchase-order", "", "Purchase order ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("purchase-order")
	cmd.MarkFlagRequired("quantity")
}

func runDoMaterialPurchaseOrderReleasesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialPurchaseOrderReleasesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	attributes := map[string]any{
		"quantity": opts.Quantity,
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{
		"purchase-order": map[string]any{
			"data": map[string]any{
				"type": "material-purchase-orders",
				"id":   opts.PurchaseOrder,
			},
		},
	}
	if opts.Trucker != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}
	if opts.TenderJobScheduleShift != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}
	if opts.JobScheduleShift != "" {
		relationships["job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShift,
			},
		}
	}

	data := map[string]any{
		"type":          "material-purchase-order-releases",
		"attributes":    attributes,
		"relationships": relationships,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-purchase-order-releases", jsonBody)
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

	row := materialPurchaseOrderReleaseRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material purchase order release %s\n", row.ID)
	return nil
}

func parseDoMaterialPurchaseOrderReleasesCreateOptions(cmd *cobra.Command) (doMaterialPurchaseOrderReleasesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	quantity, _ := cmd.Flags().GetString("quantity")
	purchaseOrder, _ := cmd.Flags().GetString("purchase-order")
	trucker, _ := cmd.Flags().GetString("trucker")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialPurchaseOrderReleasesCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		Status:                 status,
		Quantity:               quantity,
		PurchaseOrder:          purchaseOrder,
		Trucker:                trucker,
		TenderJobScheduleShift: tenderJobScheduleShift,
		JobScheduleShift:       jobScheduleShift,
	}, nil
}

func materialPurchaseOrderReleaseRowFromSingle(resp jsonAPISingleResponse) materialPurchaseOrderReleaseRow {
	resource := resp.Data
	row := materialPurchaseOrderReleaseRow{
		ID:                     resource.ID,
		ReleaseNumber:          stringAttr(resource.Attributes, "release-number"),
		Status:                 stringAttr(resource.Attributes, "status"),
		Quantity:               floatAttr(resource.Attributes, "quantity"),
		PurchaseOrderID:        relationshipIDFromMap(resource.Relationships, "purchase-order"),
		TruckerID:              relationshipIDFromMap(resource.Relationships, "trucker"),
		TenderJobScheduleShift: relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift"),
		JobScheduleShift:       relationshipIDFromMap(resource.Relationships, "job-schedule-shift"),
		RedemptionID:           relationshipIDFromMap(resource.Relationships, "redemption"),
	}

	return row
}
