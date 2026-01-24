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

type predictionSubjectGapsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
	Status            string
	GapType           string
}

type predictionSubjectGapRow struct {
	ID                               string   `json:"id"`
	GapType                          string   `json:"gap_type,omitempty"`
	Status                           string   `json:"status,omitempty"`
	PrimaryAmount                    *float64 `json:"primary_amount,omitempty"`
	SecondaryAmount                  *float64 `json:"secondary_amount,omitempty"`
	GapAmount                        *float64 `json:"gap_amount,omitempty"`
	ExplainedGapAmount               *float64 `json:"explained_gap_amount,omitempty"`
	PredictionSubjectID              string   `json:"prediction_subject_id,omitempty"`
	PredictionSubjectName            string   `json:"prediction_subject_name,omitempty"`
	PredictionSubjectReferenceNumber string   `json:"prediction_subject_reference_number,omitempty"`
}

func newPredictionSubjectGapsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subject gaps",
		Long: `List prediction subject gaps with filtering and pagination.

Prediction subject gaps describe the difference between primary and secondary
amounts for a prediction subject, along with status and type.

Output Columns:
  ID         Prediction subject gap identifier
  TYPE       Gap type
  STATUS     Gap status
  PRIMARY    Primary amount
  SECONDARY  Secondary amount
  GAP        Gap amount
  EXPLAINED  Explained gap amount
  SUBJECT    Prediction subject name or ID

Filters:
  --prediction-subject  Filter by prediction subject ID
  --status              Filter by status (pending, approved)
  --gap-type            Filter by gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subject gaps
  xbe view prediction-subject-gaps list

  # Filter by prediction subject
  xbe view prediction-subject-gaps list --prediction-subject 123

  # Filter by status
  xbe view prediction-subject-gaps list --status approved

  # Filter by gap type
  xbe view prediction-subject-gaps list --gap-type actual_vs_consensus

  # Output as JSON
  xbe view prediction-subject-gaps list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectGapsList,
	}
	initPredictionSubjectGapsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectGapsCmd.AddCommand(newPredictionSubjectGapsListCmd())
}

func initPredictionSubjectGapsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("status", "", "Filter by status (pending, approved)")
	cmd.Flags().String("gap-type", "", "Filter by gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectGapsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectGapsListOptions(cmd)
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
	query.Set("fields[prediction-subject-gaps]", "gap-type,status,primary-amount,secondary-amount,gap-amount,explained-gap-amount,prediction-subject")
	query.Set("include", "prediction-subject")
	query.Set("fields[prediction-subjects]", "name,reference-number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[prediction-subject]", opts.PredictionSubject)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[gap-type]", opts.GapType)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-gaps", query)
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

	rows := buildPredictionSubjectGapRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectGapsTable(cmd, rows)
}

func parsePredictionSubjectGapsListOptions(cmd *cobra.Command) (predictionSubjectGapsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	status, _ := cmd.Flags().GetString("status")
	gapType, _ := cmd.Flags().GetString("gap-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectGapsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
		Status:            status,
		GapType:           gapType,
	}, nil
}

func buildPredictionSubjectGapRows(resp jsonAPIResponse) []predictionSubjectGapRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]predictionSubjectGapRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := predictionSubjectGapRow{ID: resource.ID}
		row.GapType = stringAttr(attrs, "gap-type")
		row.Status = stringAttr(attrs, "status")

		if amount, ok := floatAttrValue(attrs, "primary-amount"); ok {
			row.PrimaryAmount = &amount
		}
		if amount, ok := floatAttrValue(attrs, "secondary-amount"); ok {
			row.SecondaryAmount = &amount
		}
		if amount, ok := floatAttrValue(attrs, "gap-amount"); ok {
			row.GapAmount = &amount
		}
		if amount, ok := floatAttrValue(attrs, "explained-gap-amount"); ok {
			row.ExplainedGapAmount = &amount
		}

		if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
			row.PredictionSubjectID = rel.Data.ID
			if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.PredictionSubjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				row.PredictionSubjectReferenceNumber = strings.TrimSpace(stringAttr(inc.Attributes, "reference-number"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPredictionSubjectGapsTable(cmd *cobra.Command, rows []predictionSubjectGapRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subject gaps found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tSTATUS\tPRIMARY\tSECONDARY\tGAP\tEXPLAINED\tSUBJECT")
	for _, row := range rows {
		subjectLabel := firstNonEmpty(
			row.PredictionSubjectName,
			row.PredictionSubjectReferenceNumber,
			row.PredictionSubjectID,
		)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.GapType,
			row.Status,
			formatPredictionSubjectGapAmount(row.PrimaryAmount),
			formatPredictionSubjectGapAmount(row.SecondaryAmount),
			formatPredictionSubjectGapAmount(row.GapAmount),
			formatPredictionSubjectGapAmount(row.ExplainedGapAmount),
			subjectLabel,
		)
	}
	return writer.Flush()
}

func predictionSubjectGapRowFromSingle(resp jsonAPISingleResponse) predictionSubjectGapRow {
	attrs := resp.Data.Attributes
	row := predictionSubjectGapRow{ID: resp.Data.ID}
	row.GapType = stringAttr(attrs, "gap-type")
	row.Status = stringAttr(attrs, "status")
	if amount, ok := floatAttrValue(attrs, "primary-amount"); ok {
		row.PrimaryAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "secondary-amount"); ok {
		row.SecondaryAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "gap-amount"); ok {
		row.GapAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "explained-gap-amount"); ok {
		row.ExplainedGapAmount = &amount
	}
	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}
	return row
}

func formatPredictionSubjectGapAmount(amount *float64) string {
	if amount == nil {
		return ""
	}
	return strconv.FormatFloat(*amount, 'f', -1, 64)
}
