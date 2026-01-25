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

type rateAgreementCopierWorksListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Broker                 string
	RateAgreementTemplate  string
	TargetOrganizationType string
	TargetOrganizationID   string
	CreatedBy              string
}

type rateAgreementCopierWorkRow struct {
	ID                      string `json:"id"`
	RateAgreementTemplateID string `json:"rate_agreement_template_id,omitempty"`
	TargetOrganizationType  string `json:"target_organization_type,omitempty"`
	TargetOrganizationID    string `json:"target_organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
	Note                    string `json:"note,omitempty"`
	ScheduledAt             string `json:"scheduled_at,omitempty"`
	ProcessedAt             string `json:"processed_at,omitempty"`
}

func newRateAgreementCopierWorksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rate agreement copier works",
		Long: `List rate agreement copier works with filtering and pagination.

Output Columns:
  ID            Work identifier
  TEMPLATE      Rate agreement template ID
  TARGET        Target organization (type/id)
  SCHEDULED_AT  Scheduled timestamp
  PROCESSED_AT  Processed timestamp
  CREATED_BY    Creator user ID

Filters:
  --rate-agreement-template   Filter by rate agreement template ID
  --target-organization-type  Filter by target organization type (Customer, Trucker)
  --target-organization-id    Filter by target organization ID
  --broker                    Filter by broker ID
  --created-by                Filter by creator user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List copier works
  xbe view rate-agreement-copier-works list

  # Filter by rate agreement template
  xbe view rate-agreement-copier-works list --rate-agreement-template 123

  # Filter by target organization
  xbe view rate-agreement-copier-works list --target-organization-type Customer --target-organization-id 456

  # Output as JSON
  xbe view rate-agreement-copier-works list --json`,
		Args: cobra.NoArgs,
		RunE: runRateAgreementCopierWorksList,
	}
	initRateAgreementCopierWorksListFlags(cmd)
	return cmd
}

func init() {
	rateAgreementCopierWorksCmd.AddCommand(newRateAgreementCopierWorksListCmd())
}

func initRateAgreementCopierWorksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("rate-agreement-template", "", "Filter by rate agreement template ID")
	cmd.Flags().String("target-organization-type", "", "Filter by target organization type (Customer, Trucker)")
	cmd.Flags().String("target-organization-id", "", "Filter by target organization ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementCopierWorksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRateAgreementCopierWorksListOptions(cmd)
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
	query.Set("fields[rate-agreement-copier-works]", "note,scheduled-at,processed-at,rate-agreement-template,target-organization,broker,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[rate_agreement_template]", opts.RateAgreementTemplate)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	targetType := normalizeTargetOrganizationTypeForFilter(opts.TargetOrganizationType)
	if targetType != "" && opts.TargetOrganizationID != "" {
		query.Set("filter[target_organization]", targetType+"|"+opts.TargetOrganizationID)
	} else if targetType != "" {
		setFilterIfPresent(query, "filter[target_organization_type]", targetType)
	} else if opts.TargetOrganizationID != "" {
		setFilterIfPresent(query, "filter[target_organization_id]", opts.TargetOrganizationID)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreement-copier-works", query)
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

	rows := buildRateAgreementCopierWorkRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRateAgreementCopierWorksTable(cmd, rows)
}

func parseRateAgreementCopierWorksListOptions(cmd *cobra.Command) (rateAgreementCopierWorksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	rateAgreementTemplate, _ := cmd.Flags().GetString("rate-agreement-template")
	targetOrganizationType, _ := cmd.Flags().GetString("target-organization-type")
	targetOrganizationID, _ := cmd.Flags().GetString("target-organization-id")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementCopierWorksListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		RateAgreementTemplate:  rateAgreementTemplate,
		TargetOrganizationType: targetOrganizationType,
		TargetOrganizationID:   targetOrganizationID,
		Broker:                 broker,
		CreatedBy:              createdBy,
	}, nil
}

func buildRateAgreementCopierWorkRows(resp jsonAPIResponse) []rateAgreementCopierWorkRow {
	rows := make([]rateAgreementCopierWorkRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rateAgreementCopierWorkRow{
			ID:          resource.ID,
			Note:        stringAttr(attrs, "note"),
			ScheduledAt: formatDateTime(stringAttr(attrs, "scheduled-at")),
			ProcessedAt: formatDateTime(stringAttr(attrs, "processed-at")),
		}

		if rel, ok := resource.Relationships["rate-agreement-template"]; ok && rel.Data != nil {
			row.RateAgreementTemplateID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["target-organization"]; ok && rel.Data != nil {
			row.TargetOrganizationType = rel.Data.Type
			row.TargetOrganizationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRateAgreementCopierWorksTable(cmd *cobra.Command, rows []rateAgreementCopierWorkRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rate agreement copier works found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTEMPLATE\tTARGET\tSCHEDULED_AT\tPROCESSED_AT\tCREATED_BY")
	for _, row := range rows {
		target := ""
		if row.TargetOrganizationType != "" && row.TargetOrganizationID != "" {
			target = row.TargetOrganizationType + "/" + row.TargetOrganizationID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RateAgreementTemplateID,
			target,
			row.ScheduledAt,
			row.ProcessedAt,
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
