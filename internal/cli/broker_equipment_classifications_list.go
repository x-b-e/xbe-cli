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

type brokerEquipmentClassificationsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	CreatedAtMin            string
	CreatedAtMax            string
	IsCreatedAt             string
	UpdatedAtMin            string
	UpdatedAtMax            string
	IsUpdatedAt             string
	Broker                  string
	EquipmentClassification string
}

type brokerEquipmentClassificationRow struct {
	ID                        string `json:"id"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
}

func newBrokerEquipmentClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker equipment classifications",
		Long: `List broker equipment classifications with filtering and pagination.

Broker equipment classifications link brokers to equipment classifications
they can use. Equipment classifications must be non-root (have a parent).

Output Columns:
  ID                 Broker equipment classification identifier
  BROKER ID          Broker ID
  EQUIP CLASS ID     Equipment classification ID

Filters:
  --broker                   Filter by broker ID
  --equipment-classification Filter by equipment classification ID
  --created-at-min           Filter by created-at on/after (ISO 8601)
  --created-at-max           Filter by created-at on/before (ISO 8601)
  --is-created-at            Filter by has created-at (true/false)
  --updated-at-min           Filter by updated-at on/after (ISO 8601)
  --updated-at-max           Filter by updated-at on/before (ISO 8601)
  --is-updated-at            Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List broker equipment classifications
  xbe view broker-equipment-classifications list

  # Filter by broker
  xbe view broker-equipment-classifications list --broker 123

  # Output as JSON
  xbe view broker-equipment-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runBrokerEquipmentClassificationsList,
	}
	initBrokerEquipmentClassificationsListFlags(cmd)
	return cmd
}

func init() {
	brokerEquipmentClassificationsCmd.AddCommand(newBrokerEquipmentClassificationsListCmd())
}

func initBrokerEquipmentClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("equipment-classification", "", "Filter by equipment classification ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerEquipmentClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerEquipmentClassificationsListOptions(cmd)
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
	query.Set("fields[broker-equipment-classifications]", "broker,equipment-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[equipment_classification]", opts.EquipmentClassification)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-equipment-classifications", query)
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

	rows := buildBrokerEquipmentClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerEquipmentClassificationsTable(cmd, rows)
}

func parseBrokerEquipmentClassificationsListOptions(cmd *cobra.Command) (brokerEquipmentClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerEquipmentClassificationsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		Broker:                  broker,
		EquipmentClassification: equipmentClassification,
		CreatedAtMin:            createdAtMin,
		CreatedAtMax:            createdAtMax,
		IsCreatedAt:             isCreatedAt,
		UpdatedAtMin:            updatedAtMin,
		UpdatedAtMax:            updatedAtMax,
		IsUpdatedAt:             isUpdatedAt,
	}, nil
}

func buildBrokerEquipmentClassificationRows(resp jsonAPIResponse) []brokerEquipmentClassificationRow {
	rows := make([]brokerEquipmentClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBrokerEquipmentClassificationRow(resource))
	}
	return rows
}

func buildBrokerEquipmentClassificationRow(resource jsonAPIResource) brokerEquipmentClassificationRow {
	row := brokerEquipmentClassificationRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["equipment-classification"]; ok && rel.Data != nil {
		row.EquipmentClassificationID = rel.Data.ID
	}

	return row
}

func brokerEquipmentClassificationRowFromSingle(resp jsonAPISingleResponse) brokerEquipmentClassificationRow {
	return buildBrokerEquipmentClassificationRow(resp.Data)
}

func renderBrokerEquipmentClassificationsTable(cmd *cobra.Command, rows []brokerEquipmentClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker equipment classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBROKER ID\tEQUIP CLASS ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.BrokerID,
			row.EquipmentClassificationID,
		)
	}
	return writer.Flush()
}
