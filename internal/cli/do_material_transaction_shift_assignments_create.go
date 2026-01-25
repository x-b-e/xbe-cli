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

type doMaterialTransactionShiftAssignmentsCreateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	MaterialTransactionIDs                     []string
	TenderJobScheduleShift                     string
	SkipMaterialTransactionShiftSkewValidation bool
	EnableLinkInvoiced                         bool
	Comment                                    string
}

func newDoMaterialTransactionShiftAssignmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction shift assignment",
		Long: `Create a material transaction shift assignment.

Required flags:
  --tender-job-schedule-shift   Tender job schedule shift ID (required)
  --material-transaction-ids    Material transaction IDs (required, comma-separated or repeated)

Optional flags:
  --comment                                           Assignment comment
  --skip-material-transaction-shift-skew-validation   Skip shift skew validation
  --enable-link-invoiced                              Enable linking invoiced material transactions`,
		Example: `  # Assign material transactions to a shift
  xbe do material-transaction-shift-assignments create \
    --tender-job-schedule-shift 123 \
    --material-transaction-ids 456,789

  # Assign with comment and validation override
  xbe do material-transaction-shift-assignments create \
    --tender-job-schedule-shift 123 \
    --material-transaction-ids 456,789 \
    --comment "Reassign for shift" \
    --skip-material-transaction-shift-skew-validation`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionShiftAssignmentsCreate,
	}
	initDoMaterialTransactionShiftAssignmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionShiftAssignmentsCmd.AddCommand(newDoMaterialTransactionShiftAssignmentsCreateCmd())
}

func initDoMaterialTransactionShiftAssignmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().StringSlice("material-transaction-ids", nil, "Material transaction IDs (required, comma-separated or repeated)")
	cmd.Flags().String("comment", "", "Assignment comment")
	cmd.Flags().Bool("skip-material-transaction-shift-skew-validation", false, "Skip shift skew validation")
	cmd.Flags().Bool("enable-link-invoiced", false, "Enable linking invoiced material transactions")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionShiftAssignmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionShiftAssignmentsCreateOptions(cmd)
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

	opts.TenderJobScheduleShift = strings.TrimSpace(opts.TenderJobScheduleShift)
	if opts.TenderJobScheduleShift == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	materialTransactionIDs := make([]string, 0, len(opts.MaterialTransactionIDs))
	for _, id := range opts.MaterialTransactionIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			materialTransactionIDs = append(materialTransactionIDs, trimmed)
		}
	}
	if len(materialTransactionIDs) == 0 {
		err := fmt.Errorf("--material-transaction-ids is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"material-transaction-ids": materialTransactionIDs,
	}
	if opts.SkipMaterialTransactionShiftSkewValidation {
		attributes["skip-material-transaction-shift-skew-validation"] = true
	}
	if opts.EnableLinkInvoiced {
		attributes["enable-link-invoiced"] = true
	}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-transaction-shift-assignments",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-shift-assignments", jsonBody)
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

	row := materialTransactionShiftAssignmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material transaction shift assignment %s\n", row.ID)
	return nil
}

func parseDoMaterialTransactionShiftAssignmentsCreateOptions(cmd *cobra.Command) (doMaterialTransactionShiftAssignmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialTransactionIDs, _ := cmd.Flags().GetStringSlice("material-transaction-ids")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	skipMaterialTransactionShiftSkewValidation, _ := cmd.Flags().GetBool("skip-material-transaction-shift-skew-validation")
	enableLinkInvoiced, _ := cmd.Flags().GetBool("enable-link-invoiced")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTransactionShiftAssignmentsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		MaterialTransactionIDs: materialTransactionIDs,
		TenderJobScheduleShift: tenderJobScheduleShift,
		SkipMaterialTransactionShiftSkewValidation: skipMaterialTransactionShiftSkewValidation,
		EnableLinkInvoiced:                         enableLinkInvoiced,
		Comment:                                    comment,
	}, nil
}
