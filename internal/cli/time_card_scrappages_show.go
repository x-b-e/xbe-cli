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

type timeCardScrappagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newTimeCardScrappagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card scrappage details",
		Long: `Show the full details of a time card scrappage.

Output Fields:
  ID
  Time Card ID
  Comment

Arguments:
  <id>    The scrappage ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a scrappage
  xbe view time-card-scrappages show 123

  # Output as JSON
  xbe view time-card-scrappages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardScrappagesShow,
	}
	initTimeCardScrappagesShowFlags(cmd)
	return cmd
}

func init() {
	timeCardScrappagesCmd.AddCommand(newTimeCardScrappagesShowCmd())
}

func initTimeCardScrappagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardScrappagesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardScrappagesShowOptions(cmd)
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
		return fmt.Errorf("time card scrappage id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-scrappages]", "comment,time-card")

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-scrappages/"+id, query)
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

	details := buildTimeCardScrappageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardScrappageDetails(cmd, details)
}

func parseTimeCardScrappagesShowOptions(cmd *cobra.Command) (timeCardScrappagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardScrappagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderTimeCardScrappageDetails(cmd *cobra.Command, details timeCardScrappageRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card ID: %s\n", details.TimeCardID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	return nil
}
