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

type doTenderRatersCreateOptions struct {
	BaseURL                                        string
	Token                                          string
	JSON                                           bool
	Tender                                         string
	ReplaceRates                                   string
	ReplaceShiftSetTimeCardConstraints             string
	PersistChanges                                 string
	SkipAdjustmentCostIndexValuePresenceValidation string
	SkipValidateCustomerTenderHourlyRates          string
}

type tenderRaterRow struct {
	ID                                             string `json:"id"`
	TenderID                                       string `json:"tender_id,omitempty"`
	ReplaceRates                                   bool   `json:"replace_rates"`
	ReplaceShiftSetTimeCardConstraints             bool   `json:"replace_shift_set_time_card_constraints"`
	PersistChanges                                 bool   `json:"persist_changes"`
	SkipAdjustmentCostIndexValuePresenceValidation bool   `json:"skip_adjustment_cost_index_value_presence_validation"`
	SkipValidateCustomerTenderHourlyRates          bool   `json:"skip_validate_customer_tender_hourly_rates"`
}

func newDoTenderRatersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Rate a tender",
		Long: `Rate a tender against applicable rate agreements.

Required flags:
  --tender  Tender ID (required)

Optional flags:
  --replace-rates                                      Replace existing rates (true/false)
  --replace-shift-set-time-card-constraints            Replace existing shift set time card constraints (true/false)
  --persist-changes                                    Persist generated rates/constraints (true/false)
  --skip-adjustment-cost-index-value-presence-validation  Skip adjustment cost index validation (true/false)
  --skip-validate-customer-tender-hourly-rates         Skip customer tender hourly rate validation (true/false)`,
		Example: `  # Rate a tender without persisting changes
  xbe do tender-raters create --tender 123

  # Rate a tender and persist changes
  xbe do tender-raters create --tender 123 --persist-changes true

  # Output as JSON
  xbe do tender-raters create --tender 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTenderRatersCreate,
	}
	initDoTenderRatersCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderRatersCmd.AddCommand(newDoTenderRatersCreateCmd())
}

func initDoTenderRatersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender", "", "Tender ID (required)")
	cmd.Flags().String("replace-rates", "", "Replace existing rates (true/false)")
	cmd.Flags().String("replace-shift-set-time-card-constraints", "", "Replace shift set time card constraints (true/false)")
	cmd.Flags().String("persist-changes", "", "Persist generated rates/constraints (true/false)")
	cmd.Flags().String("skip-adjustment-cost-index-value-presence-validation", "", "Skip adjustment cost index validation (true/false)")
	cmd.Flags().String("skip-validate-customer-tender-hourly-rates", "", "Skip customer tender hourly rate validation (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderRatersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderRatersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	opts.Tender = strings.TrimSpace(opts.Tender)
	if opts.Tender == "" {
		err := fmt.Errorf("--tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.ReplaceRates) != "" {
		value, err := parseTenderRaterBool(opts.ReplaceRates, "replace-rates")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["replace-rates"] = value
	}
	if strings.TrimSpace(opts.ReplaceShiftSetTimeCardConstraints) != "" {
		value, err := parseTenderRaterBool(opts.ReplaceShiftSetTimeCardConstraints, "replace-shift-set-time-card-constraints")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["replace-shift-set-time-card-constraints"] = value
	}
	if strings.TrimSpace(opts.PersistChanges) != "" {
		value, err := parseTenderRaterBool(opts.PersistChanges, "persist-changes")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["persist-changes"] = value
	}
	if strings.TrimSpace(opts.SkipAdjustmentCostIndexValuePresenceValidation) != "" {
		value, err := parseTenderRaterBool(opts.SkipAdjustmentCostIndexValuePresenceValidation, "skip-adjustment-cost-index-value-presence-validation")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["skip-adjustment-cost-index-value-presence-validation"] = value
	}
	if strings.TrimSpace(opts.SkipValidateCustomerTenderHourlyRates) != "" {
		value, err := parseTenderRaterBool(opts.SkipValidateCustomerTenderHourlyRates, "skip-validate-customer-tender-hourly-rates")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["skip-validate-customer-tender-hourly-rates"] = value
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": "tenders",
				"id":   opts.Tender,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-raters",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tender-raters", jsonBody)
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

	row := tenderRaterRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender rater %s\n", row.ID)
	return nil
}

func tenderRaterRowFromSingle(resp jsonAPISingleResponse) tenderRaterRow {
	attrs := resp.Data.Attributes
	row := tenderRaterRow{
		ID:                                 resp.Data.ID,
		ReplaceRates:                       boolAttr(attrs, "replace-rates"),
		ReplaceShiftSetTimeCardConstraints: boolAttr(attrs, "replace-shift-set-time-card-constraints"),
		PersistChanges:                     boolAttr(attrs, "persist-changes"),
		SkipAdjustmentCostIndexValuePresenceValidation: boolAttr(attrs, "skip-adjustment-cost-index-value-presence-validation"),
		SkipValidateCustomerTenderHourlyRates:          boolAttr(attrs, "skip-validate-customer-tender-hourly-rates"),
	}

	if rel, ok := resp.Data.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderID = rel.Data.ID
	}

	return row
}

func parseTenderRaterBool(value string, flagName string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("--%s must be true or false", flagName)
	}
}

func parseDoTenderRatersCreateOptions(cmd *cobra.Command) (doTenderRatersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tender, _ := cmd.Flags().GetString("tender")
	replaceRates, _ := cmd.Flags().GetString("replace-rates")
	replaceConstraints, _ := cmd.Flags().GetString("replace-shift-set-time-card-constraints")
	persistChanges, _ := cmd.Flags().GetString("persist-changes")
	skipAdjustmentCostIndexValuePresenceValidation, _ := cmd.Flags().GetString("skip-adjustment-cost-index-value-presence-validation")
	skipValidateCustomerTenderHourlyRates, _ := cmd.Flags().GetString("skip-validate-customer-tender-hourly-rates")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderRatersCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Tender:                             tender,
		ReplaceRates:                       replaceRates,
		ReplaceShiftSetTimeCardConstraints: replaceConstraints,
		PersistChanges:                     persistChanges,
		SkipAdjustmentCostIndexValuePresenceValidation: skipAdjustmentCostIndexValuePresenceValidation,
		SkipValidateCustomerTenderHourlyRates:          skipValidateCustomerTenderHourlyRates,
	}, nil
}
