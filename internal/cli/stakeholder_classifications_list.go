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

type stakeholderClassificationsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Slug           string
	LeverageFactor string
}

func newStakeholderClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stakeholder classifications",
		Long: `List stakeholder classifications with filtering and pagination.

Stakeholder classifications categorize project stakeholders by their role
and influence level.

Output Columns:
  ID              Classification identifier
  TITLE           Classification title
  SLUG            URL-friendly identifier
  LEVERAGE        Leverage factor (influence level)

Filters:
  --slug            Filter by slug
  --leverage-factor Filter by leverage factor`,
		Example: `  # List all stakeholder classifications
  xbe view stakeholder-classifications list

  # Filter by slug
  xbe view stakeholder-classifications list --slug "owner"

  # Filter by leverage factor
  xbe view stakeholder-classifications list --leverage-factor 5

  # Output as JSON
  xbe view stakeholder-classifications list --json`,
		RunE: runStakeholderClassificationsList,
	}
	initStakeholderClassificationsListFlags(cmd)
	return cmd
}

func init() {
	stakeholderClassificationsCmd.AddCommand(newStakeholderClassificationsListCmd())
}

func initStakeholderClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("leverage-factor", "", "Filter by leverage factor")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runStakeholderClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseStakeholderClassificationsListOptions(cmd)
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
	query.Set("sort", "title")
	query.Set("fields[stakeholder-classifications]", "title,slug,leverage-factor,objectives-narrative-explicit,objectives-narrative")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[slug]", opts.Slug)
	setFilterIfPresent(query, "filter[leverage-factor]", opts.LeverageFactor)

	body, _, err := client.Get(cmd.Context(), "/v1/stakeholder-classifications", query)
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

	rows := buildStakeholderClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderStakeholderClassificationsTable(cmd, rows)
}

func parseStakeholderClassificationsListOptions(cmd *cobra.Command) (stakeholderClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	slug, _ := cmd.Flags().GetString("slug")
	leverageFactor, _ := cmd.Flags().GetString("leverage-factor")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return stakeholderClassificationsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Slug:           slug,
		LeverageFactor: leverageFactor,
	}, nil
}

type stakeholderClassificationRow struct {
	ID                          string `json:"id"`
	Title                       string `json:"title"`
	Slug                        string `json:"slug"`
	LeverageFactor              *int   `json:"leverage_factor,omitempty"`
	ObjectivesNarrativeExplicit string `json:"objectives_narrative_explicit,omitempty"`
	ObjectivesNarrative         string `json:"objectives_narrative,omitempty"`
}

func buildStakeholderClassificationRows(resp jsonAPIResponse) []stakeholderClassificationRow {
	rows := make([]stakeholderClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := stakeholderClassificationRow{
			ID:                          resource.ID,
			Title:                       stringAttr(resource.Attributes, "title"),
			Slug:                        stringAttr(resource.Attributes, "slug"),
			ObjectivesNarrativeExplicit: stringAttr(resource.Attributes, "objectives-narrative-explicit"),
			ObjectivesNarrative:         stringAttr(resource.Attributes, "objectives-narrative"),
		}

		if lf, ok := resource.Attributes["leverage-factor"]; ok && lf != nil {
			if lfFloat, ok := lf.(float64); ok {
				lfInt := int(lfFloat)
				row.LeverageFactor = &lfInt
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderStakeholderClassificationsTable(cmd *cobra.Command, rows []stakeholderClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No stakeholder classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTITLE\tSLUG\tLEVERAGE")
	for _, row := range rows {
		leverage := ""
		if row.LeverageFactor != nil {
			leverage = strconv.Itoa(*row.LeverageFactor)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Title, 30),
			truncateString(row.Slug, 25),
			leverage,
		)
	}
	return writer.Flush()
}
