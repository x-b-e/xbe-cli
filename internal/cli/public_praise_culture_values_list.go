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

type publicPraiseCultureValuesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	PublicPraise string
	CultureValue string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type publicPraiseCultureValueRow struct {
	ID             string `json:"id"`
	PublicPraiseID string `json:"public_praise_id,omitempty"`
	CultureValueID string `json:"culture_value_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

func newPublicPraiseCultureValuesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List public praise culture values",
		Long: `List public praise culture values.

Output Columns:
  ID             Public praise culture value identifier
  PUBLIC PRAISE  Public praise ID
  CULTURE VALUE  Culture value ID
  CREATED AT     When the link was created

Filters:
  --public-praise    Filter by public praise ID
  --culture-value    Filter by culture value ID
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List public praise culture values
  xbe view public-praise-culture-values list

  # Filter by public praise
  xbe view public-praise-culture-values list --public-praise 123

  # Filter by culture value
  xbe view public-praise-culture-values list --culture-value 456

  # Output as JSON
  xbe view public-praise-culture-values list --json`,
		Args: cobra.NoArgs,
		RunE: runPublicPraiseCultureValuesList,
	}
	initPublicPraiseCultureValuesListFlags(cmd)
	return cmd
}

func init() {
	publicPraiseCultureValuesCmd.AddCommand(newPublicPraiseCultureValuesListCmd())
}

func initPublicPraiseCultureValuesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("public-praise", "", "Filter by public praise ID")
	cmd.Flags().String("culture-value", "", "Filter by culture value ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPublicPraiseCultureValuesList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePublicPraiseCultureValuesListOptions(cmd)
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
	query.Set("fields[public-praise-culture-values]", "created-at,updated-at,public-praise,culture-value")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[public-praise]", opts.PublicPraise)
	setFilterIfPresent(query, "filter[culture-value]", opts.CultureValue)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/public-praise-culture-values", query)
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

	rows := buildPublicPraiseCultureValueRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPublicPraiseCultureValuesTable(cmd, rows)
}

func parsePublicPraiseCultureValuesListOptions(cmd *cobra.Command) (publicPraiseCultureValuesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	publicPraise, _ := cmd.Flags().GetString("public-praise")
	cultureValue, _ := cmd.Flags().GetString("culture-value")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return publicPraiseCultureValuesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		PublicPraise: publicPraise,
		CultureValue: cultureValue,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildPublicPraiseCultureValueRows(resp jsonAPIResponse) []publicPraiseCultureValueRow {
	rows := make([]publicPraiseCultureValueRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPublicPraiseCultureValueRow(resource))
	}
	return rows
}

func buildPublicPraiseCultureValueRow(resource jsonAPIResource) publicPraiseCultureValueRow {
	attrs := resource.Attributes
	return publicPraiseCultureValueRow{
		ID:             resource.ID,
		PublicPraiseID: relationshipIDFromMap(resource.Relationships, "public-praise"),
		CultureValueID: relationshipIDFromMap(resource.Relationships, "culture-value"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}
}

func buildPublicPraiseCultureValueRowFromSingle(resp jsonAPISingleResponse) publicPraiseCultureValueRow {
	return buildPublicPraiseCultureValueRow(resp.Data)
}

func renderPublicPraiseCultureValuesTable(cmd *cobra.Command, rows []publicPraiseCultureValueRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No public praise culture values found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPUBLIC PRAISE\tCULTURE VALUE\tCREATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.PublicPraiseID,
			row.CultureValueID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
