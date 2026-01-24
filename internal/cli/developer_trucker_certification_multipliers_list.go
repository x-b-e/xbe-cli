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

type developerTruckerCertificationMultipliersListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	DeveloperTruckerCertification string
	Trailer                       string
}

type developerTruckerCertificationMultiplierRow struct {
	ID                              string  `json:"id"`
	DeveloperTruckerCertificationID string  `json:"developer_trucker_certification_id,omitempty"`
	TrailerID                       string  `json:"trailer_id,omitempty"`
	Multiplier                      float64 `json:"multiplier,omitempty"`
}

func newDeveloperTruckerCertificationMultipliersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer trucker certification multipliers",
		Long: `List developer trucker certification multipliers with filtering and pagination.

Output Columns:
  ID              Multiplier identifier
  CERTIFICATION   Developer trucker certification ID
  TRAILER         Trailer ID
  MULTIPLIER      Trailer multiplier (0-1)

Filters:
  --developer-trucker-certification  Filter by developer trucker certification ID
  --trailer                          Filter by trailer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List multipliers
  xbe view developer-trucker-certification-multipliers list

  # Filter by developer trucker certification
  xbe view developer-trucker-certification-multipliers list --developer-trucker-certification 123

  # Filter by trailer
  xbe view developer-trucker-certification-multipliers list --trailer 456

  # JSON output
  xbe view developer-trucker-certification-multipliers list --json`,
		Args: cobra.NoArgs,
		RunE: runDeveloperTruckerCertificationMultipliersList,
	}
	initDeveloperTruckerCertificationMultipliersListFlags(cmd)
	return cmd
}

func init() {
	developerTruckerCertificationMultipliersCmd.AddCommand(newDeveloperTruckerCertificationMultipliersListCmd())
}

func initDeveloperTruckerCertificationMultipliersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("developer-trucker-certification", "", "Filter by developer trucker certification ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperTruckerCertificationMultipliersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperTruckerCertificationMultipliersListOptions(cmd)
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
	query.Set("fields[developer-trucker-certification-multipliers]", "multiplier,developer-trucker-certification,trailer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[developer_trucker_certification]", opts.DeveloperTruckerCertification)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)

	body, _, err := client.Get(cmd.Context(), "/v1/developer-trucker-certification-multipliers", query)
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

	rows := buildDeveloperTruckerCertificationMultiplierRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperTruckerCertificationMultipliersTable(cmd, rows)
}

func parseDeveloperTruckerCertificationMultipliersListOptions(cmd *cobra.Command) (developerTruckerCertificationMultipliersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	developerTruckerCertification, _ := cmd.Flags().GetString("developer-trucker-certification")
	trailer, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerTruckerCertificationMultipliersListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		DeveloperTruckerCertification: developerTruckerCertification,
		Trailer:                       trailer,
	}, nil
}

func buildDeveloperTruckerCertificationMultiplierRows(resp jsonAPIResponse) []developerTruckerCertificationMultiplierRow {
	rows := make([]developerTruckerCertificationMultiplierRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := developerTruckerCertificationMultiplierRow{
			ID: resource.ID,
		}

		if multiplier, ok := floatAttrValue(resource.Attributes, "multiplier"); ok {
			row.Multiplier = multiplier
		}

		if rel, ok := resource.Relationships["developer-trucker-certification"]; ok && rel.Data != nil {
			row.DeveloperTruckerCertificationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}

		rows = append(rows, row)
	}

	return rows
}

func renderDeveloperTruckerCertificationMultipliersTable(cmd *cobra.Command, rows []developerTruckerCertificationMultiplierRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer trucker certification multipliers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCERTIFICATION\tTRAILER\tMULTIPLIER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.DeveloperTruckerCertificationID, 36),
			truncateString(row.TrailerID, 36),
			formatMultiplier(row.Multiplier),
		)
	}
	return writer.Flush()
}

func formatMultiplier(value float64) string {
	text := strconv.FormatFloat(value, 'f', 6, 64)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" {
		return "0"
	}
	return text
}
