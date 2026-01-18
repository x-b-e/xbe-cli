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

type jobProductionPlansRecapOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newJobProductionPlansRecapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recap <id>",
		Short: "Show recap posts for a job production plan",
		Long: `Show recap posts for a specific job production plan.

Retrieves posts of type 'job_production_plan_recap' where the creator is
the specified job production plan. These recaps summarize the day's
production activity.

Arguments:
  <id>    The job production plan ID (required)`,
		Example: `  # View recaps for a plan
  xbe view job-production-plans recap 1461191

  # Get recap as JSON
  xbe view job-production-plans recap 1461191 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlansRecap,
	}
	initJobProductionPlansRecapFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlansCmd.AddCommand(newJobProductionPlansRecapCmd())
}

func initJobProductionPlansRecapFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlansRecap(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlansRecapOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-published-at")
	query.Set("fields[posts]", "post-type,published-at,text-content,short-text-content,creator-name,status")
	query.Set("filter[post-type]", "job_production_plan_recap")
	query.Set("filter[creator]", "JobProductionPlan|"+id)

	body, _, err := client.Get(cmd.Context(), "/v1/posts", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.JSON {
		rows := buildRecapRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRecapFeed(cmd, resp)
}

func parseJobProductionPlansRecapOptions(cmd *cobra.Command) (jobProductionPlansRecapOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlansRecapOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlansRecapOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlansRecapOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlansRecapOptions{}, err
	}

	return jobProductionPlansRecapOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

type recapRow struct {
	ID        string `json:"id"`
	Published string `json:"published"`
	Content   string `json:"content"`
}

func buildRecapRows(resp jsonAPIResponse) []recapRow {
	rows := make([]recapRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		// Prefer full text_content, fall back to short_text_content
		content := strings.TrimSpace(stringAttr(resource.Attributes, "text-content"))
		if content == "" {
			content = strings.TrimSpace(stringAttr(resource.Attributes, "short-text-content"))
		}
		rows = append(rows, recapRow{
			ID:        resource.ID,
			Published: formatDate(stringAttr(resource.Attributes, "published-at")),
			Content:   content,
		})
	}
	return rows
}

func renderRecapFeed(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildRecapRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No recaps found for this job production plan.")
		return nil
	}

	out := cmd.OutOrStdout()

	for i, row := range rows {
		// Header line
		fmt.Fprintf(out, "[%s] Recap - %s\n", row.ID, row.Published)
		fmt.Fprintln(out, strings.Repeat("-", 60))

		// Content (strip markdown for cleaner display)
		if row.Content != "" {
			content := stripMarkdown(row.Content)
			fmt.Fprintln(out, content)
		}

		// Blank line between recaps
		if i < len(rows)-1 {
			fmt.Fprintln(out)
		}
	}

	return nil
}
