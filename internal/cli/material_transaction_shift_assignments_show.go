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

type materialTransactionShiftAssignmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTransactionShiftAssignmentDetails struct {
	ID                       string   `json:"id"`
	MaterialTransactionIDs   []string `json:"material_transaction_ids,omitempty"`
	TenderJobScheduleShiftID string   `json:"tender_job_schedule_shift_id,omitempty"`
	JobProductionPlanID      string   `json:"job_production_plan_id,omitempty"`
	JobProductionPlan        string   `json:"job_production_plan,omitempty"`
	TruckerID                string   `json:"trucker_id,omitempty"`
	TruckerName              string   `json:"trucker_name,omitempty"`
	BrokerID                 string   `json:"broker_id,omitempty"`
	BrokerName               string   `json:"broker_name,omitempty"`
	CreatedByID              string   `json:"created_by_id,omitempty"`
	CreatedByName            string   `json:"created_by_name,omitempty"`
	BrokerTenderID           string   `json:"broker_tender_id,omitempty"`
	TimeCardIDs              []string `json:"time_card_ids,omitempty"`
	ProcessedAt              string   `json:"processed_at,omitempty"`
	IsProcessed              bool     `json:"is_processed,omitempty"`
	EnableLinkInvoiced       bool     `json:"enable_link_invoiced,omitempty"`
	Comment                  string   `json:"comment,omitempty"`
	AssignmentResults        any      `json:"assignment_results,omitempty"`
}

func newMaterialTransactionShiftAssignmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material transaction shift assignment details",
		Long: `Show the full details of a material transaction shift assignment.

Arguments:
  <id>  The assignment ID (required).`,
		Example: `  # Show an assignment
  xbe view material-transaction-shift-assignments show 123

  # Output as JSON
  xbe view material-transaction-shift-assignments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTransactionShiftAssignmentsShow,
	}
	initMaterialTransactionShiftAssignmentsShowFlags(cmd)
	return cmd
}

func init() {
	materialTransactionShiftAssignmentsCmd.AddCommand(newMaterialTransactionShiftAssignmentsShowCmd())
}

func initMaterialTransactionShiftAssignmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionShiftAssignmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTransactionShiftAssignmentsShowOptions(cmd)
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
		return fmt.Errorf("material transaction shift assignment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-transaction-shift-assignments]", "material-transaction-ids,tender-job-schedule-shift,job-production-plan,trucker,broker,created-by,broker-tender,time-cards,assignment-results,processed-at,is-processed,enable-link-invoiced,comment")
	query.Set("include", "tender-job-schedule-shift,job-production-plan,trucker,broker,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-shift-assignments/"+id, query)
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

	details := buildMaterialTransactionShiftAssignmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTransactionShiftAssignmentDetails(cmd, details)
}

func parseMaterialTransactionShiftAssignmentsShowOptions(cmd *cobra.Command) (materialTransactionShiftAssignmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionShiftAssignmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTransactionShiftAssignmentDetails(resp jsonAPISingleResponse) materialTransactionShiftAssignmentDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := materialTransactionShiftAssignmentDetails{
		ID:                     resp.Data.ID,
		MaterialTransactionIDs: stringSliceAttr(attrs, "material-transaction-ids"),
		ProcessedAt:            formatDateTime(stringAttr(attrs, "processed-at")),
		IsProcessed:            boolAttr(attrs, "is-processed"),
		EnableLinkInvoiced:     boolAttr(attrs, "enable-link-invoiced"),
		Comment:                stringAttr(attrs, "comment"),
		AssignmentResults:      anyAttr(attrs, "assignment-results"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if jpp, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlan = firstNonEmpty(
				stringAttr(jpp.Attributes, "job-number"),
				stringAttr(jpp.Attributes, "job-name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["broker-tender"]; ok && rel.Data != nil {
		details.BrokerTenderID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["time-cards"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			ids := make([]string, 0, len(refs))
			for _, ref := range refs {
				if ref.ID != "" {
					ids = append(ids, ref.ID)
				}
			}
			details.TimeCardIDs = ids
		}
	}

	return details
}

func renderMaterialTransactionShiftAssignmentDetails(cmd *cobra.Command, details materialTransactionShiftAssignmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.JobProductionPlanID != "" || details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", formatRelated(details.JobProductionPlan, details.JobProductionPlanID))
	}
	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}
	if details.BrokerTenderID != "" {
		fmt.Fprintf(out, "Broker Tender: %s\n", details.BrokerTenderID)
	}
	if len(details.MaterialTransactionIDs) > 0 {
		fmt.Fprintf(out, "Material Transactions: %d\n", len(details.MaterialTransactionIDs))
		fmt.Fprintf(out, "Material Transaction IDs: %s\n", strings.Join(details.MaterialTransactionIDs, ", "))
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	fmt.Fprintf(out, "Processed: %t\n", details.IsProcessed)
	fmt.Fprintf(out, "Enable Link Invoiced: %t\n", details.EnableLinkInvoiced)
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Cards: %d\n", len(details.TimeCardIDs))
		fmt.Fprintf(out, "Time Card IDs: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}

	if details.AssignmentResults != nil {
		if formatted := formatAnyJSON(details.AssignmentResults); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Assignment Results:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
