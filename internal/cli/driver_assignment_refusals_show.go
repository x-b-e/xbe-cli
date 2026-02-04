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

type driverAssignmentRefusalsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverAssignmentRefusalDetails struct {
	ID                       string `json:"id"`
	Comment                  string `json:"comment,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	DriverID                 string `json:"driver_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
}

func newDriverAssignmentRefusalsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver assignment refusal details",
		Long: `Show the full details of a driver assignment refusal.

Output Fields:
  ID
  Tender Job Schedule Shift ID
  Driver ID
  Created By (user ID)
  Comment
  Created At
  Updated At

Arguments:
  <id>    The refusal ID (required). You can find IDs using the list command.`,
		Example: `  # Show a refusal
  xbe view driver-assignment-refusals show 123

  # Get JSON output
  xbe view driver-assignment-refusals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverAssignmentRefusalsShow,
	}
	initDriverAssignmentRefusalsShowFlags(cmd)
	return cmd
}

func init() {
	driverAssignmentRefusalsCmd.AddCommand(newDriverAssignmentRefusalsShowCmd())
}

func initDriverAssignmentRefusalsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverAssignmentRefusalsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDriverAssignmentRefusalsShowOptions(cmd)
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
		return fmt.Errorf("driver assignment refusal id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-assignment-refusals/"+id, nil)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildDriverAssignmentRefusalDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverAssignmentRefusalDetails(cmd, details)
}

func parseDriverAssignmentRefusalsShowOptions(cmd *cobra.Command) (driverAssignmentRefusalsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverAssignmentRefusalsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverAssignmentRefusalDetails(resp jsonAPISingleResponse) driverAssignmentRefusalDetails {
	attrs := resp.Data.Attributes
	details := driverAssignmentRefusalDetails{
		ID:        resp.Data.ID,
		Comment:   stringAttr(attrs, "comment"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderDriverAssignmentRefusalDetails(cmd *cobra.Command, details driverAssignmentRefusalDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
