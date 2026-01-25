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

type doTenderAcceptancesCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	TenderID                          string
	TenderType                        string
	Comment                           string
	SkipCertificationValidation       bool
	RejectedTenderJobScheduleShiftIDs []string
}

func newDoTenderAcceptancesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Accept a tender",
		Long: `Accept a tender.

Required flags:
  --tender-type   Tender type (customer-tenders, broker-tenders) (required)
  --tender        Tender ID (required)

Optional flags:
  --comment                             Acceptance comment
  --skip-certification-validation       Skip certification requirement validation
  --rejected-tender-job-schedule-shift-ids  Tender job schedule shift IDs to reject (comma-separated or repeated)`,
		Example: `  # Accept a customer tender
  xbe do tender-acceptances create --tender-type customer-tenders --tender 12345

  # Accept with comment and skip certification validation
  xbe do tender-acceptances create --tender-type broker-tenders --tender 67890 \\
    --comment "Accepted by dispatch" --skip-certification-validation

  # Accept and reject one shift
  xbe do tender-acceptances create --tender-type broker-tenders --tender 67890 \\
    --rejected-tender-job-schedule-shift-ids 111,112

  # JSON output
  xbe do tender-acceptances create --tender-type customer-tenders --tender 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderAcceptancesCreate,
	}
	initDoTenderAcceptancesCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderAcceptancesCmd.AddCommand(newDoTenderAcceptancesCreateCmd())
}

func initDoTenderAcceptancesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-type", "", "Tender type (customer-tenders, broker-tenders) (required)")
	cmd.Flags().String("tender", "", "Tender ID (required)")
	cmd.Flags().String("comment", "", "Acceptance comment")
	cmd.Flags().Bool("skip-certification-validation", false, "Skip certification requirement validation")
	cmd.Flags().StringSlice("rejected-tender-job-schedule-shift-ids", nil, "Tender job schedule shift IDs to reject (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderAcceptancesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderAcceptancesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TenderType) == "" {
		err := fmt.Errorf("--tender-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TenderID) == "" {
		err := fmt.Errorf("--tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("skip-certification-validation") {
		attributes["skip-certification-validation"] = opts.SkipCertificationValidation
	}
	if len(opts.RejectedTenderJobScheduleShiftIDs) > 0 {
		attributes["rejected-tender-job-schedule-shift-ids"] = opts.RejectedTenderJobScheduleShiftIDs
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": opts.TenderType,
				"id":   opts.TenderID,
			},
		},
	}

	data := map[string]any{
		"type":          "tender-acceptances",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-acceptances", jsonBody)
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

	row := buildTenderAcceptanceRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender acceptance %s\n", row.ID)
	return nil
}

func parseDoTenderAcceptancesCreateOptions(cmd *cobra.Command) (doTenderAcceptancesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderType, _ := cmd.Flags().GetString("tender-type")
	tenderID, _ := cmd.Flags().GetString("tender")
	comment, _ := cmd.Flags().GetString("comment")
	skipCertificationValidation, _ := cmd.Flags().GetBool("skip-certification-validation")
	rejectedTenderJobScheduleShiftIDs, _ := cmd.Flags().GetStringSlice("rejected-tender-job-schedule-shift-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderAcceptancesCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		TenderID:                          tenderID,
		TenderType:                        tenderType,
		Comment:                           comment,
		SkipCertificationValidation:       skipCertificationValidation,
		RejectedTenderJobScheduleShiftIDs: rejectedTenderJobScheduleShiftIDs,
	}, nil
}
