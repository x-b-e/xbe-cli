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

type doShiftFeedbacksCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	Rating                   int
	Note                     string
	TenderJobScheduleShiftID string
	ReasonID                 string
}

func newDoShiftFeedbacksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new shift feedback",
		Long: `Create a new shift feedback.

Required flags:
  --tender-job-schedule-shift   Tender job schedule shift ID (required)
  --reason                      Shift feedback reason ID (required)
  --rating                      Rating (required)

Optional flags:
  --note                        Feedback note`,
		Example: `  # Create a shift feedback
  xbe do shift-feedbacks create \
    --tender-job-schedule-shift 123 \
    --reason 456 \
    --rating 5

  # Create with a note
  xbe do shift-feedbacks create \
    --tender-job-schedule-shift 123 \
    --reason 456 \
    --rating 5 \
    --note "Great work on this shift"`,
		Args: cobra.NoArgs,
		RunE: runDoShiftFeedbacksCreate,
	}
	initDoShiftFeedbacksCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftFeedbacksCmd.AddCommand(newDoShiftFeedbacksCreateCmd())
}

func initDoShiftFeedbacksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("rating", 0, "Rating (required)")
	cmd.Flags().String("note", "", "Feedback note")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("reason", "", "Shift feedback reason ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftFeedbacksCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftFeedbacksCreateOptions(cmd)
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

	if opts.TenderJobScheduleShiftID == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ReasonID == "" {
		err := fmt.Errorf("--reason is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("rating") {
		err := fmt.Errorf("--rating is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"rating": opts.Rating,
	}

	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShiftID,
			},
		},
		"reason": map[string]any{
			"data": map[string]any{
				"type": "shift-feedback-reasons",
				"id":   opts.ReasonID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "shift-feedbacks",
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

	body, _, err := client.Post(cmd.Context(), "/v1/shift-feedbacks", jsonBody)
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

	row := buildShiftFeedbackRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created shift feedback %s\n", row.ID)
	return nil
}

func parseDoShiftFeedbacksCreateOptions(cmd *cobra.Command) (doShiftFeedbacksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rating, _ := cmd.Flags().GetInt("rating")
	note, _ := cmd.Flags().GetString("note")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	reasonID, _ := cmd.Flags().GetString("reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doShiftFeedbacksCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		Rating:                   rating,
		Note:                     note,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		ReasonID:                 reasonID,
	}, nil
}

func buildShiftFeedbackRowFromSingle(resp jsonAPISingleResponse) shiftFeedbackRow {
	attrs := resp.Data.Attributes

	row := shiftFeedbackRow{
		ID:           resp.Data.ID,
		Rating:       intAttr(attrs, "rating"),
		Note:         stringAttr(attrs, "note"),
		CreatedByBot: boolAttr(attrs, "created-by-bot"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["reason"]; ok && rel.Data != nil {
		row.ReasonID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}

	return row
}
