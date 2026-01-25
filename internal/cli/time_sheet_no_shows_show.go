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

type timeSheetNoShowsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeSheetNoShowDetails struct {
	ID           string `json:"id"`
	TimeSheetID  string `json:"time_sheet_id,omitempty"`
	NoShowReason string `json:"no_show_reason,omitempty"`
	CreatedByID  string `json:"created_by_id,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

func newTimeSheetNoShowsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time sheet no-show details",
		Long: `Show the full details of a time sheet no-show.

Output Fields:
  ID
  Time Sheet ID
  No-Show Reason
  Created By (user ID)
  Created At
  Updated At

Arguments:
  <id>    The no-show ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time sheet no-show
  xbe view time-sheet-no-shows show 123

  # Output as JSON
  xbe view time-sheet-no-shows show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeSheetNoShowsShow,
	}
	initTimeSheetNoShowsShowFlags(cmd)
	return cmd
}

func init() {
	timeSheetNoShowsCmd.AddCommand(newTimeSheetNoShowsShowCmd())
}

func initTimeSheetNoShowsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetNoShowsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeSheetNoShowsShowOptions(cmd)
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
		return fmt.Errorf("time sheet no-show id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-sheet-no-shows]", "no-show-reason,time-sheet,created-by,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-no-shows/"+id, query)
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

	details := buildTimeSheetNoShowDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeSheetNoShowDetails(cmd, details)
}

func parseTimeSheetNoShowsShowOptions(cmd *cobra.Command) (timeSheetNoShowsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetNoShowsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeSheetNoShowDetails(resp jsonAPISingleResponse) timeSheetNoShowDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := timeSheetNoShowDetails{
		ID:           resource.ID,
		NoShowReason: stringAttr(attrs, "no-show-reason"),
		CreatedAt:    formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:    formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		details.TimeSheetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderTimeSheetNoShowDetails(cmd *cobra.Command, details timeSheetNoShowDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet ID: %s\n", details.TimeSheetID)
	}
	if details.NoShowReason != "" {
		fmt.Fprintf(out, "No-Show Reason: %s\n", details.NoShowReason)
	}
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
