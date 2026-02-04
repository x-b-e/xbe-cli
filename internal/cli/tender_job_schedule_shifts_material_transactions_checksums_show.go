package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type tenderJobScheduleShiftsMaterialTransactionsChecksumsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderJobScheduleShiftsMaterialTransactionsChecksumDetails struct {
	ID                                     string   `json:"id"`
	RawJobNumber                           string   `json:"raw_job_number,omitempty"`
	TransactionAtMin                       string   `json:"transaction_at_min,omitempty"`
	TransactionAtMax                       string   `json:"transaction_at_max,omitempty"`
	MaterialSiteIDs                        []string `json:"material_site_ids,omitempty"`
	JobProductionPlanID                    string   `json:"job_production_plan_id,omitempty"`
	Checksum                               []string `json:"checksum,omitempty"`
	TenderJobScheduleShiftIDs              []string `json:"tender_job_schedule_shift_ids,omitempty"`
	RawMaterialTransactionIDs              []string `json:"raw_material_transaction_ids,omitempty"`
	MaterialTransactionIDs                 []string `json:"material_transaction_ids,omitempty"`
	JobIDs                                 []string `json:"job_ids,omitempty"`
	MaterialSiteRelationshipIDs            []string `json:"material_site_relationship_ids,omitempty"`
	MaterialTransactionsOnlyViaJppIDs      []string `json:"material_transactions_only_via_jpp_ids,omitempty"`
	RawMaterialTransactionsOnlyViaJppIDs   []string `json:"raw_material_transactions_only_via_jpp_ids,omitempty"`
	RawMaterialTransactionsDescriptions    any      `json:"raw_material_transactions_descriptions,omitempty"`
	RawMaterialTransactionsOnlyViaJppDescs any      `json:"raw_material_transactions_only_via_jpp_descriptions,omitempty"`
	RawMaterialTransactionsTrucksSummary   any      `json:"raw_material_transactions_trucks_summary,omitempty"`
	TenderJobScheduleShiftsDescriptions    any      `json:"tender_job_schedule_shifts_descriptions,omitempty"`
	TenderJobScheduleShiftsOnlyViaJppDescs any      `json:"tender_job_schedule_shifts_only_via_jpp_descriptions,omitempty"`
}

type tenderJobScheduleShiftsMaterialTransactionsChecksumResponse struct {
	Data     jsonAPIResource   `json:"data"`
	Included []jsonAPIResource `json:"included"`
	Meta     map[string]any    `json:"meta"`
}

func newTenderJobScheduleShiftsMaterialTransactionsChecksumsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender job schedule shift material transaction checksum details",
		Long: `Show tender job schedule shift material transaction checksum details.

Checksum records compare raw material transactions to tender job schedule shifts
for a given job number and transaction window.

Output Fields:
  ID                          Checksum record identifier
  Raw Job Number              Raw job number (comma-separated)
  Transaction At Min/Max      Transaction window range
  Material Site IDs           Material site IDs used for filtering
  Job Production Plan ID      Job production plan context (optional)
  Checksum                    Diagnostic checksum output (meta)
  Related IDs                 Matched relationships (shifts, transactions, jobs)
  Description Arrays          Raw/tender description payloads from the server

Arguments:
  <id>  The checksum ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show checksum details
  xbe view tender-job-schedule-shifts-material-transactions-checksums show 123

  # Output as JSON
  xbe view tender-job-schedule-shifts-material-transactions-checksums show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderJobScheduleShiftsMaterialTransactionsChecksumsShow,
	}
	initTenderJobScheduleShiftsMaterialTransactionsChecksumsShowFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftsMaterialTransactionsChecksumsCmd.AddCommand(newTenderJobScheduleShiftsMaterialTransactionsChecksumsShowCmd())
}

func initTenderJobScheduleShiftsMaterialTransactionsChecksumsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftsMaterialTransactionsChecksumsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTenderJobScheduleShiftsMaterialTransactionsChecksumsShowOptions(cmd)
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
		return fmt.Errorf("checksum id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-job-schedule-shifts-material-transactions-checksums]",
		"raw-job-number,transaction-at-min,transaction-at-max,material-site-ids,job-production-plan-id,raw-material-transactions-descriptions,raw-material-transactions-only-via-jpp-descriptions,raw-material-transactions-trucks-summary,tender-job-schedule-shifts-descriptions,tender-job-schedule-shifts-only-via-jpp-descriptions,tender-job-schedule-shifts,raw-material-transactions,material-transactions,jobs,material-sites,material-transactions-only-via-jpp,raw-material-transactions-only-via-jpp")

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shifts-material-transactions-checksums/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp tenderJobScheduleShiftsMaterialTransactionsChecksumResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildTenderJobScheduleShiftsMaterialTransactionsChecksumDetails(resp.Data, resp.Meta)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderJobScheduleShiftsMaterialTransactionsChecksumDetails(cmd, details)
}

func parseTenderJobScheduleShiftsMaterialTransactionsChecksumsShowOptions(cmd *cobra.Command) (tenderJobScheduleShiftsMaterialTransactionsChecksumsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftsMaterialTransactionsChecksumsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderJobScheduleShiftsMaterialTransactionsChecksumDetails(resource jsonAPIResource, meta map[string]any) tenderJobScheduleShiftsMaterialTransactionsChecksumDetails {
	attrs := resource.Attributes
	details := tenderJobScheduleShiftsMaterialTransactionsChecksumDetails{
		ID:                                     resource.ID,
		RawJobNumber:                           stringAttr(attrs, "raw-job-number"),
		TransactionAtMin:                       formatDateTime(stringAttr(attrs, "transaction-at-min")),
		TransactionAtMax:                       formatDateTime(stringAttr(attrs, "transaction-at-max")),
		MaterialSiteIDs:                        stringSliceAttr(attrs, "material-site-ids"),
		JobProductionPlanID:                    stringAttr(attrs, "job-production-plan-id"),
		Checksum:                               checksumMeta(meta),
		RawMaterialTransactionsDescriptions:    attrValue(attrs, "raw-material-transactions-descriptions"),
		RawMaterialTransactionsOnlyViaJppDescs: attrValue(attrs, "raw-material-transactions-only-via-jpp-descriptions"),
		RawMaterialTransactionsTrucksSummary:   attrValue(attrs, "raw-material-transactions-trucks-summary"),
		TenderJobScheduleShiftsDescriptions:    attrValue(attrs, "tender-job-schedule-shifts-descriptions"),
		TenderJobScheduleShiftsOnlyViaJppDescs: attrValue(attrs, "tender-job-schedule-shifts-only-via-jpp-descriptions"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["raw-material-transactions"]; ok {
		details.RawMaterialTransactionIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["material-transactions"]; ok {
		details.MaterialTransactionIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["jobs"]; ok {
		details.JobIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["material-sites"]; ok {
		details.MaterialSiteRelationshipIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["material-transactions-only-via-jpp"]; ok {
		details.MaterialTransactionsOnlyViaJppIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["raw-material-transactions-only-via-jpp"]; ok {
		details.RawMaterialTransactionsOnlyViaJppIDs = relationshipIDList(rel)
	}

	return details
}

func renderTenderJobScheduleShiftsMaterialTransactionsChecksumDetails(cmd *cobra.Command, details tenderJobScheduleShiftsMaterialTransactionsChecksumDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RawJobNumber != "" {
		fmt.Fprintf(out, "Raw Job Number: %s\n", details.RawJobNumber)
	}
	if details.TransactionAtMin != "" {
		fmt.Fprintf(out, "Transaction At Min: %s\n", details.TransactionAtMin)
	}
	if details.TransactionAtMax != "" {
		fmt.Fprintf(out, "Transaction At Max: %s\n", details.TransactionAtMax)
	}
	if len(details.MaterialSiteIDs) > 0 {
		fmt.Fprintf(out, "Material Site IDs: %s\n", strings.Join(details.MaterialSiteIDs, ", "))
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}

	if len(details.Checksum) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Checksum:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, line := range details.Checksum {
			fmt.Fprintln(out, line)
		}
	}

	renderJSONSection(out, "Raw Material Transactions Descriptions", details.RawMaterialTransactionsDescriptions)
	renderJSONSection(out, "Raw Material Transactions Only Via JPP Descriptions", details.RawMaterialTransactionsOnlyViaJppDescs)
	renderJSONSection(out, "Raw Material Transactions Trucks Summary", details.RawMaterialTransactionsTrucksSummary)
	renderJSONSection(out, "Tender Job Schedule Shifts Descriptions", details.TenderJobScheduleShiftsDescriptions)
	renderJSONSection(out, "Tender Job Schedule Shifts Only Via JPP Descriptions", details.TenderJobScheduleShiftsOnlyViaJppDescs)

	renderIDListSection(out, "Tender Job Schedule Shifts", details.TenderJobScheduleShiftIDs)
	renderIDListSection(out, "Raw Material Transactions", details.RawMaterialTransactionIDs)
	renderIDListSection(out, "Material Transactions", details.MaterialTransactionIDs)
	renderIDListSection(out, "Jobs", details.JobIDs)
	renderIDListSection(out, "Material Sites (Relationship)", details.MaterialSiteRelationshipIDs)
	renderIDListSection(out, "Material Transactions Only Via JPP", details.MaterialTransactionsOnlyViaJppIDs)
	renderIDListSection(out, "Raw Material Transactions Only Via JPP", details.RawMaterialTransactionsOnlyViaJppIDs)

	return nil
}

func renderIDListSection(out io.Writer, title string, ids []string) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, title+":")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	for _, id := range ids {
		fmt.Fprintln(out, "  - "+id)
	}
}

func renderJSONSection(out io.Writer, title string, value any) {
	pretty := formatJSONValue(value)
	if pretty == "" {
		return
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, title+":")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintln(out, pretty)
}

func attrValue(attrs map[string]any, key string) any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		if len(typed) == 0 {
			return nil
		}
	case []string:
		if len(typed) == 0 {
			return nil
		}
	case map[string]any:
		if len(typed) == 0 {
			return nil
		}
	}
	return value
}

func checksumMeta(meta map[string]any) []string {
	if meta == nil {
		return nil
	}
	raw, ok := meta["checksum"]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if item == nil {
				continue
			}
			out = append(out, fmt.Sprintf("%v", item))
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		value := strings.TrimSpace(fmt.Sprintf("%v", typed))
		if value == "" {
			return nil
		}
		return []string{value}
	}
}
