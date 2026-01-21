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

type cultureValuesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Organization string
}

type cultureValueRow struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	SequencePosition int    `json:"sequence_position"`
}

func newCultureValuesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List culture values",
		Long: `List culture values with filtering and pagination.

Culture values define organizational values used for public praise and
recognition. They help reinforce company culture.

Output Columns:
  ID           Culture value identifier
  NAME         Value name
  DESCRIPTION  Value description
  POSITION     Display order position

Filters:
  --organization  Filter by organization (format: Type|ID, e.g., Broker|123)`,
		Example: `  # List all culture values
  xbe view culture-values list

  # Filter by broker organization
  xbe view culture-values list --organization "Broker|123"

  # Output as JSON
  xbe view culture-values list --json`,
		RunE: runCultureValuesList,
	}
	initCultureValuesListFlags(cmd)
	return cmd
}

func init() {
	cultureValuesCmd.AddCommand(newCultureValuesListCmd())
}

func initCultureValuesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("organization", "", "Filter by organization ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCultureValuesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCultureValuesListOptions(cmd)
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
	query.Set("fields[culture-values]", "name,description,sequence-position")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[organization]", opts.Organization)

	body, _, err := client.Get(cmd.Context(), "/v1/culture-values", query)
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

	rows := buildCultureValueRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCultureValuesTable(cmd, rows)
}

func parseCultureValuesListOptions(cmd *cobra.Command) (cultureValuesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	organization, _ := cmd.Flags().GetString("organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return cultureValuesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Organization: organization,
	}, nil
}

func buildCultureValueRows(resp jsonAPIResponse) []cultureValueRow {
	rows := make([]cultureValueRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := cultureValueRow{
			ID:               resource.ID,
			Name:             stringAttr(resource.Attributes, "name"),
			Description:      stringAttr(resource.Attributes, "description"),
			SequencePosition: intAttr(resource.Attributes, "sequence-position"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderCultureValuesTable(cmd *cobra.Command, rows []cultureValueRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No culture values found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tPOSITION")
	for _, row := range rows {
		position := ""
		if row.SequencePosition > 0 {
			position = strconv.Itoa(row.SequencePosition)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 40),
			position,
		)
	}
	return writer.Flush()
}
