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

type timeSheetUnsubmissionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newTimeSheetUnsubmissionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet unsubmission details",
		Long: `Show the full details of a time sheet unsubmission.

Output Fields:
  ID         Unsubmission identifier
  Time Sheet Time sheet ID
  Comment    Status change comment

Arguments:
  <id>    The time sheet unsubmission ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet unsubmission
  xbe view time-sheet-unsubmissions show 123

  # Output as JSON
  xbe view time-sheet-unsubmissions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetUnsubmissionsShow,
	}
	initTimeSheetUnsubmissionsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetUnsubmissionsCmd.AddCommand(newTimeSheetUnsubmissionsShowCmd())
}

func initTimeSheetUnsubmissionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetUnsubmissionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetUnsubmissionsShowOptions(cmd)
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
		return fmt.Errorf("time sheet unsubmission id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-unsubmissions]", "comment,time-sheet")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-unsubmissions/"+id, query)
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

	details := buildTimeSheetUnsubmissionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetUnsubmissionDetails(cmd, details)
}

func parseTimeSheetUnsubmissionsShowOptions(cmd *cobra.Command) (timeSheetUnsubmissionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetUnsubmissionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderTimeSheetUnsubmissionDetails(cmd *cobra.Command, details timeSheetUnsubmissionRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet: %s\n", details.TimeSheetID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
