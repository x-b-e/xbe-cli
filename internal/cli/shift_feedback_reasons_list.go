package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type shiftFeedbackReasonsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Name                string
	Kind                string
	Slug                string
	DefaultRating       string
	HasBot              string
	HasCorrectiveAction string
}

func newShiftFeedbackReasonsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shift feedback reasons",
		Long: `List shift feedback reasons.

Shift feedback reasons define the types of feedback that can be given for shifts.

Output Columns:
  ID               Reason identifier
  NAME             Reason name
  KIND             Feedback kind (positive, negative, etc.)
  DEFAULT RATING   Default rating value
  SLUG             URL-friendly identifier

Filters:
  --name            Filter by name
  --kind            Filter by kind
  --slug            Filter by slug
  --default-rating  Filter by default rating
  --has-bot         Filter by bot presence (true/false)`,
		Example: `  # List all shift feedback reasons
  xbe view shift-feedback-reasons list

  # Filter by kind
  xbe view shift-feedback-reasons list --kind positive

  # Filter by slug
  xbe view shift-feedback-reasons list --slug "late-arrival"

  # Output as JSON
  xbe view shift-feedback-reasons list --json`,
		RunE: runShiftFeedbackReasonsList,
	}
	initShiftFeedbackReasonsListFlags(cmd)
	return cmd
}

func init() {
	shiftFeedbackReasonsCmd.AddCommand(newShiftFeedbackReasonsListCmd())
}

func initShiftFeedbackReasonsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("kind", "", "Filter by kind")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("default-rating", "", "Filter by default rating")
	cmd.Flags().String("has-bot", "", "Filter by bot presence (true/false)")
	cmd.Flags().String("has-corrective-action", "", "Filter by having corrective action (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runShiftFeedbackReasonsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseShiftFeedbackReasonsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "name")
	query.Set("fields[shift-feedback-reasons]", "name,kind,default-rating,slug,corrective-action")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[slug]", opts.Slug)
	setFilterIfPresent(query, "filter[default-rating]", opts.DefaultRating)
	setFilterIfPresent(query, "filter[has-bot]", opts.HasBot)
	setFilterIfPresent(query, "filter[has-corrective-action]", opts.HasCorrectiveAction)

	body, _, err := client.Get(cmd.Context(), "/v1/shift-feedback-reasons", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildShiftFeedbackReasonRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderShiftFeedbackReasonsTable(cmd, rows)
}

func parseShiftFeedbackReasonsListOptions(cmd *cobra.Command) (shiftFeedbackReasonsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	kind, _ := cmd.Flags().GetString("kind")
	slug, _ := cmd.Flags().GetString("slug")
	defaultRating, _ := cmd.Flags().GetString("default-rating")
	hasBot, _ := cmd.Flags().GetString("has-bot")
	hasCorrectiveAction, _ := cmd.Flags().GetString("has-corrective-action")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return shiftFeedbackReasonsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Name:                name,
		Kind:                kind,
		Slug:                slug,
		DefaultRating:       defaultRating,
		HasBot:              hasBot,
		HasCorrectiveAction: hasCorrectiveAction,
	}, nil
}

type shiftFeedbackReasonRow struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Kind             string `json:"kind"`
	DefaultRating    any    `json:"default_rating,omitempty"`
	Slug             string `json:"slug"`
	CorrectiveAction string `json:"corrective_action,omitempty"`
}

func buildShiftFeedbackReasonRows(resp jsonAPIResponse) []shiftFeedbackReasonRow {
	rows := make([]shiftFeedbackReasonRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := shiftFeedbackReasonRow{
			ID:               resource.ID,
			Name:             stringAttr(resource.Attributes, "name"),
			Kind:             stringAttr(resource.Attributes, "kind"),
			DefaultRating:    resource.Attributes["default-rating"],
			Slug:             stringAttr(resource.Attributes, "slug"),
			CorrectiveAction: stringAttr(resource.Attributes, "corrective-action"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderShiftFeedbackReasonsTable(cmd *cobra.Command, rows []shiftFeedbackReasonRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift feedback reasons found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tKIND\tDEFAULT RATING\tSLUG")
	for _, row := range rows {
		rating := ""
		if row.DefaultRating != nil {
			rating = fmt.Sprintf("%v", row.DefaultRating)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.Kind,
			rating,
			truncateString(row.Slug, 25),
		)
	}
	return writer.Flush()
}
