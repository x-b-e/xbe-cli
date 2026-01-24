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

type doTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	RawJobNumber        string
	TransactionAtMin    string
	TransactionAtMax    string
	MaterialSiteIDs     string
	JobProductionPlanID string
}

func newDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tender job schedule shift material transaction checksum",
		Long: `Create a tender job schedule shift material transaction checksum.

Checksums compare raw material transactions to tender job schedule shifts over
a specified job number and time window.

Required flags:
  --raw-job-number      Raw job number (comma-separated for multiple)
  --transaction-at-min  Transaction window start (ISO 8601)
  --transaction-at-max  Transaction window end (ISO 8601)

Optional flags:
  --material-site-ids        Comma-separated material site IDs
  --job-production-plan-id   Job production plan ID (for policy authorization)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a checksum for a job number
  xbe do tender-job-schedule-shifts-material-transactions-checksums create \
    --raw-job-number 3882 \
    --transaction-at-min 2025-01-01T00:00:00Z \
    --transaction-at-max 2025-01-02T00:00:00Z

  # Include material site IDs and output JSON
  xbe do tender-job-schedule-shifts-material-transactions-checksums create \
    --raw-job-number 3882 \
    --transaction-at-min 2025-01-01T00:00:00Z \
    --transaction-at-max 2025-01-02T00:00:00Z \
    --material-site-ids 101,102 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreate,
	}
	initDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftsMaterialTransactionsChecksumsCmd.AddCommand(newDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateCmd())
}

func initDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("raw-job-number", "", "Raw job number (comma-separated for multiple)")
	cmd.Flags().String("transaction-at-min", "", "Transaction window start (ISO 8601)")
	cmd.Flags().String("transaction-at-max", "", "Transaction window end (ISO 8601)")
	cmd.Flags().String("material-site-ids", "", "Comma-separated material site IDs")
	cmd.Flags().String("job-production-plan-id", "", "Job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("raw-job-number")
	_ = cmd.MarkFlagRequired("transaction-at-min")
	_ = cmd.MarkFlagRequired("transaction-at-max")
}

func runDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.RawJobNumber) == "" {
		err := fmt.Errorf("--raw-job-number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TransactionAtMin) == "" {
		err := fmt.Errorf("--transaction-at-min is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TransactionAtMax) == "" {
		err := fmt.Errorf("--transaction-at-max is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"raw-job-number":     opts.RawJobNumber,
		"transaction-at-min": opts.TransactionAtMin,
		"transaction-at-max": opts.TransactionAtMax,
	}

	materialSiteIDs := splitCommaSeparated(opts.MaterialSiteIDs)
	if len(materialSiteIDs) > 0 {
		attributes["material-site-ids"] = materialSiteIDs
	}
	if strings.TrimSpace(opts.JobProductionPlanID) != "" {
		attributes["job-production-plan-id"] = opts.JobProductionPlanID
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tender-job-schedule-shifts-material-transactions-checksums",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-job-schedule-shifts-material-transactions-checksums", jsonBody)
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

func parseDoTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rawJobNumber, _ := cmd.Flags().GetString("raw-job-number")
	transactionAtMin, _ := cmd.Flags().GetString("transaction-at-min")
	transactionAtMax, _ := cmd.Flags().GetString("transaction-at-max")
	materialSiteIDs, _ := cmd.Flags().GetString("material-site-ids")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftsMaterialTransactionsChecksumsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		RawJobNumber:        rawJobNumber,
		TransactionAtMin:    transactionAtMin,
		TransactionAtMax:    transactionAtMax,
		MaterialSiteIDs:     materialSiteIDs,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func splitCommaSeparated(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
