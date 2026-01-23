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

type developersListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Broker    string
	Name      string
	ExactName string
}

type developerRow struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	WeigherSealLabel        string `json:"weigher_seal_label,omitempty"`
	IsPrevailingWage        bool   `json:"is_prevailing_wage"`
	IsCertificationRequired bool   `json:"is_certification_required"`
}

func newDevelopersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developers",
		Long: `List developers with filtering and pagination.

Developers are companies that develop projects.

Output Columns:
  ID                      Developer identifier
  NAME                    Developer name
  WEIGHER SEAL            Weigher seal label
  PREVAILING WAGE         Whether prevailing wage applies
  CERTIFICATION REQUIRED  Whether certification is required`,
		Example: `  # List all developers
  xbe view developers list

  # Search by name (fuzzy match)
  xbe view developers list --name "Acme"

  # Search by exact name
  xbe view developers list --exact-name "Acme Construction"

  # Filter by broker
  xbe view developers list --broker 123

  # Get results as JSON
  xbe view developers list --json`,
		RunE: runDevelopersList,
	}
	initDevelopersListFlags(cmd)
	return cmd
}

func init() {
	developersCmd.AddCommand(newDevelopersListCmd())
}

func initDevelopersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (fuzzy match)")
	cmd.Flags().String("exact-name", "", "Filter by exact name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDevelopersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDevelopersListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[exact_name]", opts.ExactName)

	body, _, err := client.Get(cmd.Context(), "/v1/developers", query)
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

	rows := buildDeveloperRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDevelopersTable(cmd, rows)
}

func parseDevelopersListOptions(cmd *cobra.Command) (developersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return developersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return developersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return developersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return developersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return developersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return developersListOptions{}, err
	}
	exactName, err := cmd.Flags().GetString("exact-name")
	if err != nil {
		return developersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return developersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return developersListOptions{}, err
	}

	return developersListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Broker:    broker,
		Name:      name,
		ExactName: exactName,
	}, nil
}

func buildDeveloperRows(resp jsonAPIResponse) []developerRow {
	rows := make([]developerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerRow{
			ID:                      resource.ID,
			Name:                    stringAttr(resource.Attributes, "name"),
			WeigherSealLabel:        stringAttr(resource.Attributes, "weigher-seal-label"),
			IsPrevailingWage:        boolAttr(resource.Attributes, "is-prevailing-wage"),
			IsCertificationRequired: boolAttr(resource.Attributes, "is-certification-required"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderDevelopersTable(cmd *cobra.Command, rows []developerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tWEIGHER SEAL\tPREVAILING WAGE\tCERT REQUIRED")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.WeigherSealLabel, 15),
			formatYesNo(row.IsPrevailingWage),
			formatYesNo(row.IsCertificationRequired),
		)
	}
	return writer.Flush()
}

func formatYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
