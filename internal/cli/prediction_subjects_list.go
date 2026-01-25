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

type predictionSubjectsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Parent              string
	BusinessUnit        string
	Broker              string
	CreatedBy           string
	ActualDueAt         string
	ActualDueAtMin      string
	ActualDueAtMax      string
	HasActualDueAt      string
	PredictionsDueAt    string
	PredictionsDueAtMin string
	PredictionsDueAtMax string
	HasPredictionsDueAt string
	Actual              string
	ActualMin           string
	ActualMax           string
	Status              string
	Name                string
	ReferenceNumber     string
	TaggedWith          string
	TaggedWithAny       string
	TaggedWithAll       string
	InTagCategory       string
	Bidder              string
	LowestBidAmountMin  string
	LowestBidAmountMax  string
}

type predictionSubjectRow struct {
	ID               string `json:"id"`
	Name             string `json:"name,omitempty"`
	Status           string `json:"status,omitempty"`
	Kind             string `json:"kind,omitempty"`
	PredictionsDueAt string `json:"predictions_due_at,omitempty"`
	ActualDueAt      string `json:"actual_due_at,omitempty"`
	Actual           string `json:"actual,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	ParentType       string `json:"parent_type,omitempty"`
	ParentID         string `json:"parent_id,omitempty"`
}

func newPredictionSubjectsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subjects",
		Long: `List prediction subjects.

Output Columns:
  ID                Prediction subject identifier
  NAME              Subject name
  STATUS            Current status
  KIND              Prediction kind
  PREDICTIONS DUE   When predictions are due
  ACTUAL DUE        When the actual value is due
  ACTUAL            Actual value (if set)
  BROKER ID         Associated broker ID
  PARENT            Parent type and ID

Filters:
  --parent                 Filter by parent (e.g., Project|123 or Broker|456)
  --business-unit          Filter by business unit ID
  --broker                 Filter by broker ID
  --created-by             Filter by created-by user ID
  --actual-due-at           Filter by actual due date (YYYY-MM-DD)
  --actual-due-at-min       Filter by minimum actual due date
  --actual-due-at-max       Filter by maximum actual due date
  --has-actual-due-at       Filter by presence of actual due date (true/false)
  --predictions-due-at      Filter by predictions due date (YYYY-MM-DD)
  --predictions-due-at-min  Filter by minimum predictions due date
  --predictions-due-at-max  Filter by maximum predictions due date
  --has-predictions-due-at  Filter by presence of predictions due date (true/false)
  --actual                  Filter by actual value (exact)
  --actual-min              Filter by minimum actual value
  --actual-max              Filter by maximum actual value
  --status                  Filter by status (default excludes abandoned)
  --name                    Filter by name (fuzzy match)
  --reference-number        Filter by reference number
  --tagged-with             Filter by tag name (comma-separated)
  --tagged-with-any         Filter by any tag names (comma-separated)
  --tagged-with-all         Filter by all tag names (comma-separated)
  --in-tag-category         Filter by tag category slug (comma-separated)
  --bidder                  Filter by bidder ID (comma-separated)
  --lowest-bid-amount-min   Filter by minimum lowest bid amount
  --lowest-bid-amount-max   Filter by maximum lowest bid amount

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subjects
  xbe view prediction-subjects list

  # Filter by broker
  xbe view prediction-subjects list --broker 123

  # Filter by parent
  xbe view prediction-subjects list --parent Project|456

  # Filter by date range
  xbe view prediction-subjects list --predictions-due-at-min 2025-01-01 --predictions-due-at-max 2025-12-31

  # Filter by tag
  xbe view prediction-subjects list --tagged-with "urgent"

  # Output as JSON
  xbe view prediction-subjects list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectsList,
	}
	initPredictionSubjectsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectsCmd.AddCommand(newPredictionSubjectsListCmd())
}

func initPredictionSubjectsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("parent", "", "Filter by parent (e.g., Project|123)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("actual-due-at", "", "Filter by actual due date (YYYY-MM-DD)")
	cmd.Flags().String("actual-due-at-min", "", "Filter by minimum actual due date (YYYY-MM-DD)")
	cmd.Flags().String("actual-due-at-max", "", "Filter by maximum actual due date (YYYY-MM-DD)")
	cmd.Flags().String("has-actual-due-at", "", "Filter by presence of actual due date (true/false)")
	cmd.Flags().String("predictions-due-at", "", "Filter by predictions due date (YYYY-MM-DD)")
	cmd.Flags().String("predictions-due-at-min", "", "Filter by minimum predictions due date (YYYY-MM-DD)")
	cmd.Flags().String("predictions-due-at-max", "", "Filter by maximum predictions due date (YYYY-MM-DD)")
	cmd.Flags().String("has-predictions-due-at", "", "Filter by presence of predictions due date (true/false)")
	cmd.Flags().String("actual", "", "Filter by actual value (exact)")
	cmd.Flags().String("actual-min", "", "Filter by minimum actual value")
	cmd.Flags().String("actual-max", "", "Filter by maximum actual value")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("reference-number", "", "Filter by reference number")
	cmd.Flags().String("tagged-with", "", "Filter by tag name (comma-separated)")
	cmd.Flags().String("tagged-with-any", "", "Filter by any tag names (comma-separated)")
	cmd.Flags().String("tagged-with-all", "", "Filter by all tag names (comma-separated)")
	cmd.Flags().String("in-tag-category", "", "Filter by tag category slug (comma-separated)")
	cmd.Flags().String("bidder", "", "Filter by bidder ID (comma-separated)")
	cmd.Flags().String("lowest-bid-amount-min", "", "Filter by minimum lowest bid amount")
	cmd.Flags().String("lowest-bid-amount-max", "", "Filter by maximum lowest bid amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectsListOptions(cmd)
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
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[parent]", opts.Parent)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[actual-due-at]", opts.ActualDueAt)
	setFilterIfPresent(query, "filter[actual-due-at-min]", opts.ActualDueAtMin)
	setFilterIfPresent(query, "filter[actual-due-at-max]", opts.ActualDueAtMax)
	setFilterIfPresent(query, "filter[has-actual-due-at]", opts.HasActualDueAt)
	setFilterIfPresent(query, "filter[predictions-due-at]", opts.PredictionsDueAt)
	setFilterIfPresent(query, "filter[predictions-due-at-min]", opts.PredictionsDueAtMin)
	setFilterIfPresent(query, "filter[predictions-due-at-max]", opts.PredictionsDueAtMax)
	setFilterIfPresent(query, "filter[has-predictions-due-at]", opts.HasPredictionsDueAt)
	setFilterIfPresent(query, "filter[actual]", opts.Actual)
	setFilterIfPresent(query, "filter[actual-min]", opts.ActualMin)
	setFilterIfPresent(query, "filter[actual-max]", opts.ActualMax)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[reference-number]", opts.ReferenceNumber)
	setFilterIfPresent(query, "filter[tagged-with]", opts.TaggedWith)
	setFilterIfPresent(query, "filter[tagged-with-any]", opts.TaggedWithAny)
	setFilterIfPresent(query, "filter[tagged-with-all]", opts.TaggedWithAll)
	setFilterIfPresent(query, "filter[in-tag-category]", opts.InTagCategory)
	setFilterIfPresent(query, "filter[bidder]", opts.Bidder)
	setFilterIfPresent(query, "filter[lowest-bid-amount-min]", opts.LowestBidAmountMin)
	setFilterIfPresent(query, "filter[lowest-bid-amount-max]", opts.LowestBidAmountMax)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subjects", query)
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

	rows := buildPredictionSubjectRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectsTable(cmd, rows)
}

func parsePredictionSubjectsListOptions(cmd *cobra.Command) (predictionSubjectsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	parent, _ := cmd.Flags().GetString("parent")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	actualDueAt, _ := cmd.Flags().GetString("actual-due-at")
	actualDueAtMin, _ := cmd.Flags().GetString("actual-due-at-min")
	actualDueAtMax, _ := cmd.Flags().GetString("actual-due-at-max")
	hasActualDueAt, _ := cmd.Flags().GetString("has-actual-due-at")
	predictionsDueAt, _ := cmd.Flags().GetString("predictions-due-at")
	predictionsDueAtMin, _ := cmd.Flags().GetString("predictions-due-at-min")
	predictionsDueAtMax, _ := cmd.Flags().GetString("predictions-due-at-max")
	hasPredictionsDueAt, _ := cmd.Flags().GetString("has-predictions-due-at")
	actual, _ := cmd.Flags().GetString("actual")
	actualMin, _ := cmd.Flags().GetString("actual-min")
	actualMax, _ := cmd.Flags().GetString("actual-max")
	status, _ := cmd.Flags().GetString("status")
	name, _ := cmd.Flags().GetString("name")
	referenceNumber, _ := cmd.Flags().GetString("reference-number")
	taggedWith, _ := cmd.Flags().GetString("tagged-with")
	taggedWithAny, _ := cmd.Flags().GetString("tagged-with-any")
	taggedWithAll, _ := cmd.Flags().GetString("tagged-with-all")
	inTagCategory, _ := cmd.Flags().GetString("in-tag-category")
	bidder, _ := cmd.Flags().GetString("bidder")
	lowestBidAmountMin, _ := cmd.Flags().GetString("lowest-bid-amount-min")
	lowestBidAmountMax, _ := cmd.Flags().GetString("lowest-bid-amount-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Parent:              parent,
		BusinessUnit:        businessUnit,
		Broker:              broker,
		CreatedBy:           createdBy,
		ActualDueAt:         actualDueAt,
		ActualDueAtMin:      actualDueAtMin,
		ActualDueAtMax:      actualDueAtMax,
		HasActualDueAt:      hasActualDueAt,
		PredictionsDueAt:    predictionsDueAt,
		PredictionsDueAtMin: predictionsDueAtMin,
		PredictionsDueAtMax: predictionsDueAtMax,
		HasPredictionsDueAt: hasPredictionsDueAt,
		Actual:              actual,
		ActualMin:           actualMin,
		ActualMax:           actualMax,
		Status:              status,
		Name:                name,
		ReferenceNumber:     referenceNumber,
		TaggedWith:          taggedWith,
		TaggedWithAny:       taggedWithAny,
		TaggedWithAll:       taggedWithAll,
		InTagCategory:       inTagCategory,
		Bidder:              bidder,
		LowestBidAmountMin:  lowestBidAmountMin,
		LowestBidAmountMax:  lowestBidAmountMax,
	}, nil
}

func buildPredictionSubjectRows(resp jsonAPIResponse) []predictionSubjectRow {
	rows := make([]predictionSubjectRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, predictionSubjectRowFromResource(resource))
	}
	return rows
}

func predictionSubjectRowFromResource(resource jsonAPIResource) predictionSubjectRow {
	row := predictionSubjectRow{
		ID:               resource.ID,
		Name:             stringAttr(resource.Attributes, "name"),
		Status:           stringAttr(resource.Attributes, "status"),
		Kind:             stringAttr(resource.Attributes, "kind"),
		PredictionsDueAt: stringAttr(resource.Attributes, "predictions-due-at"),
		ActualDueAt:      stringAttr(resource.Attributes, "actual-due-at"),
		Actual:           stringAttr(resource.Attributes, "actual"),
		BrokerID:         relationshipIDFromMap(resource.Relationships, "broker"),
	}

	if rel, ok := resource.Relationships["parent"]; ok && rel.Data != nil {
		row.ParentType = rel.Data.Type
		row.ParentID = rel.Data.ID
	}

	return row
}

func buildPredictionSubjectRowFromSingle(resp jsonAPISingleResponse) predictionSubjectRow {
	return predictionSubjectRowFromResource(resp.Data)
}

func renderPredictionSubjectsTable(cmd *cobra.Command, rows []predictionSubjectRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subjects found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSTATUS\tKIND\tPREDICTIONS DUE\tACTUAL DUE\tACTUAL\tBROKER ID\tPARENT")
	for _, row := range rows {
		parent := ""
		if row.ParentType != "" && row.ParentID != "" {
			parent = row.ParentType + "/" + row.ParentID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			row.Status,
			row.Kind,
			row.PredictionsDueAt,
			row.ActualDueAt,
			row.Actual,
			row.BrokerID,
			parent,
		)
	}
	return writer.Flush()
}
