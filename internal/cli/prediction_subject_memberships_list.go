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

type predictionSubjectMembershipsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	PredictionSubject string
	User              string
}

type predictionSubjectMembershipRow struct {
	ID                                         string `json:"id"`
	PredictionSubjectID                        string `json:"prediction_subject_id,omitempty"`
	UserID                                     string `json:"user_id,omitempty"`
	CreatedByID                                string `json:"created_by_id,omitempty"`
	CanManageMemberships                       bool   `json:"can_manage_memberships"`
	CanSeePredictionsWithoutCreatingPrediction bool   `json:"can_see_predictions_without_creating_prediction"`
	CanUpdatePredictionConsensus               bool   `json:"can_update_prediction_consensus"`
	CanUpdateOrDestroyOthersPredictions        bool   `json:"can_update_or_destroy_others_predictions"`
	CanManageGaps                              bool   `json:"can_manage_gaps"`
}

func newPredictionSubjectMembershipsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction subject memberships",
		Long: `List prediction subject memberships with filtering and pagination.

Output Columns:
  ID             Prediction subject membership identifier
  SUBJECT        Prediction subject ID
  USER           User ID
  MANAGE         Can manage memberships
  SEE PRED       Can see predictions without creating one
  CONSENSUS      Can update prediction consensus
  UPDATE OTHERS  Can update or destroy others' predictions
  GAPS           Can manage gaps

Filters:
  --prediction-subject  Filter by prediction subject ID
  --user                Filter by user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction subject memberships
  xbe view prediction-subject-memberships list

  # Filter by prediction subject
  xbe view prediction-subject-memberships list --prediction-subject 123

  # Filter by user
  xbe view prediction-subject-memberships list --user 456

  # Output as JSON
  xbe view prediction-subject-memberships list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionSubjectMembershipsList,
	}
	initPredictionSubjectMembershipsListFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectMembershipsCmd.AddCommand(newPredictionSubjectMembershipsListCmd())
}

func initPredictionSubjectMembershipsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectMembershipsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionSubjectMembershipsListOptions(cmd)
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
	query.Set("fields[prediction-subject-memberships]", "can-manage-memberships,can-see-predictions-without-creating-prediction,can-update-prediction-consensus,can-update-or-destroy-others-predictions,can-manage-gaps,prediction-subject,user,created-by")

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
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-memberships", query)
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

	rows := buildPredictionSubjectMembershipRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionSubjectMembershipsTable(cmd, rows)
}

func parsePredictionSubjectMembershipsListOptions(cmd *cobra.Command) (predictionSubjectMembershipsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectMembershipsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		PredictionSubject: predictionSubject,
		User:              user,
	}, nil
}

func buildPredictionSubjectMembershipRows(resp jsonAPIResponse) []predictionSubjectMembershipRow {
	rows := make([]predictionSubjectMembershipRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPredictionSubjectMembershipRow(resource))
	}
	return rows
}

func predictionSubjectMembershipRowFromSingle(resp jsonAPISingleResponse) predictionSubjectMembershipRow {
	return buildPredictionSubjectMembershipRow(resp.Data)
}

func buildPredictionSubjectMembershipRow(resource jsonAPIResource) predictionSubjectMembershipRow {
	attrs := resource.Attributes
	row := predictionSubjectMembershipRow{
		ID:                   resource.ID,
		CanManageMemberships: boolAttr(attrs, "can-manage-memberships"),
		CanSeePredictionsWithoutCreatingPrediction: boolAttr(attrs, "can-see-predictions-without-creating-prediction"),
		CanUpdatePredictionConsensus:               boolAttr(attrs, "can-update-prediction-consensus"),
		CanUpdateOrDestroyOthersPredictions:        boolAttr(attrs, "can-update-or-destroy-others-predictions"),
		CanManageGaps:                              boolAttr(attrs, "can-manage-gaps"),
	}

	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderPredictionSubjectMembershipsTable(cmd *cobra.Command, rows []predictionSubjectMembershipRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction subject memberships found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUBJECT\tUSER\tMANAGE\tSEE PRED\tCONSENSUS\tUPDATE OTHERS\tGAPS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PredictionSubjectID,
			row.UserID,
			yesNoLower(row.CanManageMemberships),
			yesNoLower(row.CanSeePredictionsWithoutCreatingPrediction),
			yesNoLower(row.CanUpdatePredictionConsensus),
			yesNoLower(row.CanUpdateOrDestroyOthersPredictions),
			yesNoLower(row.CanManageGaps),
		)
	}
	return writer.Flush()
}

func yesNoLower(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
