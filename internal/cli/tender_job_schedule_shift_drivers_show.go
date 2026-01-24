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

type tenderJobScheduleShiftDriversShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderJobScheduleShiftDriverDetails struct {
	ID                       string `json:"id"`
	IsPrimary                bool   `json:"is_primary"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	UserID                   string `json:"user_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
}

func newTenderJobScheduleShiftDriversShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender job schedule shift driver details",
		Long: `Show the full details of a tender job schedule shift driver.

Output Fields:
  ID
  Tender Job Schedule Shift ID
  User ID
  Is Primary
  Created By (user ID)
  Created At
  Updated At

Arguments:
  <id>    The shift driver ID (required). You can find IDs using the list command.`,
		Example: `  # Show shift driver details
  xbe view tender-job-schedule-shift-drivers show 123

  # Get JSON output
  xbe view tender-job-schedule-shift-drivers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderJobScheduleShiftDriversShow,
	}
	initTenderJobScheduleShiftDriversShowFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftDriversCmd.AddCommand(newTenderJobScheduleShiftDriversShowCmd())
}

func initTenderJobScheduleShiftDriversShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftDriversShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTenderJobScheduleShiftDriversShowOptions(cmd)
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
		return fmt.Errorf("tender job schedule shift driver id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-drivers/"+id, nil)
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

	details := buildTenderJobScheduleShiftDriverDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderJobScheduleShiftDriverDetails(cmd, details)
}

func parseTenderJobScheduleShiftDriversShowOptions(cmd *cobra.Command) (tenderJobScheduleShiftDriversShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftDriversShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderJobScheduleShiftDriverDetails(resp jsonAPISingleResponse) tenderJobScheduleShiftDriverDetails {
	attrs := resp.Data.Attributes
	details := tenderJobScheduleShiftDriverDetails{
		ID:        resp.Data.ID,
		IsPrimary: boolAttr(attrs, "is-primary"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderTenderJobScheduleShiftDriverDetails(cmd *cobra.Command, details tenderJobScheduleShiftDriverDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	fmt.Fprintf(out, "Is Primary: %t\n", details.IsPrimary)
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
