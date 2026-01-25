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

type developerTruckerCertificationClassificationsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Developer string
	Broker    string
}

func newDeveloperTruckerCertificationClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer trucker certification classifications",
		Long: `List developer trucker certification classifications with filtering and pagination.

These classifications define types of certifications that truckers can have for a developer.

Output Columns:
  ID          Classification identifier
  NAME        Classification name
  DEVELOPER   Developer ID

Filters:
  --developer  Filter by developer ID
  --broker     Filter by broker ID`,
		Example: `  # List all developer trucker certification classifications
  xbe view developer-trucker-certification-classifications list

  # Filter by developer
  xbe view developer-trucker-certification-classifications list --developer 123

  # Filter by broker
  xbe view developer-trucker-certification-classifications list --broker 456

  # Output as JSON
  xbe view developer-trucker-certification-classifications list --json`,
		RunE: runDeveloperTruckerCertificationClassificationsList,
	}
	initDeveloperTruckerCertificationClassificationsListFlags(cmd)
	return cmd
}

func init() {
	developerTruckerCertificationClassificationsCmd.AddCommand(newDeveloperTruckerCertificationClassificationsListCmd())
}

func initDeveloperTruckerCertificationClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperTruckerCertificationClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperTruckerCertificationClassificationsListOptions(cmd)
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
	query.Set("fields[developer-trucker-certification-classifications]", "name,developer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/developer-trucker-certification-classifications", query)
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

	rows := buildDeveloperTruckerCertificationClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperTruckerCertificationClassificationsTable(cmd, rows)
}

func parseDeveloperTruckerCertificationClassificationsListOptions(cmd *cobra.Command) (developerTruckerCertificationClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	developer, _ := cmd.Flags().GetString("developer")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerTruckerCertificationClassificationsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Developer: developer,
		Broker:    broker,
	}, nil
}

type developerTruckerCertificationClassificationRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DeveloperID string `json:"developer_id,omitempty"`
}

func buildDeveloperTruckerCertificationClassificationRows(resp jsonAPIResponse) []developerTruckerCertificationClassificationRow {
	rows := make([]developerTruckerCertificationClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerTruckerCertificationClassificationRow{
			ID:   resource.ID,
			Name: stringAttr(resource.Attributes, "name"),
		}

		if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
			row.DeveloperID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDeveloperTruckerCertificationClassificationsTable(cmd *cobra.Command, rows []developerTruckerCertificationClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer trucker certification classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDEVELOPER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 40),
			row.DeveloperID,
		)
	}
	return writer.Flush()
}
