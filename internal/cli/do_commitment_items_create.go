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

type doCommitmentItemsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	CommitmentType             string
	CommitmentID               string
	StartOn                    string
	EndOn                      string
	Years                      string
	Months                     string
	Weeks                      string
	DaysOfWeek                 string
	TimesOfDay                 string
	AdjustmentSequencePosition int
	Label                      string
	Status                     string
	AdjustmentConstant         string
	AdjustmentCoefficient      float64
	AdjustmentInput            string
}

func newDoCommitmentItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a commitment item",
		Long: `Create a commitment item.

Required flags:
  --commitment-type  Commitment type (required, e.g., customer-commitments, broker-commitments)
  --commitment-id    Commitment ID (required)

Optional attributes:
  --label                        Label for the commitment item
  --status                       Status (editing, active, inactive)
  --start-on                     Start date (YYYY-MM-DD)
  --end-on                       End date (YYYY-MM-DD)
  --years                        Years JSON array (e.g., [2026,2027])
  --months                       Months JSON array (1-12)
  --weeks                        Weeks JSON array (1-53)
  --days-of-week                 Days of week JSON array (0-6, Sunday=0)
  --times-of-day                 Times of day JSON array ("day", "night")
  --adjustment-sequence-position Adjustment sequence position
  --adjustment-coefficient       Adjustment coefficient (default 1.0)
  --adjustment-constant           Adjustment constant (ProbabilityDistribution JSON object)
  --adjustment-input              Adjustment input (ProbabilityDistribution JSON object)

Probability distribution JSON example:
  {"class_name":"NormalDistribution","mean":1.0,"standard_deviation":0.2}

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a commitment item
  xbe do commitment-items create \
    --commitment-type customer-commitments \
    --commitment-id 123 \
    --label "Daytime schedule" \
    --status editing \
    --start-on 2026-01-01 \
    --end-on 2026-12-31 \
    --days-of-week '[1,2,3,4,5]' \
    --times-of-day '["day"]' \
    --adjustment-coefficient 1.05

  # Create with adjustment distribution
  xbe do commitment-items create \
    --commitment-type broker-commitments \
    --commitment-id 456 \
    --adjustment-constant '{"class_name":"NormalDistribution","mean":1.0,"standard_deviation":0.2}'`,
		Args: cobra.NoArgs,
		RunE: runDoCommitmentItemsCreate,
	}
	initDoCommitmentItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doCommitmentItemsCmd.AddCommand(newDoCommitmentItemsCreateCmd())
}

func initDoCommitmentItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("commitment-type", "", "Commitment type (required)")
	cmd.Flags().String("commitment-id", "", "Commitment ID (required)")
	cmd.Flags().String("label", "", "Commitment item label")
	cmd.Flags().String("status", "", "Status (editing, active, inactive)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("years", "", "Years JSON array (e.g., [2026,2027])")
	cmd.Flags().String("months", "", "Months JSON array (1-12)")
	cmd.Flags().String("weeks", "", "Weeks JSON array (1-53)")
	cmd.Flags().String("days-of-week", "", "Days of week JSON array (0-6, Sunday=0)")
	cmd.Flags().String("times-of-day", "", "Times of day JSON array (\"day\", \"night\")")
	cmd.Flags().Int("adjustment-sequence-position", 0, "Adjustment sequence position")
	cmd.Flags().Float64("adjustment-coefficient", 0, "Adjustment coefficient")
	cmd.Flags().String("adjustment-constant", "", "Adjustment constant (ProbabilityDistribution JSON object)")
	cmd.Flags().String("adjustment-input", "", "Adjustment input (ProbabilityDistribution JSON object)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("commitment-type")
	_ = cmd.MarkFlagRequired("commitment-id")
}

func runDoCommitmentItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCommitmentItemsCreateOptions(cmd)
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

	if opts.CommitmentType == "" {
		err := fmt.Errorf("--commitment-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CommitmentID == "" {
		err := fmt.Errorf("--commitment-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Label != "" {
		attributes["label"] = opts.Label
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.EndOn != "" {
		attributes["end-on"] = opts.EndOn
	}
	if opts.Years != "" {
		years, err := parseCommitmentItemIntList(opts.Years, "years")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["years"] = years
	}
	if opts.Months != "" {
		months, err := parseCommitmentItemIntList(opts.Months, "months")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["months"] = months
	}
	if opts.Weeks != "" {
		weeks, err := parseCommitmentItemIntList(opts.Weeks, "weeks")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["weeks"] = weeks
	}
	if opts.DaysOfWeek != "" {
		days, err := parseCommitmentItemIntList(opts.DaysOfWeek, "days-of-week")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["days-of-week"] = days
	}
	if opts.TimesOfDay != "" {
		times, err := parseCommitmentItemStringList(opts.TimesOfDay, "times-of-day")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["times-of-day"] = times
	}
	if cmd.Flags().Changed("adjustment-sequence-position") {
		attributes["adjustment-sequence-position"] = opts.AdjustmentSequencePosition
	}
	if cmd.Flags().Changed("adjustment-coefficient") {
		attributes["adjustment-coefficient"] = opts.AdjustmentCoefficient
	}
	if opts.AdjustmentConstant != "" {
		value, err := parseCommitmentItemJSONObject(opts.AdjustmentConstant, "adjustment-constant")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["adjustment-constant"] = value
	}
	if opts.AdjustmentInput != "" {
		value, err := parseCommitmentItemJSONObject(opts.AdjustmentInput, "adjustment-input")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["adjustment-input"] = value
	}

	relationships := map[string]any{
		"commitment": map[string]any{
			"data": map[string]any{
				"type": opts.CommitmentType,
				"id":   opts.CommitmentID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "commitment-items",
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

	body, _, err := client.Post(cmd.Context(), "/v1/commitment-items", jsonBody)
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

	row := buildCommitmentItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created commitment item %s\n", row.ID)
	return nil
}

func parseDoCommitmentItemsCreateOptions(cmd *cobra.Command) (doCommitmentItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	commitmentType, _ := cmd.Flags().GetString("commitment-type")
	commitmentID, _ := cmd.Flags().GetString("commitment-id")
	label, _ := cmd.Flags().GetString("label")
	status, _ := cmd.Flags().GetString("status")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	years, _ := cmd.Flags().GetString("years")
	months, _ := cmd.Flags().GetString("months")
	weeks, _ := cmd.Flags().GetString("weeks")
	daysOfWeek, _ := cmd.Flags().GetString("days-of-week")
	timesOfDay, _ := cmd.Flags().GetString("times-of-day")
	adjustmentSequencePosition, _ := cmd.Flags().GetInt("adjustment-sequence-position")
	adjustmentCoefficient, _ := cmd.Flags().GetFloat64("adjustment-coefficient")
	adjustmentConstant, _ := cmd.Flags().GetString("adjustment-constant")
	adjustmentInput, _ := cmd.Flags().GetString("adjustment-input")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCommitmentItemsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		CommitmentType:             commitmentType,
		CommitmentID:               commitmentID,
		Label:                      label,
		Status:                     status,
		StartOn:                    startOn,
		EndOn:                      endOn,
		Years:                      years,
		Months:                     months,
		Weeks:                      weeks,
		DaysOfWeek:                 daysOfWeek,
		TimesOfDay:                 timesOfDay,
		AdjustmentSequencePosition: adjustmentSequencePosition,
		AdjustmentCoefficient:      adjustmentCoefficient,
		AdjustmentConstant:         adjustmentConstant,
		AdjustmentInput:            adjustmentInput,
	}, nil
}
