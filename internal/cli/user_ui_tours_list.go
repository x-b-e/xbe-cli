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

type userUiToursListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	NotID        string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type userUiTourRow struct {
	ID                 string `json:"id"`
	CompletedAt        string `json:"completed_at,omitempty"`
	SkippedAt          string `json:"skipped_at,omitempty"`
	UserID             string `json:"user_id,omitempty"`
	UserName           string `json:"user_name,omitempty"`
	UserEmail          string `json:"user_email,omitempty"`
	UiTourID           string `json:"ui_tour_id,omitempty"`
	UiTourName         string `json:"ui_tour_name,omitempty"`
	UiTourAbbreviation string `json:"ui_tour_abbreviation,omitempty"`
}

func newUserUiToursListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user UI tours",
		Long: `List user UI tours with filtering and pagination.

User UI tours track completion or skip status for a specific user and UI tour.

Output Columns:
  ID           User UI tour identifier
  USER         User name/email or ID
  UI TOUR      UI tour name/abbreviation or ID
  COMPLETED AT Completion timestamp
  SKIPPED AT   Skipped timestamp

Filters:
  --not-id          Exclude user UI tour IDs (comma-separated)
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by has created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List user UI tours
  xbe view user-ui-tours list

  # Sort by created-at
  xbe view user-ui-tours list --sort created-at

  # Filter by created-at window
  xbe view user-ui-tours list --created-at-min 2025-01-01T00:00:00Z --created-at-max 2025-12-31T23:59:59Z

  # Output as JSON
  xbe view user-ui-tours list --json`,
		Args: cobra.NoArgs,
		RunE: runUserUiToursList,
	}
	initUserUiToursListFlags(cmd)
	return cmd
}

func init() {
	userUiToursCmd.AddCommand(newUserUiToursListCmd())
}

func initUserUiToursListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("not-id", "", "Exclude user UI tour IDs (comma-separated)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserUiToursList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserUiToursListOptions(cmd)
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
	query.Set("fields[user-ui-tours]", "completed-at,skipped-at,user,ui-tour")
	query.Set("include", "user,ui-tour")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[ui-tours]", "name,abbreviation")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[not-id]", opts.NotID)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/user-ui-tours", query)
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

	rows := buildUserUiTourRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserUiToursTable(cmd, rows)
}

func parseUserUiToursListOptions(cmd *cobra.Command) (userUiToursListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	notID, _ := cmd.Flags().GetString("not-id")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userUiToursListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		NotID:        notID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildUserUiTourRows(resp jsonAPIResponse) []userUiTourRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]userUiTourRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildUserUiTourRow(resource, included))
	}
	return rows
}

func userUiTourRowFromSingle(resp jsonAPISingleResponse) userUiTourRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildUserUiTourRow(resp.Data, included)
}

func buildUserUiTourRow(resource jsonAPIResource, included map[string]jsonAPIResource) userUiTourRow {
	row := userUiTourRow{
		ID:          resource.ID,
		CompletedAt: strings.TrimSpace(stringAttr(resource.Attributes, "completed-at")),
		SkippedAt:   strings.TrimSpace(stringAttr(resource.Attributes, "skipped-at")),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			row.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}

	if rel, ok := resource.Relationships["ui-tour"]; ok && rel.Data != nil {
		row.UiTourID = rel.Data.ID
		if tour, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UiTourName = strings.TrimSpace(stringAttr(tour.Attributes, "name"))
			row.UiTourAbbreviation = strings.TrimSpace(stringAttr(tour.Attributes, "abbreviation"))
		}
	}

	return row
}

func renderUserUiToursTable(cmd *cobra.Command, rows []userUiTourRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user UI tours found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tUI TOUR\tCOMPLETED AT\tSKIPPED AT")
	for _, row := range rows {
		userLabel := firstNonEmpty(row.UserName, row.UserEmail, row.UserID)
		uiTourLabel := firstNonEmpty(row.UiTourName, row.UiTourAbbreviation, row.UiTourID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(userLabel, 30),
			truncateString(uiTourLabel, 30),
			row.CompletedAt,
			row.SkippedAt,
		)
	}
	return writer.Flush()
}
