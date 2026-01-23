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

type doCommitmentItemsUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
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

func newDoCommitmentItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a commitment item",
		Long: `Update a commitment item.

Use flags to specify which fields to update. Only specified fields are modified.

Updatable attributes:
  --label                        Update label
  --status                       Update status (editing, active, inactive)
  --start-on                     Update start date (YYYY-MM-DD)
  --end-on                       Update end date (YYYY-MM-DD)
  --years                        Update years JSON array
  --months                       Update months JSON array
  --weeks                        Update weeks JSON array
  --days-of-week                 Update days of week JSON array
  --times-of-day                 Update times of day JSON array
  --adjustment-sequence-position Update adjustment sequence position
  --adjustment-coefficient       Update adjustment coefficient
  --adjustment-constant           Update adjustment constant (ProbabilityDistribution JSON object)
  --adjustment-input              Update adjustment input (ProbabilityDistribution JSON object)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update label and status
  xbe do commitment-items update 123 --label "Updated" --status active

  # Update schedule
  xbe do commitment-items update 123 --days-of-week '[1,2,3]' --times-of-day '["day"]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCommitmentItemsUpdate,
	}
	initDoCommitmentItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCommitmentItemsCmd.AddCommand(newDoCommitmentItemsUpdateCmd())
}

func initDoCommitmentItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
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
}

func runDoCommitmentItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCommitmentItemsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("label") {
		attributes["label"] = opts.Label
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("years") {
		values, err := parseCommitmentItemIntList(opts.Years, "years")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["years"] = values
	}
	if cmd.Flags().Changed("months") {
		values, err := parseCommitmentItemIntList(opts.Months, "months")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["months"] = values
	}
	if cmd.Flags().Changed("weeks") {
		values, err := parseCommitmentItemIntList(opts.Weeks, "weeks")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["weeks"] = values
	}
	if cmd.Flags().Changed("days-of-week") {
		values, err := parseCommitmentItemIntList(opts.DaysOfWeek, "days-of-week")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["days-of-week"] = values
	}
	if cmd.Flags().Changed("times-of-day") {
		values, err := parseCommitmentItemStringList(opts.TimesOfDay, "times-of-day")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["times-of-day"] = values
	}
	if cmd.Flags().Changed("adjustment-sequence-position") {
		attributes["adjustment-sequence-position"] = opts.AdjustmentSequencePosition
	}
	if cmd.Flags().Changed("adjustment-coefficient") {
		attributes["adjustment-coefficient"] = opts.AdjustmentCoefficient
	}
	if cmd.Flags().Changed("adjustment-constant") {
		value, err := parseCommitmentItemJSONObject(opts.AdjustmentConstant, "adjustment-constant")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["adjustment-constant"] = value
	}
	if cmd.Flags().Changed("adjustment-input") {
		value, err := parseCommitmentItemJSONObject(opts.AdjustmentInput, "adjustment-input")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["adjustment-input"] = value
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "commitment-items",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/commitment-items/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated commitment item %s\n", row.ID)
	return nil
}

func parseDoCommitmentItemsUpdateOptions(cmd *cobra.Command, args []string) (doCommitmentItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
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

	return doCommitmentItemsUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
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
