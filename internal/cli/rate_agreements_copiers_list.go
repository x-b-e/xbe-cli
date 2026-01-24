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

type rateAgreementsCopiersListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Broker                string
	RateAgreementTemplate string
	CreatedBy             string
}

type rateAgreementsCopierRow struct {
	ID                    string `json:"id"`
	RateAgreementTemplate string `json:"rate_agreement_template_id,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	CreatedByID           string `json:"created_by_id,omitempty"`
	TargetType            string `json:"target_type,omitempty"`
	TargetCount           int    `json:"target_count,omitempty"`
	Note                  string `json:"note,omitempty"`
	ScheduledAt           string `json:"scheduled_at,omitempty"`
	ProcessedAt           string `json:"processed_at,omitempty"`
	CopiersResultsCount   int    `json:"results_count,omitempty"`
	CopiersErrorsCount    int    `json:"errors_count,omitempty"`
}

func newRateAgreementsCopiersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rate agreements copiers",
		Long: `List rate agreements copiers.

Rate agreements copiers copy a template rate agreement to multiple
customers or truckers.

Output Columns:
  ID            Copier identifier
  TEMPLATE      Rate agreement template ID
  BROKER        Broker ID
  CREATED BY    User ID (if present)
  TARGETS       Target type and count (customers:3 or truckers:5)
  SCHEDULED AT  Scheduled timestamp
  PROCESSED AT  Processed timestamp (if present)
  RESULTS       Copier success count (if present)
  ERRORS        Copier error count (if present)
  NOTE          Copier note (if present)

Filters:
  --broker                   Filter by broker ID
  --rate-agreement-template  Filter by template rate agreement ID
  --created-by               Filter by creator user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List rate agreements copiers
  xbe view rate-agreements-copiers list

  # Filter by broker and template
  xbe view rate-agreements-copiers list --broker 123 --rate-agreement-template 456

  # Output as JSON
  xbe view rate-agreements-copiers list --json`,
		Args: cobra.NoArgs,
		RunE: runRateAgreementsCopiersList,
	}
	initRateAgreementsCopiersListFlags(cmd)
	return cmd
}

func init() {
	rateAgreementsCopiersCmd.AddCommand(newRateAgreementsCopiersListCmd())
}

func initRateAgreementsCopiersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("rate-agreement-template", "", "Filter by template rate agreement ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementsCopiersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRateAgreementsCopiersListOptions(cmd)
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
	query.Set("fields[rate-agreements-copiers]", "note,scheduled-at,processed-at,copiers-results,copiers-errors,rate-agreement-template,broker,created-by,target-customers,target-truckers")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[rate-agreement-template]", opts.RateAgreementTemplate)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreements-copiers", query)
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

	rows := buildRateAgreementsCopierRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRateAgreementsCopiersTable(cmd, rows)
}

func parseRateAgreementsCopiersListOptions(cmd *cobra.Command) (rateAgreementsCopiersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	rateAgreementTemplate, _ := cmd.Flags().GetString("rate-agreement-template")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementsCopiersListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Broker:                broker,
		RateAgreementTemplate: rateAgreementTemplate,
		CreatedBy:             createdBy,
	}, nil
}

func buildRateAgreementsCopierRows(resp jsonAPIResponse) []rateAgreementsCopierRow {
	rows := make([]rateAgreementsCopierRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRateAgreementsCopierRow(resource))
	}
	return rows
}

func buildRateAgreementsCopierRow(resource jsonAPIResource) rateAgreementsCopierRow {
	attrs := resource.Attributes
	row := rateAgreementsCopierRow{
		ID:                  resource.ID,
		Note:                stringAttr(attrs, "note"),
		ScheduledAt:         formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt:         formatDateTime(stringAttr(attrs, "processed-at")),
		CopiersResultsCount: mapLenAttr(attrs, "copiers-results"),
		CopiersErrorsCount:  mapLenAttr(attrs, "copiers-errors"),
	}

	if rel, ok := resource.Relationships["rate-agreement-template"]; ok && rel.Data != nil {
		row.RateAgreementTemplate = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	customerCount := relationshipCount(resource.Relationships["target-customers"])
	truckerCount := relationshipCount(resource.Relationships["target-truckers"])
	if customerCount > 0 && truckerCount > 0 {
		row.TargetType = "mixed"
		row.TargetCount = customerCount + truckerCount
	} else if customerCount > 0 {
		row.TargetType = "customers"
		row.TargetCount = customerCount
	} else if truckerCount > 0 {
		row.TargetType = "truckers"
		row.TargetCount = truckerCount
	}

	return row
}

func renderRateAgreementsCopiersTable(cmd *cobra.Command, rows []rateAgreementsCopierRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rate agreements copiers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTEMPLATE\tBROKER\tCREATED BY\tTARGETS\tSCHEDULED AT\tPROCESSED AT\tRESULTS\tERRORS\tNOTE")
	for _, row := range rows {
		targets := formatRateAgreementsCopierTargets(row.TargetType, row.TargetCount)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
			row.ID,
			row.RateAgreementTemplate,
			row.BrokerID,
			row.CreatedByID,
			targets,
			row.ScheduledAt,
			row.ProcessedAt,
			row.CopiersResultsCount,
			row.CopiersErrorsCount,
			truncateString(row.Note, 40),
		)
	}
	return writer.Flush()
}

func formatRateAgreementsCopierTargets(targetType string, targetCount int) string {
	if targetType == "" {
		return ""
	}
	if targetCount > 0 {
		return fmt.Sprintf("%s:%d", targetType, targetCount)
	}
	return targetType
}

func mapAttr(attrs map[string]any, key string) map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed
	default:
		return nil
	}
}

func mapLenAttr(attrs map[string]any, key string) int {
	return len(mapAttr(attrs, key))
}

func formatMap(value map[string]any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}

func relationshipCount(rel jsonAPIRelationship) int {
	return len(relationshipIDs(rel))
}

func relationshipIDsToStrings(rel jsonAPIRelationship) []string {
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.ID)
	}
	return out
}
