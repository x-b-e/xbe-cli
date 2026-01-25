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

type profitImprovementsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Title                     string
	Description               string
	AmountEstimated           string
	AmountEstimatedMin        string
	AmountEstimatedMax        string
	AmountValidated           string
	AmountValidatedMin        string
	AmountValidatedMax        string
	ImpactFrequencyEstimated  string
	ImpactIntervalEstimated   string
	ImpactFrequencyValidated  string
	ImpactIntervalValidated   string
	ImpactStartOnEstimated    string
	ImpactStartOnEstimatedMin string
	ImpactStartOnEstimatedMax string
	HasImpactStartOnEstimated string
	ImpactEndOnEstimated      string
	ImpactEndOnEstimatedMin   string
	ImpactEndOnEstimatedMax   string
	HasImpactEndOnEstimated   string
	ImpactStartOnValidated    string
	ImpactStartOnValidatedMin string
	ImpactStartOnValidatedMax string
	HasImpactStartOnValidated string
	ImpactEndOnValidated      string
	ImpactEndOnValidatedMin   string
	ImpactEndOnValidatedMax   string
	HasImpactEndOnValidated   string
	GainShareFeeStartOn       string
	GainShareFeeStartOnMin    string
	GainShareFeeStartOnMax    string
	HasGainShareFeeStartOn    string
	GainShareFeeEndOn         string
	GainShareFeeEndOnMin      string
	GainShareFeeEndOnMax      string
	HasGainShareFeeEndOn      string
	GainShareFeePercentage    string
	GainShareFeePercentageMin string
	GainShareFeePercentageMax string
	Status                    string
	ProfitImprovementCategory string
	Original                  string
	CreatedBy                 string
	OwnedBy                   string
	ValidatedBy               string
	Organization              string
	OrganizationID            string
	OrganizationType          string
	NotOrganizationType       string
	Broker                    string
}

type profitImprovementRow struct {
	ID               string `json:"id"`
	Title            string `json:"title,omitempty"`
	Status           string `json:"status,omitempty"`
	AmountEstimated  any    `json:"amount_estimated,omitempty"`
	AmountValidated  any    `json:"amount_validated,omitempty"`
	CategoryID       string `json:"profit_improvement_category_id,omitempty"`
	CategoryName     string `json:"profit_improvement_category_name,omitempty"`
	OwnedByID        string `json:"owned_by_id,omitempty"`
	OwnedByName      string `json:"owned_by_name,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
}

func newProfitImprovementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profit improvements",
		Long: `List profit improvements with filtering and pagination.

Output Columns:
  ID             Profit improvement identifier
  STATUS         Current status
  TITLE          Profit improvement title
  AMOUNT EST     Estimated impact amount
  AMOUNT VAL     Validated impact amount
  CATEGORY       Profit improvement category (name or ID)
  OWNER          Owner (name or ID)
  ORG            Organization (name or Type/ID)

Filters:
  --title                       Filter by title (partial match)
  --description                 Filter by description (partial match)
  --status                      Filter by status
  --profit-improvement-category Filter by category ID
  --original                    Filter by original profit improvement ID
  --created-by                  Filter by creator user ID
  --owned-by                    Filter by owner user ID
  --validated-by                Filter by validator user ID
  --organization                Filter by organization (Type|ID)
  --organization-id             Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type           Filter by organization type
  --not-organization-type       Filter by excluding organization type
  --broker                      Filter by broker ID

Value Filters:
  --amount-estimated, --amount-estimated-min, --amount-estimated-max
  --amount-validated, --amount-validated-min, --amount-validated-max
  --gain-share-fee-percentage, --gain-share-fee-percentage-min, --gain-share-fee-percentage-max

Date Filters (YYYY-MM-DD):
  --impact-start-on-estimated, --impact-start-on-estimated-min, --impact-start-on-estimated-max
  --impact-end-on-estimated, --impact-end-on-estimated-min, --impact-end-on-estimated-max
  --impact-start-on-validated, --impact-start-on-validated-min, --impact-start-on-validated-max
  --impact-end-on-validated, --impact-end-on-validated-min, --impact-end-on-validated-max
  --gain-share-fee-start-on, --gain-share-fee-start-on-min, --gain-share-fee-start-on-max
  --gain-share-fee-end-on, --gain-share-fee-end-on-min, --gain-share-fee-end-on-max

Presence Filters (true/false):
  --has-impact-start-on-estimated
  --has-impact-end-on-estimated
  --has-impact-start-on-validated
  --has-impact-end-on-validated
  --has-gain-share-fee-start-on
  --has-gain-share-fee-end-on

Enum Filters:
  --impact-frequency-estimated (one_time/recurring)
  --impact-interval-estimated (monthly/quarterly/annual)
  --impact-frequency-validated (one_time/recurring)
  --impact-interval-validated (monthly/quarterly/annual)`,
		Example: `  # List profit improvements
  xbe view profit-improvements list

  # Filter by status and category
  xbe view profit-improvements list --status submitted --profit-improvement-category 12

  # Filter by date range
  xbe view profit-improvements list --impact-start-on-estimated-min 2024-01-01 --impact-start-on-estimated-max 2024-12-31

  # Output as JSON
  xbe view profit-improvements list --json`,
		RunE: runProfitImprovementsList,
	}
	initProfitImprovementsListFlags(cmd)
	return cmd
}

func init() {
	profitImprovementsCmd.AddCommand(newProfitImprovementsListCmd())
}

func initProfitImprovementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (e.g., -created-at)")
	cmd.Flags().String("title", "", "Filter by title (partial match)")
	cmd.Flags().String("description", "", "Filter by description (partial match)")
	cmd.Flags().String("amount-estimated", "", "Filter by estimated amount")
	cmd.Flags().String("amount-estimated-min", "", "Filter by minimum estimated amount")
	cmd.Flags().String("amount-estimated-max", "", "Filter by maximum estimated amount")
	cmd.Flags().String("amount-validated", "", "Filter by validated amount")
	cmd.Flags().String("amount-validated-min", "", "Filter by minimum validated amount")
	cmd.Flags().String("amount-validated-max", "", "Filter by maximum validated amount")
	cmd.Flags().String("impact-frequency-estimated", "", "Filter by estimated impact frequency")
	cmd.Flags().String("impact-interval-estimated", "", "Filter by estimated impact interval")
	cmd.Flags().String("impact-frequency-validated", "", "Filter by validated impact frequency")
	cmd.Flags().String("impact-interval-validated", "", "Filter by validated impact interval")
	cmd.Flags().String("impact-start-on-estimated", "", "Filter by estimated impact start date")
	cmd.Flags().String("impact-start-on-estimated-min", "", "Filter by minimum estimated impact start date")
	cmd.Flags().String("impact-start-on-estimated-max", "", "Filter by maximum estimated impact start date")
	cmd.Flags().String("has-impact-start-on-estimated", "", "Filter by presence of estimated impact start date (true/false)")
	cmd.Flags().String("impact-end-on-estimated", "", "Filter by estimated impact end date")
	cmd.Flags().String("impact-end-on-estimated-min", "", "Filter by minimum estimated impact end date")
	cmd.Flags().String("impact-end-on-estimated-max", "", "Filter by maximum estimated impact end date")
	cmd.Flags().String("has-impact-end-on-estimated", "", "Filter by presence of estimated impact end date (true/false)")
	cmd.Flags().String("impact-start-on-validated", "", "Filter by validated impact start date")
	cmd.Flags().String("impact-start-on-validated-min", "", "Filter by minimum validated impact start date")
	cmd.Flags().String("impact-start-on-validated-max", "", "Filter by maximum validated impact start date")
	cmd.Flags().String("has-impact-start-on-validated", "", "Filter by presence of validated impact start date (true/false)")
	cmd.Flags().String("impact-end-on-validated", "", "Filter by validated impact end date")
	cmd.Flags().String("impact-end-on-validated-min", "", "Filter by minimum validated impact end date")
	cmd.Flags().String("impact-end-on-validated-max", "", "Filter by maximum validated impact end date")
	cmd.Flags().String("has-impact-end-on-validated", "", "Filter by presence of validated impact end date (true/false)")
	cmd.Flags().String("gain-share-fee-start-on", "", "Filter by gain share fee start date")
	cmd.Flags().String("gain-share-fee-start-on-min", "", "Filter by minimum gain share fee start date")
	cmd.Flags().String("gain-share-fee-start-on-max", "", "Filter by maximum gain share fee start date")
	cmd.Flags().String("has-gain-share-fee-start-on", "", "Filter by presence of gain share fee start date (true/false)")
	cmd.Flags().String("gain-share-fee-end-on", "", "Filter by gain share fee end date")
	cmd.Flags().String("gain-share-fee-end-on-min", "", "Filter by minimum gain share fee end date")
	cmd.Flags().String("gain-share-fee-end-on-max", "", "Filter by maximum gain share fee end date")
	cmd.Flags().String("has-gain-share-fee-end-on", "", "Filter by presence of gain share fee end date (true/false)")
	cmd.Flags().String("gain-share-fee-percentage", "", "Filter by gain share fee percentage")
	cmd.Flags().String("gain-share-fee-percentage-min", "", "Filter by minimum gain share fee percentage")
	cmd.Flags().String("gain-share-fee-percentage-max", "", "Filter by maximum gain share fee percentage")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("profit-improvement-category", "", "Filter by profit improvement category ID")
	cmd.Flags().String("original", "", "Filter by original profit improvement ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("owned-by", "", "Filter by owner user ID")
	cmd.Flags().String("validated-by", "", "Filter by validator user ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProfitImprovementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProfitImprovementsListOptions(cmd)
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
	query.Set("fields[profit-improvements]", "title,status,amount-estimated,amount-validated,organization")
	query.Set("include", "profit-improvement-category,owned-by,organization")
	query.Set("fields[profit-improvement-categories]", "name")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[title]", opts.Title)
	setFilterIfPresent(query, "filter[description]", opts.Description)
	setFilterIfPresent(query, "filter[amount-estimated]", opts.AmountEstimated)
	setFilterIfPresent(query, "filter[amount-estimated-min]", opts.AmountEstimatedMin)
	setFilterIfPresent(query, "filter[amount-estimated-max]", opts.AmountEstimatedMax)
	setFilterIfPresent(query, "filter[amount-validated]", opts.AmountValidated)
	setFilterIfPresent(query, "filter[amount-validated-min]", opts.AmountValidatedMin)
	setFilterIfPresent(query, "filter[amount-validated-max]", opts.AmountValidatedMax)
	setFilterIfPresent(query, "filter[impact-frequency-estimated]", opts.ImpactFrequencyEstimated)
	setFilterIfPresent(query, "filter[impact-interval-estimated]", opts.ImpactIntervalEstimated)
	setFilterIfPresent(query, "filter[impact-frequency-validated]", opts.ImpactFrequencyValidated)
	setFilterIfPresent(query, "filter[impact-interval-validated]", opts.ImpactIntervalValidated)
	setFilterIfPresent(query, "filter[impact-start-on-estimated]", opts.ImpactStartOnEstimated)
	setFilterIfPresent(query, "filter[impact-start-on-estimated-min]", opts.ImpactStartOnEstimatedMin)
	setFilterIfPresent(query, "filter[impact-start-on-estimated-max]", opts.ImpactStartOnEstimatedMax)
	setFilterIfPresent(query, "filter[has-impact-start-on-estimated]", opts.HasImpactStartOnEstimated)
	setFilterIfPresent(query, "filter[impact-end-on-estimated]", opts.ImpactEndOnEstimated)
	setFilterIfPresent(query, "filter[impact-end-on-estimated-min]", opts.ImpactEndOnEstimatedMin)
	setFilterIfPresent(query, "filter[impact-end-on-estimated-max]", opts.ImpactEndOnEstimatedMax)
	setFilterIfPresent(query, "filter[has-impact-end-on-estimated]", opts.HasImpactEndOnEstimated)
	setFilterIfPresent(query, "filter[impact-start-on-validated]", opts.ImpactStartOnValidated)
	setFilterIfPresent(query, "filter[impact-start-on-validated-min]", opts.ImpactStartOnValidatedMin)
	setFilterIfPresent(query, "filter[impact-start-on-validated-max]", opts.ImpactStartOnValidatedMax)
	setFilterIfPresent(query, "filter[has-impact-start-on-validated]", opts.HasImpactStartOnValidated)
	setFilterIfPresent(query, "filter[impact-end-on-validated]", opts.ImpactEndOnValidated)
	setFilterIfPresent(query, "filter[impact-end-on-validated-min]", opts.ImpactEndOnValidatedMin)
	setFilterIfPresent(query, "filter[impact-end-on-validated-max]", opts.ImpactEndOnValidatedMax)
	setFilterIfPresent(query, "filter[has-impact-end-on-validated]", opts.HasImpactEndOnValidated)
	setFilterIfPresent(query, "filter[gain-share-fee-start-on]", opts.GainShareFeeStartOn)
	setFilterIfPresent(query, "filter[gain-share-fee-start-on-min]", opts.GainShareFeeStartOnMin)
	setFilterIfPresent(query, "filter[gain-share-fee-start-on-max]", opts.GainShareFeeStartOnMax)
	setFilterIfPresent(query, "filter[has-gain-share-fee-start-on]", opts.HasGainShareFeeStartOn)
	setFilterIfPresent(query, "filter[gain-share-fee-end-on]", opts.GainShareFeeEndOn)
	setFilterIfPresent(query, "filter[gain-share-fee-end-on-min]", opts.GainShareFeeEndOnMin)
	setFilterIfPresent(query, "filter[gain-share-fee-end-on-max]", opts.GainShareFeeEndOnMax)
	setFilterIfPresent(query, "filter[has-gain-share-fee-end-on]", opts.HasGainShareFeeEndOn)
	setFilterIfPresent(query, "filter[gain-share-fee-percentage]", opts.GainShareFeePercentage)
	setFilterIfPresent(query, "filter[gain-share-fee-percentage-min]", opts.GainShareFeePercentageMin)
	setFilterIfPresent(query, "filter[gain-share-fee-percentage-max]", opts.GainShareFeePercentageMax)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[profit-improvement-category]", opts.ProfitImprovementCategory)
	setFilterIfPresent(query, "filter[original]", opts.Original)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[owned-by]", opts.OwnedBy)
	setFilterIfPresent(query, "filter[validated-by]", opts.ValidatedBy)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization_id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/profit-improvements", query)
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

	rows := buildProfitImprovementRows(resp)
	if strings.TrimSpace(opts.NotOrganizationType) != "" {
		rows = filterProfitImprovementsByOrganizationType(rows, opts.NotOrganizationType)
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProfitImprovementsTable(cmd, rows)
}

func parseProfitImprovementsListOptions(cmd *cobra.Command) (profitImprovementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	amountEstimated, _ := cmd.Flags().GetString("amount-estimated")
	amountEstimatedMin, _ := cmd.Flags().GetString("amount-estimated-min")
	amountEstimatedMax, _ := cmd.Flags().GetString("amount-estimated-max")
	amountValidated, _ := cmd.Flags().GetString("amount-validated")
	amountValidatedMin, _ := cmd.Flags().GetString("amount-validated-min")
	amountValidatedMax, _ := cmd.Flags().GetString("amount-validated-max")
	impactFrequencyEstimated, _ := cmd.Flags().GetString("impact-frequency-estimated")
	impactIntervalEstimated, _ := cmd.Flags().GetString("impact-interval-estimated")
	impactFrequencyValidated, _ := cmd.Flags().GetString("impact-frequency-validated")
	impactIntervalValidated, _ := cmd.Flags().GetString("impact-interval-validated")
	impactStartOnEstimated, _ := cmd.Flags().GetString("impact-start-on-estimated")
	impactStartOnEstimatedMin, _ := cmd.Flags().GetString("impact-start-on-estimated-min")
	impactStartOnEstimatedMax, _ := cmd.Flags().GetString("impact-start-on-estimated-max")
	hasImpactStartOnEstimated, _ := cmd.Flags().GetString("has-impact-start-on-estimated")
	impactEndOnEstimated, _ := cmd.Flags().GetString("impact-end-on-estimated")
	impactEndOnEstimatedMin, _ := cmd.Flags().GetString("impact-end-on-estimated-min")
	impactEndOnEstimatedMax, _ := cmd.Flags().GetString("impact-end-on-estimated-max")
	hasImpactEndOnEstimated, _ := cmd.Flags().GetString("has-impact-end-on-estimated")
	impactStartOnValidated, _ := cmd.Flags().GetString("impact-start-on-validated")
	impactStartOnValidatedMin, _ := cmd.Flags().GetString("impact-start-on-validated-min")
	impactStartOnValidatedMax, _ := cmd.Flags().GetString("impact-start-on-validated-max")
	hasImpactStartOnValidated, _ := cmd.Flags().GetString("has-impact-start-on-validated")
	impactEndOnValidated, _ := cmd.Flags().GetString("impact-end-on-validated")
	impactEndOnValidatedMin, _ := cmd.Flags().GetString("impact-end-on-validated-min")
	impactEndOnValidatedMax, _ := cmd.Flags().GetString("impact-end-on-validated-max")
	hasImpactEndOnValidated, _ := cmd.Flags().GetString("has-impact-end-on-validated")
	gainShareFeeStartOn, _ := cmd.Flags().GetString("gain-share-fee-start-on")
	gainShareFeeStartOnMin, _ := cmd.Flags().GetString("gain-share-fee-start-on-min")
	gainShareFeeStartOnMax, _ := cmd.Flags().GetString("gain-share-fee-start-on-max")
	hasGainShareFeeStartOn, _ := cmd.Flags().GetString("has-gain-share-fee-start-on")
	gainShareFeeEndOn, _ := cmd.Flags().GetString("gain-share-fee-end-on")
	gainShareFeeEndOnMin, _ := cmd.Flags().GetString("gain-share-fee-end-on-min")
	gainShareFeeEndOnMax, _ := cmd.Flags().GetString("gain-share-fee-end-on-max")
	hasGainShareFeeEndOn, _ := cmd.Flags().GetString("has-gain-share-fee-end-on")
	gainShareFeePercentage, _ := cmd.Flags().GetString("gain-share-fee-percentage")
	gainShareFeePercentageMin, _ := cmd.Flags().GetString("gain-share-fee-percentage-min")
	gainShareFeePercentageMax, _ := cmd.Flags().GetString("gain-share-fee-percentage-max")
	status, _ := cmd.Flags().GetString("status")
	profitImprovementCategory, _ := cmd.Flags().GetString("profit-improvement-category")
	original, _ := cmd.Flags().GetString("original")
	createdBy, _ := cmd.Flags().GetString("created-by")
	ownedBy, _ := cmd.Flags().GetString("owned-by")
	validatedBy, _ := cmd.Flags().GetString("validated-by")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return profitImprovementsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Title:                     title,
		Description:               description,
		AmountEstimated:           amountEstimated,
		AmountEstimatedMin:        amountEstimatedMin,
		AmountEstimatedMax:        amountEstimatedMax,
		AmountValidated:           amountValidated,
		AmountValidatedMin:        amountValidatedMin,
		AmountValidatedMax:        amountValidatedMax,
		ImpactFrequencyEstimated:  impactFrequencyEstimated,
		ImpactIntervalEstimated:   impactIntervalEstimated,
		ImpactFrequencyValidated:  impactFrequencyValidated,
		ImpactIntervalValidated:   impactIntervalValidated,
		ImpactStartOnEstimated:    impactStartOnEstimated,
		ImpactStartOnEstimatedMin: impactStartOnEstimatedMin,
		ImpactStartOnEstimatedMax: impactStartOnEstimatedMax,
		HasImpactStartOnEstimated: hasImpactStartOnEstimated,
		ImpactEndOnEstimated:      impactEndOnEstimated,
		ImpactEndOnEstimatedMin:   impactEndOnEstimatedMin,
		ImpactEndOnEstimatedMax:   impactEndOnEstimatedMax,
		HasImpactEndOnEstimated:   hasImpactEndOnEstimated,
		ImpactStartOnValidated:    impactStartOnValidated,
		ImpactStartOnValidatedMin: impactStartOnValidatedMin,
		ImpactStartOnValidatedMax: impactStartOnValidatedMax,
		HasImpactStartOnValidated: hasImpactStartOnValidated,
		ImpactEndOnValidated:      impactEndOnValidated,
		ImpactEndOnValidatedMin:   impactEndOnValidatedMin,
		ImpactEndOnValidatedMax:   impactEndOnValidatedMax,
		HasImpactEndOnValidated:   hasImpactEndOnValidated,
		GainShareFeeStartOn:       gainShareFeeStartOn,
		GainShareFeeStartOnMin:    gainShareFeeStartOnMin,
		GainShareFeeStartOnMax:    gainShareFeeStartOnMax,
		HasGainShareFeeStartOn:    hasGainShareFeeStartOn,
		GainShareFeeEndOn:         gainShareFeeEndOn,
		GainShareFeeEndOnMin:      gainShareFeeEndOnMin,
		GainShareFeeEndOnMax:      gainShareFeeEndOnMax,
		HasGainShareFeeEndOn:      hasGainShareFeeEndOn,
		GainShareFeePercentage:    gainShareFeePercentage,
		GainShareFeePercentageMin: gainShareFeePercentageMin,
		GainShareFeePercentageMax: gainShareFeePercentageMax,
		Status:                    status,
		ProfitImprovementCategory: profitImprovementCategory,
		Original:                  original,
		CreatedBy:                 createdBy,
		OwnedBy:                   ownedBy,
		ValidatedBy:               validatedBy,
		Organization:              organization,
		OrganizationID:            organizationID,
		OrganizationType:          organizationType,
		NotOrganizationType:       notOrganizationType,
		Broker:                    broker,
	}, nil
}

func buildProfitImprovementRows(resp jsonAPIResponse) []profitImprovementRow {
	included := indexIncludedResources(resp.Included)
	rows := make([]profitImprovementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, profitImprovementRowFromResource(resource, included))
	}
	return rows
}

func profitImprovementRowFromSingle(resp jsonAPISingleResponse) profitImprovementRow {
	included := indexIncludedResources(resp.Included)
	return profitImprovementRowFromResource(resp.Data, included)
}

func profitImprovementRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) profitImprovementRow {
	row := profitImprovementRow{
		ID:              resource.ID,
		Title:           stringAttr(resource.Attributes, "title"),
		Status:          stringAttr(resource.Attributes, "status"),
		AmountEstimated: resource.Attributes["amount-estimated"],
		AmountValidated: resource.Attributes["amount-validated"],
	}

	if rel, ok := resource.Relationships["profit-improvement-category"]; ok && rel.Data != nil {
		row.CategoryID = rel.Data.ID
		if cat, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CategoryName = stringAttr(cat.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["owned-by"]; ok && rel.Data != nil {
		row.OwnedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.OwnedByName = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationID = rel.Data.ID
		row.OrganizationType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.OrganizationName = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	return row
}

func renderProfitImprovementsTable(cmd *cobra.Command, rows []profitImprovementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profit improvements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTITLE\tAMOUNT EST\tAMOUNT VAL\tCATEGORY\tOWNER\tORG")
	for _, row := range rows {
		category := firstNonEmpty(row.CategoryName, row.CategoryID)
		owner := firstNonEmpty(row.OwnedByName, row.OwnedByID)
		org := row.OrganizationName
		if org == "" && row.OrganizationType != "" && row.OrganizationID != "" {
			org = fmt.Sprintf("%s/%s", row.OrganizationType, row.OrganizationID)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Title, 40),
			formatAnyValue(row.AmountEstimated),
			formatAnyValue(row.AmountValidated),
			truncateString(category, 30),
			truncateString(owner, 25),
			truncateString(org, 30),
		)
	}
	return writer.Flush()
}

func filterProfitImprovementsByOrganizationType(rows []profitImprovementRow, organizationType string) []profitImprovementRow {
	filterType := normalizeOrganizationType(organizationType)
	if filterType == "" {
		return rows
	}
	filtered := make([]profitImprovementRow, 0, len(rows))
	for _, row := range rows {
		if normalizeOrganizationType(row.OrganizationType) != filterType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func indexIncludedResources(included []jsonAPIResource) map[string]jsonAPIResource {
	indexed := make(map[string]jsonAPIResource, len(included))
	for _, resource := range included {
		indexed[resourceKey(resource.Type, resource.ID)] = resource
	}
	return indexed
}

func formatAnyValue(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}
