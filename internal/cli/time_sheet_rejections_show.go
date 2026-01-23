package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeSheetRejectionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetRejectionDetails struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetRejectionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet rejection details",
		Long: `Show full details of a time sheet rejection.

Output Fields:
  ID          Rejection identifier
  Time Sheet  Time sheet ID
  Comment     Comment (if provided)

Arguments:
  <id>    Time sheet rejection ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet rejection
  xbe view time-sheet-rejections show 123

  # JSON output
  xbe view time-sheet-rejections show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetRejectionsShow,
	}
	initTimeSheetRejectionsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetRejectionsCmd.AddCommand(newTimeSheetRejectionsShowCmd())
}

func initTimeSheetRejectionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetRejectionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetRejectionsShowOptions(cmd)
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
		return fmt.Errorf("time sheet rejection id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-rejections]", "time-sheet,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/time-sheet-rejections/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTimeSheetRejectionsShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildTimeSheetRejectionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetRejectionDetails(cmd, details)
}

func renderTimeSheetRejectionsShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), timeSheetRejectionDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Time sheet rejections are write-only; show is not available.")
	return nil
}

func parseTimeSheetRejectionsShowOptions(cmd *cobra.Command) (timeSheetRejectionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetRejectionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetRejectionDetails(resp jsonAPISingleResponse) timeSheetRejectionDetails {
	attrs := resp.Data.Attributes
	details := timeSheetRejectionDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}

	return details
}

func renderTimeSheetRejectionDetails(cmd *cobra.Command, details timeSheetRejectionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet: %s\n", details.TimeSheetID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
