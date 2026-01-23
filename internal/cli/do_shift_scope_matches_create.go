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

type doShiftScopeMatchesCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	Tender                     string
	Rate                       string
	ShiftSetTimeCardConstraint string
	ShiftScope                 string
	TenderJobScheduleShift     string
	ShowMatchingShiftSQL       bool
}

type shiftScopeMatchRow struct {
	ID                           string `json:"id"`
	TenderID                     string `json:"tender_id,omitempty"`
	TenderJobScheduleShiftID     string `json:"tender_job_schedule_shift_id,omitempty"`
	RateID                       string `json:"rate_id,omitempty"`
	ShiftSetTimeCardConstraintID string `json:"shift_set_time_card_constraint_id,omitempty"`
	ShiftScopeID                 string `json:"shift_scope_id,omitempty"`
	Matches                      *bool  `json:"matches,omitempty"`
	MatchSummary                 string `json:"match_summary,omitempty"`
	MatchDetails                 any    `json:"match_details,omitempty"`
	MatchSQL                     string `json:"match_sql,omitempty"`
}

func newDoShiftScopeMatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Match shift scopes",
		Long: `Match shift scopes against a tender.

Required:
  --tender  Tender ID

Provide one of:
  --rate                          Rate ID
  --shift-set-time-card-constraint  Shift set time card constraint ID

Optional:
  --shift-scope                 Shift scope ID (defaults from rate/constraint)
  --tender-job-schedule-shift   Tender job schedule shift ID
  --show-matching-shift-sql     Include match SQL (admins only)`,
		Example: `  # Match a tender against a rate
  xbe do shift-scope-matches create --tender 123 --rate 456

  # Match using a shift set time card constraint
  xbe do shift-scope-matches create --tender 123 --shift-set-time-card-constraint 789

  # Include match SQL when permitted
  xbe do shift-scope-matches create --tender 123 --rate 456 --show-matching-shift-sql

  # JSON output
  xbe do shift-scope-matches create --tender 123 --rate 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoShiftScopeMatchesCreate,
	}
	initDoShiftScopeMatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftScopeMatchesCmd.AddCommand(newDoShiftScopeMatchesCreateCmd())
}

func initDoShiftScopeMatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender", "", "Tender ID")
	cmd.Flags().String("rate", "", "Rate ID")
	cmd.Flags().String("shift-set-time-card-constraint", "", "Shift set time card constraint ID")
	cmd.Flags().String("shift-scope", "", "Shift scope ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().Bool("show-matching-shift-sql", false, "Include match SQL (admins only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftScopeMatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftScopeMatchesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			err := errors.New("authentication required. Run 'xbe auth login' first")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.Tender) == "" {
		err := fmt.Errorf("--tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Rate) == "" && strings.TrimSpace(opts.ShiftSetTimeCardConstraint) == "" {
		err := fmt.Errorf("--rate or --shift-set-time-card-constraint is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Rate) != "" && strings.TrimSpace(opts.ShiftSetTimeCardConstraint) != "" {
		err := fmt.Errorf("--rate and --shift-set-time-card-constraint cannot both be set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("show-matching-shift-sql") {
		attributes["show-matching-shift-sql"] = opts.ShowMatchingShiftSQL
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": "tenders",
				"id":   opts.Tender,
			},
		},
	}

	if strings.TrimSpace(opts.Rate) != "" {
		relationships["rate"] = map[string]any{
			"data": map[string]any{
				"type": "rates",
				"id":   opts.Rate,
			},
		}
	}

	if strings.TrimSpace(opts.ShiftSetTimeCardConstraint) != "" {
		relationships["shift-set-time-card-constraints"] = map[string]any{
			"data": map[string]any{
				"type": "shift-set-time-card-constraints",
				"id":   opts.ShiftSetTimeCardConstraint,
			},
		}
	}

	if strings.TrimSpace(opts.ShiftScope) != "" {
		relationships["shift-scope"] = map[string]any{
			"data": map[string]any{
				"type": "shift-scopes",
				"id":   opts.ShiftScope,
			},
		}
	}

	if strings.TrimSpace(opts.TenderJobScheduleShift) != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "shift-scope-matches",
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

	body, _, err := client.Post(cmd.Context(), "/v1/shift-scope-matches", jsonBody)
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

	row := buildShiftScopeMatchRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderShiftScopeMatch(cmd, row)
}

func parseDoShiftScopeMatchesCreateOptions(cmd *cobra.Command) (doShiftScopeMatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tender, _ := cmd.Flags().GetString("tender")
	rate, _ := cmd.Flags().GetString("rate")
	shiftSetTimeCardConstraint, _ := cmd.Flags().GetString("shift-set-time-card-constraint")
	shiftScope, _ := cmd.Flags().GetString("shift-scope")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	showMatchingShiftSQL, _ := cmd.Flags().GetBool("show-matching-shift-sql")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftScopeMatchesCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		Tender:                     tender,
		Rate:                       rate,
		ShiftSetTimeCardConstraint: shiftSetTimeCardConstraint,
		ShiftScope:                 shiftScope,
		TenderJobScheduleShift:     tenderJobScheduleShift,
		ShowMatchingShiftSQL:       showMatchingShiftSQL,
	}, nil
}

func buildShiftScopeMatchRow(resp jsonAPISingleResponse) shiftScopeMatchRow {
	resource := resp.Data
	attrs := resource.Attributes

	row := shiftScopeMatchRow{
		ID:           resource.ID,
		MatchSummary: strings.TrimSpace(stringAttr(attrs, "match-summary")),
		MatchSQL:     stringAttr(attrs, "match-sql"),
	}

	if attrs != nil {
		if _, ok := attrs["matches"]; ok {
			value := boolAttr(attrs, "matches")
			row.Matches = &value
		}
		if details, ok := attrs["match-details"]; ok && details != nil {
			row.MatchDetails = details
		}
	}

	if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
		row.TenderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["rate"]; ok && rel.Data != nil {
		row.RateID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["shift-scope"]; ok && rel.Data != nil {
		row.ShiftScopeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["shift-set-time-card-constraints"]; ok && rel.Data != nil {
		row.ShiftSetTimeCardConstraintID = rel.Data.ID
	} else if rel, ok := resource.Relationships["shift-set-time-card-constraint"]; ok && rel.Data != nil {
		row.ShiftSetTimeCardConstraintID = rel.Data.ID
	}

	return row
}

func renderShiftScopeMatch(cmd *cobra.Command, row shiftScopeMatchRow) error {
	out := cmd.OutOrStdout()

	if row.ID != "" {
		fmt.Fprintf(out, "Created shift scope match %s\n", row.ID)
	} else {
		fmt.Fprintln(out, "Created shift scope match")
	}

	if row.TenderID != "" {
		fmt.Fprintf(out, "Tender: %s\n", row.TenderID)
	}
	if row.RateID != "" {
		fmt.Fprintf(out, "Rate: %s\n", row.RateID)
	}
	if row.ShiftSetTimeCardConstraintID != "" {
		fmt.Fprintf(out, "Shift Set Time Card Constraint: %s\n", row.ShiftSetTimeCardConstraintID)
	}
	if row.ShiftScopeID != "" {
		fmt.Fprintf(out, "Shift Scope: %s\n", row.ShiftScopeID)
	}
	if row.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", row.TenderJobScheduleShiftID)
	}
	if row.Matches != nil {
		fmt.Fprintf(out, "Matches: %t\n", *row.Matches)
	}
	if row.MatchSummary != "" {
		fmt.Fprintf(out, "Summary: %s\n", row.MatchSummary)
	}

	if row.MatchSQL != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Match SQL:")
		fmt.Fprintln(out, row.MatchSQL)
	}

	if row.MatchDetails != nil {
		pretty, err := json.MarshalIndent(row.MatchDetails, "", "  ")
		if err == nil {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Match Details:")
			fmt.Fprintln(out, string(pretty))
		}
	}

	return nil
}
