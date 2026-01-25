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

type tractorTrailerCredentialClassificationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Name         string
	Organization string
}

func newTractorTrailerCredentialClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tractor/trailer credential classifications",
		Long: `List tractor and trailer credential classifications with filtering and pagination.

These classifications define types of credentials that can be assigned to tractors and trailers.

Output Columns:
  ID           Classification identifier
  NAME         Classification name
  DESCRIPTION  Description
  ISSUER       Issuer name
  EXTERNAL ID  External identifier
  ORG TYPE     Organization type
  ORG ID       Organization ID

Filters:
  --name          Filter by name
  --organization  Filter by organization (format: Type|ID, e.g., Trucker|123)`,
		Example: `  # List all tractor/trailer credential classifications
  xbe view tractor-trailer-credential-classifications list

  # Filter by name
  xbe view tractor-trailer-credential-classifications list --name "Insurance"

  # Filter by organization
  xbe view tractor-trailer-credential-classifications list --organization "Trucker|123"

  # Output as JSON
  xbe view tractor-trailer-credential-classifications list --json`,
		RunE: runTractorTrailerCredentialClassificationsList,
	}
	initTractorTrailerCredentialClassificationsListFlags(cmd)
	return cmd
}

func init() {
	tractorTrailerCredentialClassificationsCmd.AddCommand(newTractorTrailerCredentialClassificationsListCmd())
}

func initTractorTrailerCredentialClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g., Trucker|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorTrailerCredentialClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTractorTrailerCredentialClassificationsListOptions(cmd)
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
	query.Set("fields[tractor-trailer-credential-classifications]", "name,description,issuer-name,external-id,organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-trailer-credential-classifications", query)
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

	rows := buildTractorTrailerCredentialClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTractorTrailerCredentialClassificationsTable(cmd, rows)
}

func parseTractorTrailerCredentialClassificationsListOptions(cmd *cobra.Command) (tractorTrailerCredentialClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	organization, _ := cmd.Flags().GetString("organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorTrailerCredentialClassificationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Name:         name,
		Organization: organization,
	}, nil
}

type tractorTrailerCredentialClassificationRow struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	IssuerName       string `json:"issuer_name,omitempty"`
	ExternalID       string `json:"external_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
}

func buildTractorTrailerCredentialClassificationRows(resp jsonAPIResponse) []tractorTrailerCredentialClassificationRow {
	rows := make([]tractorTrailerCredentialClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tractorTrailerCredentialClassificationRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
			IssuerName:  stringAttr(resource.Attributes, "issuer-name"),
			ExternalID:  stringAttr(resource.Attributes, "external-id"),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTractorTrailerCredentialClassificationsTable(cmd *cobra.Command, rows []tractorTrailerCredentialClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tractor/trailer credential classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tISSUER\tORG TYPE\tORG ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 25),
			truncateString(row.IssuerName, 20),
			row.OrganizationType,
			row.OrganizationID,
		)
	}
	return writer.Flush()
}
