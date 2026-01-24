package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type expectedTimeOfArrivalsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type expectedTimeOfArrivalDetails struct {
	ID                     string `json:"id"`
	ExpectedAt             string `json:"expected_at,omitempty"`
	Note                   string `json:"note,omitempty"`
	Unsure                 bool   `json:"unsure"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	JobScheduleShift       string `json:"job_schedule_shift_id,omitempty"`
	CreatedBy              string `json:"created_by_id,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

func newExpectedTimeOfArrivalsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show expected time of arrival details",
		Long: `Show the full details of a specific expected time of arrival update.

Output Fields:
  ID            Expected time of arrival identifier
  EXPECTED AT   Expected arrival timestamp
  UNSURE        Whether the arrival time is unsure
  NOTE          Notes for the arrival
  TENDER SHIFT  Tender job schedule shift ID
  JOB SHIFT     Job schedule shift ID
  CREATED BY    User who created the record
  CREATED AT    Record creation timestamp
  UPDATED AT    Record update timestamp

Arguments:
  <id>  Expected time of arrival ID (required). Find IDs using the list command.`,
		Example: `  # View an expected time of arrival by ID
  xbe view expected-time-of-arrivals show 123

  # Get JSON output
  xbe view expected-time-of-arrivals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runExpectedTimeOfArrivalsShow,
	}
	initExpectedTimeOfArrivalsShowFlags(cmd)
	return cmd
}

func init() {
	expectedTimeOfArrivalsCmd.AddCommand(newExpectedTimeOfArrivalsShowCmd())
}

func initExpectedTimeOfArrivalsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExpectedTimeOfArrivalsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseExpectedTimeOfArrivalsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("expected time of arrival id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[expected-time-of-arrivals]", "expected-at,note,unsure,created-at,updated-at,tender-job-schedule-shift,job-schedule-shift,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/expected-time-of-arrivals/"+id, query)
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

	details := buildExpectedTimeOfArrivalDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderExpectedTimeOfArrivalDetails(cmd, details)
}

func parseExpectedTimeOfArrivalsShowOptions(cmd *cobra.Command) (expectedTimeOfArrivalsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return expectedTimeOfArrivalsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildExpectedTimeOfArrivalDetails(resp jsonAPISingleResponse) expectedTimeOfArrivalDetails {
	attrs := resp.Data.Attributes

	details := expectedTimeOfArrivalDetails{
		ID:         resp.Data.ID,
		ExpectedAt: formatDateTime(stringAttr(attrs, "expected-at")),
		Note:       stringAttr(attrs, "note"),
		Unsure:     boolAttr(attrs, "unsure"),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		details.JobScheduleShift = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedBy = rel.Data.ID
	}

	return details
}

func renderExpectedTimeOfArrivalDetails(cmd *cobra.Command, details expectedTimeOfArrivalDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExpectedAt != "" {
		fmt.Fprintf(out, "Expected At: %s\n", details.ExpectedAt)
	}
	fmt.Fprintf(out, "Unsure: %s\n", formatBool(details.Unsure))
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.TenderJobScheduleShift != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShift)
	}
	if details.JobScheduleShift != "" {
		fmt.Fprintf(out, "Job Schedule Shift: %s\n", details.JobScheduleShift)
	}
	if details.CreatedBy != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedBy)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
