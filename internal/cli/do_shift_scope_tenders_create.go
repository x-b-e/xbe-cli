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

type doShiftScopeTendersCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ShiftScopeID                 string
	RateID                       string
	ShiftSetTimeCardConstraintID string
	CreatedAtMin                 string
	CreatedAtMax                 string
	Limit                        int
}

type shiftScopeTenderRow struct {
	ID                           string         `json:"id"`
	ShiftScopeID                 string         `json:"shift_scope_id,omitempty"`
	RateID                       string         `json:"rate_id,omitempty"`
	ShiftSetTimeCardConstraintID string         `json:"shift_set_time_card_constraint_id,omitempty"`
	Limit                        int            `json:"limit,omitempty"`
	Filters                      map[string]any `json:"filters,omitempty"`
	TenderIDs                    []string       `json:"tender_ids,omitempty"`
}

func newDoShiftScopeTendersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Find tenders for a shift scope",
		Long: `Find tenders for a shift scope.

Provide a shift scope directly or derive it from a rate or shift set time card
constraint. You may only provide one of --rate or --shift-set-time-card-constraint.

Filters:
  --created-at-min  Minimum tender created_at (YYYY-MM-DD or RFC3339)
  --created-at-max  Maximum tender created_at (YYYY-MM-DD or RFC3339)

Optional:
  --limit           Max number of tenders to return (default 10)`,
		Example: `  # Find tenders for a shift scope
  xbe do shift-scope-tenders create --shift-scope 123

  # Limit and filter by created_at
  xbe do shift-scope-tenders create --shift-scope 123 \
    --created-at-min 2025-01-01 --created-at-max 2025-01-31 --limit 5

  # Derive shift scope from a rate
  xbe do shift-scope-tenders create --rate 456

  # JSON output
  xbe do shift-scope-tenders create --shift-scope 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoShiftScopeTendersCreate,
	}
	initDoShiftScopeTendersCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftScopeTendersCmd.AddCommand(newDoShiftScopeTendersCreateCmd())
}

func initDoShiftScopeTendersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("shift-scope", "", "Shift scope ID")
	cmd.Flags().String("rate", "", "Rate ID (derives shift scope)")
	cmd.Flags().String("shift-set-time-card-constraint", "", "Shift set time card constraint ID (derives shift scope)")
	cmd.Flags().String("created-at-min", "", "Minimum tender created_at (YYYY-MM-DD or RFC3339)")
	cmd.Flags().String("created-at-max", "", "Maximum tender created_at (YYYY-MM-DD or RFC3339)")
	cmd.Flags().Int("limit", 0, "Max number of tenders to return")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftScopeTendersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoShiftScopeTendersCreateOptions(cmd)
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

	if opts.RateID != "" && opts.ShiftSetTimeCardConstraintID != "" {
		err := fmt.Errorf("only one of --rate or --shift-set-time-card-constraint may be set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ShiftScopeID == "" && opts.RateID == "" && opts.ShiftSetTimeCardConstraintID == "" {
		err := fmt.Errorf("--shift-scope, --rate, or --shift-set-time-card-constraint is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	filters := map[string]any{}
	if opts.CreatedAtMin != "" {
		filters["created_at_min"] = opts.CreatedAtMin
	}
	if opts.CreatedAtMax != "" {
		filters["created_at_max"] = opts.CreatedAtMax
	}

	attributes := map[string]any{
		"filters": filters,
	}
	if opts.Limit > 0 {
		attributes["limit"] = opts.Limit
	}

	relationships := map[string]any{}
	if opts.ShiftScopeID != "" {
		relationships["shift-scope"] = map[string]any{
			"data": map[string]any{
				"type": "shift-scopes",
				"id":   opts.ShiftScopeID,
			},
		}
	}
	if opts.RateID != "" {
		relationships["rate"] = map[string]any{
			"data": map[string]any{
				"type": "rates",
				"id":   opts.RateID,
			},
		}
	}
	if opts.ShiftSetTimeCardConstraintID != "" {
		relationships["shift-set-time-card-constraint"] = map[string]any{
			"data": map[string]any{
				"type": "shift-set-time-card-constraints",
				"id":   opts.ShiftSetTimeCardConstraintID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "shift-scope-tenders",
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

	body, _, err := client.Post(cmd.Context(), "/v1/shift-scope-tenders", jsonBody)
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

	row := buildShiftScopeTenderRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderShiftScopeTender(cmd, row)
}

func parseDoShiftScopeTendersCreateOptions(cmd *cobra.Command) (doShiftScopeTendersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	shiftScopeID, _ := cmd.Flags().GetString("shift-scope")
	rateID, _ := cmd.Flags().GetString("rate")
	shiftSetTimeCardConstraintID, _ := cmd.Flags().GetString("shift-set-time-card-constraint")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	limit, _ := cmd.Flags().GetInt("limit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftScopeTendersCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ShiftScopeID:                 shiftScopeID,
		RateID:                       rateID,
		ShiftSetTimeCardConstraintID: shiftSetTimeCardConstraintID,
		CreatedAtMin:                 createdAtMin,
		CreatedAtMax:                 createdAtMax,
		Limit:                        limit,
	}, nil
}

func buildShiftScopeTenderRowFromSingle(resp jsonAPISingleResponse) shiftScopeTenderRow {
	resource := resp.Data
	row := shiftScopeTenderRow{
		ID:                           resource.ID,
		Limit:                        intAttr(resource.Attributes, "limit"),
		ShiftScopeID:                 relationshipIDFromMap(resource.Relationships, "shift-scope"),
		RateID:                       relationshipIDFromMap(resource.Relationships, "rate"),
		ShiftSetTimeCardConstraintID: relationshipIDFromMap(resource.Relationships, "shift-set-time-card-constraint"),
		TenderIDs:                    relationshipIDsFromMap(resource.Relationships, "tenders"),
	}

	if attrs := resource.Attributes; attrs != nil {
		if raw, ok := attrs["filters"]; ok && raw != nil {
			if typed, ok := raw.(map[string]any); ok {
				row.Filters = typed
			} else if typed, ok := raw.(map[string]interface{}); ok {
				row.Filters = typed
			}
		}
	}

	return row
}

func renderShiftScopeTender(cmd *cobra.Command, row shiftScopeTenderRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Shift scope tenders %s\n", row.ID)
	if row.ShiftScopeID != "" {
		fmt.Fprintf(out, "Shift scope: %s\n", row.ShiftScopeID)
	}
	if row.RateID != "" {
		fmt.Fprintf(out, "Rate: %s\n", row.RateID)
	}
	if row.ShiftSetTimeCardConstraintID != "" {
		fmt.Fprintf(out, "Shift set time card constraint: %s\n", row.ShiftSetTimeCardConstraintID)
	}
	if row.Limit > 0 {
		fmt.Fprintf(out, "Limit: %d\n", row.Limit)
	}
	if len(row.Filters) > 0 {
		fmt.Fprintln(out, "Filters:")
		fmt.Fprintln(out, formatJSONBlock(row.Filters, "  "))
	}

	if len(row.TenderIDs) == 0 {
		fmt.Fprintln(out, "Tenders: none")
		return nil
	}

	fmt.Fprintf(out, "Tenders (%d):\n", len(row.TenderIDs))
	for _, tenderID := range row.TenderIDs {
		fmt.Fprintf(out, "  %s\n", tenderID)
	}
	return nil
}
