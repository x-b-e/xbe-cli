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

type predictionSubjectGapPortionsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	PredictionSubjectGap string
	Status               string
}

type predictionSubjectGapPortionRow struct {
	ID                     string `json:"id"`
	Name                   string `json:"name,omitempty"`
	Amount                 any    `json:"amount,omitempty"`
	Status                 string `json:"status,omitempty"`
	Description            string `json:"description,omitempty"`
	PredictionSubjectGapID string `json:"prediction_subject_gap_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
}

func newPredictionSubjectGapPortionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subject gap portions",
		Long: `List prediction subject gap portions with filtering and pagination.

Output Columns:
  ID        Prediction subject gap portion identifier
  NAME      Portion name
  AMOUNT    Portion amount
  STATUS    Portion status
  GAP       Prediction subject gap ID
  NOTE      Portion description (truncated)

Filters:
  --prediction-subject-gap  Filter by prediction subject gap ID
  --status                  Filter by status (draft/approved)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subject gap portions
  xbe view prediction-subject-gap-portions list

  # Filter by prediction subject gap
  xbe view prediction-subject-gap-portions list --prediction-subject-gap 123

  # Filter by status
  xbe view prediction-subject-gap-portions list --status approved

  # Output as JSON
  xbe view prediction-subject-gap-portions list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectGapPortionsList,
	}
	initPredictionSubjectGapPortionsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectGapPortionsCmd.AddCommand(newPredictionSubjectGapPortionsListCmd())
}

func initPredictionSubjectGapPortionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject-gap", "", "Filter by prediction subject gap ID")
	cmd.Flags().String("status", "", "Filter by status (draft/approved)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectGapPortionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectGapPortionsListOptions(cmd)
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
	query.Set("fields[prediction-subject-gap-portions]", "name,amount,description,status,prediction-subject-gap,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[prediction_subject_gap]", opts.PredictionSubjectGap)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-gap-portions", query)
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

	rows := buildPredictionSubjectGapPortionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectGapPortionsTable(cmd, rows)
}

func parsePredictionSubjectGapPortionsListOptions(cmd *cobra.Command) (predictionSubjectGapPortionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubjectGap, _ := cmd.Flags().GetString("prediction-subject-gap")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectGapPortionsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		PredictionSubjectGap: predictionSubjectGap,
		Status:               status,
	}, nil
}

func buildPredictionSubjectGapPortionRows(resp jsonAPIResponse) []predictionSubjectGapPortionRow {
	rows := make([]predictionSubjectGapPortionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPredictionSubjectGapPortionRow(resource))
	}
	return rows
}

func predictionSubjectGapPortionRowFromSingle(resp jsonAPISingleResponse) predictionSubjectGapPortionRow {
	return buildPredictionSubjectGapPortionRow(resp.Data)
}

func buildPredictionSubjectGapPortionRow(resource jsonAPIResource) predictionSubjectGapPortionRow {
	attrs := resource.Attributes
	row := predictionSubjectGapPortionRow{
		ID:          resource.ID,
		Name:        stringAttr(attrs, "name"),
		Amount:      attrs["amount"],
		Status:      stringAttr(attrs, "status"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["prediction-subject-gap"]; ok && rel.Data != nil {
		row.PredictionSubjectGapID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderPredictionSubjectGapPortionsTable(cmd *cobra.Command, rows []predictionSubjectGapPortionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subject gap portions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tAMOUNT\tSTATUS\tGAP\tNOTE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Name,
			formatAnyValue(row.Amount),
			row.Status,
			row.PredictionSubjectGapID,
			truncateString(row.Description, 30),
		)
	}
	return writer.Flush()
}
