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

type materialTransactionTicketGeneratorsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Organization     string
	OrganizationType string
	OrganizationID   string
	Broker           string
}

func newMaterialTransactionTicketGeneratorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material transaction ticket generators",
		Long: `List material transaction ticket generators.

Output Columns:
  ID         Ticket generator identifier
  RULE       Ticket number format rule
  ORG TYPE   Organization type
  ORG ID     Organization ID
  BROKER ID  Broker ID

Filters:
  --organization           Filter by organization (e.g., Broker|123)
  --organization-type      Filter by organization type (Broker, MaterialSupplier)
  --organization-id        Filter by organization ID (requires --organization-type)
  --broker                 Filter by broker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List ticket generators
  xbe view material-transaction-ticket-generators list

  # Filter by broker
  xbe view material-transaction-ticket-generators list --broker 123

  # Filter by organization
  xbe view material-transaction-ticket-generators list --organization "Broker|123"

  # Output as JSON
  xbe view material-transaction-ticket-generators list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTransactionTicketGeneratorsList,
	}
	initMaterialTransactionTicketGeneratorsListFlags(cmd)
	return cmd
}

func init() {
	materialTransactionTicketGeneratorsCmd.AddCommand(newMaterialTransactionTicketGeneratorsListCmd())
}

func initMaterialTransactionTicketGeneratorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (Broker, MaterialSupplier)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTransactionTicketGeneratorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTransactionTicketGeneratorsListOptions(cmd)
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
	query.Set("fields[material-transaction-ticket-generators]", "format-rule,organization,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.OrganizationID != "" && opts.OrganizationType == "" && opts.Organization == "" {
		return fmt.Errorf("--organization-type is required when using --organization-id")
	}

	if opts.Organization != "" {
		query.Set("filter[organization]", opts.Organization)
	} else if combined := normalizeOrganizationFilter(opts.OrganizationType, opts.OrganizationID); combined != "" {
		query.Set("filter[organization]", combined)
	}

	setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/material-transaction-ticket-generators", query)
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

	rows := buildMaterialTransactionTicketGeneratorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTransactionTicketGeneratorsTable(cmd, rows)
}

func parseMaterialTransactionTicketGeneratorsListOptions(cmd *cobra.Command) (materialTransactionTicketGeneratorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTransactionTicketGeneratorsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Organization:     organization,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		Broker:           broker,
	}, nil
}

func renderMaterialTransactionTicketGeneratorsTable(cmd *cobra.Command, rows []materialTransactionTicketGeneratorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction ticket generators found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRULE\tORG TYPE\tORG ID\tBROKER ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", row.ID, row.FormatRule, row.OrganizationType, row.OrganizationID, row.BrokerID)
	}
	return writer.Flush()
}
