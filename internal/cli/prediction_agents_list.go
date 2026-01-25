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

type predictionAgentsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
	CreatedBy         string
	HasPrediction     string
}

type predictionAgentRow struct {
	ID                  string `json:"id"`
	PredictionSubjectID string `json:"prediction_subject_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	PredictionID        string `json:"prediction_id,omitempty"`
	HasPrediction       bool   `json:"has_prediction"`
}

func newPredictionAgentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction agents",
		Long: `List prediction agents with filtering and pagination.

Output Columns:
  ID            Prediction agent identifier
  SUBJECT       Prediction subject ID
  CREATED BY    Creator user ID
  HAS PRED      Whether the agent has a prediction
  PREDICTION    Prediction ID (if present)

Filters:
  --prediction-subject  Filter by prediction subject ID
  --created-by          Filter by creator user ID
  --has-prediction      Filter by having a prediction (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction agents
  xbe view prediction-agents list

  # Filter by prediction subject
  xbe view prediction-agents list --prediction-subject 123

  # Filter by creator
  xbe view prediction-agents list --created-by 456

  # Filter by whether a prediction exists
  xbe view prediction-agents list --has-prediction true

  # Output as JSON
  xbe view prediction-agents list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionAgentsList,
	}
	initPredictionAgentsListFlags(cmd)
	return cmd
}

func init() {
	predictionAgentsCmd.AddCommand(newPredictionAgentsListCmd())
}

func initPredictionAgentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("has-prediction", "", "Filter by having a prediction (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionAgentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionAgentsListOptions(cmd)
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
	query.Set("fields[prediction-agents]", "prediction-subject,created-by,prediction")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[prediction_subject]", opts.PredictionSubject)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[has_prediction]", opts.HasPrediction)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-agents", query)
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

	rows := buildPredictionAgentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionAgentsTable(cmd, rows)
}

func parsePredictionAgentsListOptions(cmd *cobra.Command) (predictionAgentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	createdBy, _ := cmd.Flags().GetString("created-by")
	hasPrediction, _ := cmd.Flags().GetString("has-prediction")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionAgentsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
		CreatedBy:         createdBy,
		HasPrediction:     hasPrediction,
	}, nil
}

func buildPredictionAgentRows(resp jsonAPIResponse) []predictionAgentRow {
	rows := make([]predictionAgentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPredictionAgentRow(resource))
	}
	return rows
}

func predictionAgentRowFromSingle(resp jsonAPISingleResponse) predictionAgentRow {
	return buildPredictionAgentRow(resp.Data)
}

func buildPredictionAgentRow(resource jsonAPIResource) predictionAgentRow {
	row := predictionAgentRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["prediction"]; ok && rel.Data != nil {
		row.PredictionID = rel.Data.ID
		row.HasPrediction = true
	}

	return row
}

func renderPredictionAgentsTable(cmd *cobra.Command, rows []predictionAgentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction agents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUBJECT\tCREATED BY\tHAS PRED\tPREDICTION")
	for _, row := range rows {
		hasPrediction := "no"
		if row.HasPrediction {
			hasPrediction = "yes"
		}
		predictionID := row.PredictionID
		if predictionID == "" {
			predictionID = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PredictionSubjectID,
			row.CreatedByID,
			hasPrediction,
			predictionID,
		)
	}
	return writer.Flush()
}
