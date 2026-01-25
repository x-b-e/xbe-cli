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

type lineupSummaryRequestsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupSummaryRequestDetails struct {
	ID             string   `json:"id"`
	LevelType      string   `json:"level_type,omitempty"`
	LevelID        string   `json:"level_id,omitempty"`
	StartAtMin     string   `json:"start_at_min,omitempty"`
	StartAtMax     string   `json:"start_at_max,omitempty"`
	EmailTo        []string `json:"email_to,omitempty"`
	SendIfNoShifts bool     `json:"send_if_no_shifts"`
	Note           string   `json:"note,omitempty"`
	CreatedByID    string   `json:"created_by_id,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
}

func newLineupSummaryRequestsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup summary request details",
		Long: `Show the full details of a lineup summary request.

Output Fields:
  ID
  Level Type
  Level ID
  Start At Min
  Start At Max
  Email To
  Send If No Shifts
  Note
  Created By ID
  Created At
  Updated At

Arguments:
  <id>    The lineup summary request ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a lineup summary request
  xbe view lineup-summary-requests show 123

  # Output as JSON
  xbe view lineup-summary-requests show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupSummaryRequestsShow,
	}
	initLineupSummaryRequestsShowFlags(cmd)
	return cmd
}

func init() {
	lineupSummaryRequestsCmd.AddCommand(newLineupSummaryRequestsShowCmd())
}

func initLineupSummaryRequestsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupSummaryRequestsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupSummaryRequestsShowOptions(cmd)
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
		return fmt.Errorf("lineup summary request id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-summary-requests]", "start-at-min,start-at-max,email-to,send-if-no-shifts,note,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-summary-requests/"+id, query)
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

	details := buildLineupSummaryRequestDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupSummaryRequestDetails(cmd, details)
}

func parseLineupSummaryRequestsShowOptions(cmd *cobra.Command) (lineupSummaryRequestsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupSummaryRequestsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupSummaryRequestDetails(resp jsonAPISingleResponse) lineupSummaryRequestDetails {
	attrs := resp.Data.Attributes
	details := lineupSummaryRequestDetails{
		ID:             resp.Data.ID,
		StartAtMin:     formatDateTime(stringAttr(attrs, "start-at-min")),
		StartAtMax:     formatDateTime(stringAttr(attrs, "start-at-max")),
		EmailTo:        stringSliceAttr(attrs, "email-to"),
		SendIfNoShifts: boolAttr(attrs, "send-if-no-shifts"),
		Note:           stringAttr(attrs, "note"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["level"]; ok && rel.Data != nil {
		details.LevelType = rel.Data.Type
		details.LevelID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderLineupSummaryRequestDetails(cmd *cobra.Command, details lineupSummaryRequestDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LevelType != "" {
		fmt.Fprintf(out, "Level Type: %s\n", details.LevelType)
	}
	if details.LevelID != "" {
		fmt.Fprintf(out, "Level ID: %s\n", details.LevelID)
	}
	if details.StartAtMin != "" {
		fmt.Fprintf(out, "Start At Min: %s\n", details.StartAtMin)
	}
	if details.StartAtMax != "" {
		fmt.Fprintf(out, "Start At Max: %s\n", details.StartAtMax)
	}
	if len(details.EmailTo) > 0 {
		fmt.Fprintf(out, "Email To: %s\n", strings.Join(details.EmailTo, ", "))
	}
	fmt.Fprintf(out, "Send If No Shifts: %t\n", details.SendIfNoShifts)
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
