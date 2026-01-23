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

type doTimeSheetsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	SubjectType         string
	SubjectID           string
	DriverID            string
	StartAt             string
	EndAt               string
	BreakMinutes        int
	Notes               string
	SkipValidateOverlap bool
}

func newDoTimeSheetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time sheet",
		Long: `Create a time sheet.

Required flags:
  --subject-type       Subject type (WorkOrder, CrewRequirement, TruckerShiftSet; also accepts work-orders, crew-requirements, trucker-shift-sets)
  --subject-id         Subject ID

Optional flags:
  --driver             Driver user ID (create only)
  --start-at           Start timestamp (ISO 8601)
  --end-at             End timestamp (ISO 8601)
  --break-minutes      Break minutes
  --notes              Notes
  --skip-validate-overlap  Skip overlap validation (true/false)`,
		Example: `  # Create a time sheet for a work order
  xbe do time-sheets create \\
    --subject-type WorkOrder \\
    --subject-id 123 \\
    --start-at 2026-01-01T08:00:00Z \\
    --end-at 2026-01-01T16:00:00Z \\
    --break-minutes 30

  # Create with a driver and notes
  xbe do time-sheets create \\
    --subject-type WorkOrder \\
    --subject-id 123 \\
    --driver 456 \\
    --notes \"Shift coverage\"`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetsCreate,
	}
	initDoTimeSheetsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetsCmd.AddCommand(newDoTimeSheetsCreateCmd())
}

func initDoTimeSheetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("subject-type", "", "Subject type (WorkOrder, CrewRequirement, TruckerShiftSet; also accepts work-orders, crew-requirements, trucker-shift-sets) (required)")
	cmd.Flags().String("subject-id", "", "Subject ID (required)")
	cmd.Flags().String("driver", "", "Driver user ID")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().Int("break-minutes", 0, "Break minutes")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().Bool("skip-validate-overlap", false, "Skip overlap validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTimeSheetsCreateOptions(cmd)
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

	if opts.SubjectType == "" {
		err := fmt.Errorf("--subject-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.SubjectID == "" {
		err := fmt.Errorf("--subject-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("break-minutes") {
		attributes["break-minutes"] = opts.BreakMinutes
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("skip-validate-overlap") {
		attributes["skip-validate-overlap"] = opts.SkipValidateOverlap
	}

	relationships := map[string]any{
		"subject": map[string]any{
			"data": map[string]any{
				"type": normalizeTimeSheetSubjectRelationship(opts.SubjectType),
				"id":   opts.SubjectID,
			},
		},
	}
	if opts.DriverID != "" {
		relationships["driver"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.DriverID,
			},
		}
	}

	data := map[string]any{
		"type":          "time-sheets",
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

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheets", jsonBody)
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

	row := buildTimeSheetRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet %s\n", row.ID)
	return nil
}

func parseDoTimeSheetsCreateOptions(cmd *cobra.Command) (doTimeSheetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	driverID, _ := cmd.Flags().GetString("driver")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	breakMinutes, _ := cmd.Flags().GetInt("break-minutes")
	notes, _ := cmd.Flags().GetString("notes")
	skipValidateOverlap, _ := cmd.Flags().GetBool("skip-validate-overlap")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		SubjectType:         subjectType,
		SubjectID:           subjectID,
		DriverID:            driverID,
		StartAt:             startAt,
		EndAt:               endAt,
		BreakMinutes:        breakMinutes,
		Notes:               notes,
		SkipValidateOverlap: skipValidateOverlap,
	}, nil
}
