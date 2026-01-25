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

type uiToursListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Abbreviation string
}

type uiTourRow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Description  string `json:"description,omitempty"`
}

func newUiToursListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UI tours",
		Long: `List UI tours with filtering and pagination.

Output Columns:
  ID           UI tour identifier
  NAME         UI tour name
  ABBREVIATION UI tour abbreviation
  DESCRIPTION  UI tour description (truncated)

Filters:
  --abbreviation  Filter by abbreviation

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List UI tours
  xbe view ui-tours list

  # Filter by abbreviation
  xbe view ui-tours list --abbreviation "driver-onboarding"

  # Paginate results
  xbe view ui-tours list --limit 25 --offset 50

  # Output as JSON
  xbe view ui-tours list --json`,
		RunE: runUiToursList,
	}
	initUiToursListFlags(cmd)
	return cmd
}

func init() {
	uiToursCmd.AddCommand(newUiToursListCmd())
}

func initUiToursListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUiToursList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUiToursListOptions(cmd)
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
	query.Set("fields[ui-tours]", "name,abbreviation,description")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "name")
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[abbreviation]", opts.Abbreviation)

	body, _, err := client.Get(cmd.Context(), "/v1/ui-tours", query)
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

	rows := buildUiTourRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUiToursTable(cmd, rows)
}

func parseUiToursListOptions(cmd *cobra.Command) (uiToursListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return uiToursListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Abbreviation: abbreviation,
	}, nil
}

func buildUiTourRows(resp jsonAPIResponse) []uiTourRow {
	rows := make([]uiTourRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := uiTourRow{
			ID:           resource.ID,
			Name:         stringAttr(resource.Attributes, "name"),
			Abbreviation: stringAttr(resource.Attributes, "abbreviation"),
			Description:  stringAttr(resource.Attributes, "description"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderUiToursTable(cmd *cobra.Command, rows []uiTourRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No UI tours found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tDESCRIPTION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.Abbreviation, 18),
			truncateString(row.Description, 40),
		)
	}
	return writer.Flush()
}
