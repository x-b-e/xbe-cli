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

type objectiveStakeholderClassificationsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Objective                 string
	StakeholderClassification string
	InterestDegree            string
}

type objectiveStakeholderClassificationRow struct {
	ID                             string   `json:"id"`
	ObjectiveID                    string   `json:"objective_id,omitempty"`
	ObjectiveName                  string   `json:"objective_name,omitempty"`
	StakeholderClassificationID    string   `json:"stakeholder_classification_id,omitempty"`
	StakeholderClassificationTitle string   `json:"stakeholder_classification_title,omitempty"`
	InterestDegree                 *float64 `json:"interest_degree,omitempty"`
}

func newObjectiveStakeholderClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objective stakeholder classifications",
		Long: `List objective stakeholder classifications with filtering and pagination.

Objective stakeholder classifications link objective templates to stakeholder
classifications with an interest degree between 0 and 1.

Output Columns:
  ID          Classification identifier
  OBJECTIVE   Objective name (or ID)
  STAKEHOLDER Stakeholder classification title (or ID)
  INTEREST    Interest degree (0-1)

Filters:
  --objective                   Filter by objective ID
  --stakeholder-classification  Filter by stakeholder classification ID
  --interest-degree             Filter by interest degree (0-1)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List objective stakeholder classifications
  xbe view objective-stakeholder-classifications list

  # Filter by objective
  xbe view objective-stakeholder-classifications list --objective 123

  # Filter by stakeholder classification
  xbe view objective-stakeholder-classifications list --stakeholder-classification 456

  # Filter by interest degree
  xbe view objective-stakeholder-classifications list --interest-degree 0.5

  # Output JSON
  xbe view objective-stakeholder-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runObjectiveStakeholderClassificationsList,
	}
	initObjectiveStakeholderClassificationsListFlags(cmd)
	return cmd
}

func init() {
	objectiveStakeholderClassificationsCmd.AddCommand(newObjectiveStakeholderClassificationsListCmd())
}

func initObjectiveStakeholderClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("objective", "", "Filter by objective ID")
	cmd.Flags().String("stakeholder-classification", "", "Filter by stakeholder classification ID")
	cmd.Flags().String("interest-degree", "", "Filter by interest degree (0-1)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveStakeholderClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseObjectiveStakeholderClassificationsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[objective-stakeholder-classifications]", "interest-degree,objective,stakeholder-classification")
	query.Set("include", "objective,stakeholder-classification")
	query.Set("fields[objectives]", "name")
	query.Set("fields[stakeholder-classifications]", "title")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[objective]", opts.Objective)
	setFilterIfPresent(query, "filter[stakeholder-classification]", opts.StakeholderClassification)
	setFilterIfPresent(query, "filter[interest-degree]", opts.InterestDegree)

	body, _, err := client.Get(cmd.Context(), "/v1/objective-stakeholder-classifications", query)
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

	rows := buildObjectiveStakeholderClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderObjectiveStakeholderClassificationsTable(cmd, rows)
}

func parseObjectiveStakeholderClassificationsListOptions(cmd *cobra.Command) (objectiveStakeholderClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	objective, _ := cmd.Flags().GetString("objective")
	stakeholderClassification, _ := cmd.Flags().GetString("stakeholder-classification")
	interestDegree, _ := cmd.Flags().GetString("interest-degree")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveStakeholderClassificationsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Objective:                 objective,
		StakeholderClassification: stakeholderClassification,
		InterestDegree:            interestDegree,
	}, nil
}

func buildObjectiveStakeholderClassificationRows(resp jsonAPIResponse) []objectiveStakeholderClassificationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]objectiveStakeholderClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := objectiveStakeholderClassificationRow{
			ID:             resource.ID,
			InterestDegree: floatAttrPointer(resource.Attributes, "interest-degree"),
		}

		row.ObjectiveID = relationshipIDFromMap(resource.Relationships, "objective")
		if row.ObjectiveID != "" {
			if inc, ok := included[resourceKey("objectives", row.ObjectiveID)]; ok {
				row.ObjectiveName = stringAttr(inc.Attributes, "name")
			}
		}

		row.StakeholderClassificationID = relationshipIDFromMap(resource.Relationships, "stakeholder-classification")
		if row.StakeholderClassificationID != "" {
			if inc, ok := included[resourceKey("stakeholder-classifications", row.StakeholderClassificationID)]; ok {
				row.StakeholderClassificationTitle = stringAttr(inc.Attributes, "title")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderObjectiveStakeholderClassificationsTable(cmd *cobra.Command, rows []objectiveStakeholderClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No objective stakeholder classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tOBJECTIVE\tSTAKEHOLDER\tINTEREST")
	for _, row := range rows {
		objective := row.ObjectiveName
		if objective == "" {
			objective = row.ObjectiveID
		}
		stakeholder := row.StakeholderClassificationTitle
		if stakeholder == "" {
			stakeholder = row.StakeholderClassificationID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(objective, 30),
			truncateString(stakeholder, 30),
			formatInterestDegree(row.InterestDegree),
		)
	}

	return writer.Flush()
}

func formatInterestDegree(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *value)
}
