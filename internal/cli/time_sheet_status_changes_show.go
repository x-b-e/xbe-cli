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

type timeSheetStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetStatusChangeDetails struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet status change details",
		Long: `Show the full details of a time sheet status change.

Output Fields:
  ID
  Time Sheet ID
  Status
  Changed At
  Changed By (user ID)
  Comment

Arguments:
  <id>    The status change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet status change
  xbe view time-sheet-status-changes show 123

  # Output as JSON
  xbe view time-sheet-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetStatusChangesShow,
	}
	initTimeSheetStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetStatusChangesCmd.AddCommand(newTimeSheetStatusChangesShowCmd())
}

func initTimeSheetStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetStatusChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetStatusChangesShowOptions(cmd)
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
		return fmt.Errorf("time sheet status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-status-changes]", "status,changed-at,comment,time-sheet,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-status-changes/"+id, query)
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

	details := buildTimeSheetStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetStatusChangeDetails(cmd, details)
}

func parseTimeSheetStatusChangesShowOptions(cmd *cobra.Command) (timeSheetStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetStatusChangeDetails(resp jsonAPISingleResponse) timeSheetStatusChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := timeSheetStatusChangeDetails{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
	}

	return details
}

func renderTimeSheetStatusChangeDetails(cmd *cobra.Command, details timeSheetStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet ID: %s\n", details.TimeSheetID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
