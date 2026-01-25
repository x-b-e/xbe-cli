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

type keyResultChangesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	KeyResult    string
	Objective    string
	Organization string
	Broker       string
	ChangedBy    string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type keyResultChangeRow struct {
	ID               string `json:"id"`
	KeyResultID      string `json:"key_result_id,omitempty"`
	ObjectiveID      string `json:"objective_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	ChangedByID      string `json:"changed_by_id,omitempty"`
	StartOnOld       string `json:"start_on_old,omitempty"`
	StartOnNew       string `json:"start_on_new,omitempty"`
	EndOnOld         string `json:"end_on_old,omitempty"`
	EndOnNew         string `json:"end_on_new,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newKeyResultChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key result changes",
		Long: `List key result changes with filtering and pagination.

Output Columns:
  ID          Change identifier
  KEY_RESULT  Key result ID
  OBJECTIVE   Objective ID
  START_OLD   Previous start date
  START_NEW   Updated start date
  END_OLD     Previous end date
  END_NEW     Updated end date
  CHANGED_BY  User ID who made the change (if present)
  CREATED_AT  When the change was recorded

Filters:
  --key-result     Filter by key result ID (comma-separated for multiple)
  --objective      Filter by objective ID (comma-separated for multiple)
  --organization   Filter by organization (format: Type|ID, e.g., Broker|123)
  --broker         Filter by broker ID (comma-separated for multiple)
  --changed-by     Filter by changed by user ID (comma-separated for multiple)
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List key result changes
  xbe view key-result-changes list

  # Filter by key result
  xbe view key-result-changes list --key-result 123

  # Filter by objective
  xbe view key-result-changes list --objective 456

  # Filter by organization
  xbe view key-result-changes list --organization "Broker|123"

  # Output as JSON
  xbe view key-result-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runKeyResultChangesList,
	}
	initKeyResultChangesListFlags(cmd)
	return cmd
}

func init() {
	keyResultChangesCmd.AddCommand(newKeyResultChangesListCmd())
}

func initKeyResultChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("key-result", "", "Filter by key result ID (comma-separated for multiple)")
	cmd.Flags().String("objective", "", "Filter by objective ID (comma-separated for multiple)")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("changed-by", "", "Filter by changed by user ID (comma-separated for multiple)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseKeyResultChangesListOptions(cmd)
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
	query.Set("fields[key-result-changes]", "start-on-old,start-on-new,end-on-old,end-on-new,created-at,updated-at,key-result,objective,broker,organization,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[key-result]", opts.KeyResult)
	setFilterIfPresent(query, "filter[objective]", opts.Objective)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[changed-by]", opts.ChangedBy)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/key-result-changes", query)
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

	rows := buildKeyResultChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderKeyResultChangesTable(cmd, rows)
}

func parseKeyResultChangesListOptions(cmd *cobra.Command) (keyResultChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	keyResult, _ := cmd.Flags().GetString("key-result")
	objective, _ := cmd.Flags().GetString("objective")
	organization, _ := cmd.Flags().GetString("organization")
	broker, _ := cmd.Flags().GetString("broker")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultChangesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		KeyResult:    keyResult,
		Objective:    objective,
		Organization: organization,
		Broker:       broker,
		ChangedBy:    changedBy,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildKeyResultChangeRows(resp jsonAPIResponse) []keyResultChangeRow {
	rows := make([]keyResultChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := keyResultChangeRow{
			ID:         resource.ID,
			StartOnOld: formatDate(stringAttr(attrs, "start-on-old")),
			StartOnNew: formatDate(stringAttr(attrs, "start-on-new")),
			EndOnOld:   formatDate(stringAttr(attrs, "end-on-old")),
			EndOnNew:   formatDate(stringAttr(attrs, "end-on-new")),
			CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
		}

		row.KeyResultID = relationshipIDFromMap(resource.Relationships, "key-result")
		row.ObjectiveID = relationshipIDFromMap(resource.Relationships, "objective")
		row.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
		row.ChangedByID = relationshipIDFromMap(resource.Relationships, "changed-by")

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationID = rel.Data.ID
			row.OrganizationType = rel.Data.Type
		}

		rows = append(rows, row)
	}
	return rows
}

func renderKeyResultChangesTable(cmd *cobra.Command, rows []keyResultChangeRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No key result changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tKEY_RESULT\tOBJECTIVE\tSTART_OLD\tSTART_NEW\tEND_OLD\tEND_NEW\tCHANGED_BY\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.KeyResultID,
			row.ObjectiveID,
			row.StartOnOld,
			row.StartOnNew,
			row.EndOnOld,
			row.EndOnNew,
			row.ChangedByID,
			row.CreatedAt,
		)
	}

	return writer.Flush()
}
