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

type predictionKnowledgeBasesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Broker  string
}

type predictionKnowledgeBaseRow struct {
	ID       string `json:"id"`
	BrokerID string `json:"broker_id,omitempty"`
}

func newPredictionKnowledgeBasesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction knowledge bases",
		Long: `List prediction knowledge bases.

Output Columns:
  ID         Knowledge base identifier
  BROKER ID  Associated broker ID

Filters:
  --broker   Filter by broker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction knowledge bases
  xbe view prediction-knowledge-bases list

  # Filter by broker
  xbe view prediction-knowledge-bases list --broker 123

  # Output as JSON
  xbe view prediction-knowledge-bases list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionKnowledgeBasesList,
	}
	initPredictionKnowledgeBasesListFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBasesCmd.AddCommand(newPredictionKnowledgeBasesListCmd())
}

func initPredictionKnowledgeBasesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBasesList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionKnowledgeBasesListOptions(cmd)
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
	query.Set("include", "broker")
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

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-bases", query)
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

	rows := buildPredictionKnowledgeBaseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionKnowledgeBasesTable(cmd, rows)
}

func parsePredictionKnowledgeBasesListOptions(cmd *cobra.Command) (predictionKnowledgeBasesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionKnowledgeBasesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Broker:  broker,
	}, nil
}

func buildPredictionKnowledgeBaseRows(resp jsonAPIResponse) []predictionKnowledgeBaseRow {
	rows := make([]predictionKnowledgeBaseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, predictionKnowledgeBaseRowFromResource(resource))
	}
	return rows
}

func predictionKnowledgeBaseRowFromResource(resource jsonAPIResource) predictionKnowledgeBaseRow {
	row := predictionKnowledgeBaseRow{
		ID: resource.ID,
	}
	row.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	return row
}

func renderPredictionKnowledgeBasesTable(cmd *cobra.Command, rows []predictionKnowledgeBaseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction knowledge bases found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\n", row.ID, row.BrokerID)
	}
	return writer.Flush()
}
