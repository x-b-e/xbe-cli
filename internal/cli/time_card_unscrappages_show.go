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

type timeCardUnscrappagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardUnscrappageDetails struct {
	ID         string `json:"id"`
	TimeCardID string `json:"time_card_id,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

func newTimeCardUnscrappagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card unscrappage details",
		Long: `Show full details of a time card unscrappage.

Output Fields:
  ID         Unscrappage identifier
  Time Card  Time card ID
  Comment    Comment (if provided)

Arguments:
  <id>    Time card unscrappage ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time card unscrappage
  xbe view time-card-unscrappages show 123

  # JSON output
  xbe view time-card-unscrappages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardUnscrappagesShow,
	}
	initTimeCardUnscrappagesShowFlags(cmd)
	return cmd
}

func init() {
	timeCardUnscrappagesCmd.AddCommand(newTimeCardUnscrappagesShowCmd())
}

func initTimeCardUnscrappagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardUnscrappagesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardUnscrappagesShowOptions(cmd)
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
		return fmt.Errorf("time card unscrappage id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-unscrappages]", "time-card,comment")

	body, status, err := client.Get(cmd.Context(), "/v1/time-card-unscrappages/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTimeCardUnscrappagesShowUnavailable(cmd, opts.JSON)
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

	details := buildTimeCardUnscrappageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardUnscrappageDetails(cmd, details)
}

func renderTimeCardUnscrappagesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), timeCardUnscrappageDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Time card unscrappages are write-only; show is not available.")
	return nil
}

func parseTimeCardUnscrappagesShowOptions(cmd *cobra.Command) (timeCardUnscrappagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardUnscrappagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardUnscrappageDetails(resp jsonAPISingleResponse) timeCardUnscrappageDetails {
	attrs := resp.Data.Attributes
	details := timeCardUnscrappageDetails{
		ID:      resp.Data.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resp.Data.Relationships["time-card"]; ok && rel.Data != nil {
		details.TimeCardID = rel.Data.ID
	}

	return details
}

func renderTimeCardUnscrappageDetails(cmd *cobra.Command, details timeCardUnscrappageDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	fmt.Fprintf(out, "Comment: %s\n", formatOptional(details.Comment))

	return nil
}
