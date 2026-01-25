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

type brokerCommitmentsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Status       string
	BrokerID     string
	Broker       string
	TruckerID    string
	Trucker      string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type brokerCommitmentRow struct {
	ID           string `json:"id"`
	Status       string `json:"status,omitempty"`
	Label        string `json:"label,omitempty"`
	Broker       string `json:"broker,omitempty"`
	BrokerID     string `json:"broker_id,omitempty"`
	Trucker      string `json:"trucker,omitempty"`
	TruckerID    string `json:"trucker_id,omitempty"`
	TruckScopeID string `json:"truck_scope_id,omitempty"`
}

func newBrokerCommitmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker commitments",
		Long: `List broker commitments with filtering and pagination.

Broker commitments link brokers (buyers) with truckers (sellers) for capacity or service needs.

Output Columns:
  ID          Commitment identifier
  STATUS      Commitment status
  BROKER      Broker company name
  TRUCKER     Trucker company name
  LABEL       Commitment label
  TRUCK SCOPE Truck scope ID

Filters:
  --status       Filter by status (editing, active, inactive)
  --broker-id    Filter by broker ID (uses broker_id filter)
  --broker       Filter by broker ID
  --trucker-id   Filter by trucker ID (uses trucker_id filter)
  --trucker      Filter by trucker ID
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List broker commitments
  xbe view broker-commitments list

  # Filter by status
  xbe view broker-commitments list --status active

  # Filter by broker
  xbe view broker-commitments list --broker 123

  # Filter by trucker
  xbe view broker-commitments list --trucker 456

  # Output as JSON
  xbe view broker-commitments list --json`,
		RunE: runBrokerCommitmentsList,
	}
	initBrokerCommitmentsListFlags(cmd)
	return cmd
}

func init() {
	brokerCommitmentsCmd.AddCommand(newBrokerCommitmentsListCmd())
}

func initBrokerCommitmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("status", "", "Filter by status (editing, active, inactive)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (uses broker_id filter)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (uses trucker_id filter)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerCommitmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerCommitmentsListOptions(cmd)
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
	query.Set("fields[broker-commitments]", "status,label,buyer,seller,truck-scope")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("include", "buyer,seller")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker_id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-commitments", query)
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

	rows := buildBrokerCommitmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerCommitmentsTable(cmd, rows)
}

func parseBrokerCommitmentsListOptions(cmd *cobra.Command) (brokerCommitmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	broker, _ := cmd.Flags().GetString("broker")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerCommitmentsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Status:       status,
		BrokerID:     brokerID,
		Broker:       broker,
		TruckerID:    truckerID,
		Trucker:      trucker,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildBrokerCommitmentRows(resp jsonAPIResponse) []brokerCommitmentRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]brokerCommitmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := brokerCommitmentRow{
			ID:     resource.ID,
			Status: stringAttr(resource.Attributes, "status"),
			Label:  stringAttr(resource.Attributes, "label"),
		}

		if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil && rel.Data.Type == "brokers" {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = firstNonEmpty(
					stringAttr(broker.Attributes, "company-name"),
					stringAttr(broker.Attributes, "name"),
				)
			}
		} else if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = firstNonEmpty(
					stringAttr(broker.Attributes, "company-name"),
					stringAttr(broker.Attributes, "name"),
				)
			}
		}

		if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil && rel.Data.Type == "truckers" {
			row.TruckerID = rel.Data.ID
			if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Trucker = firstNonEmpty(
					stringAttr(trucker.Attributes, "company-name"),
					stringAttr(trucker.Attributes, "name"),
				)
			}
		} else if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
			if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Trucker = firstNonEmpty(
					stringAttr(trucker.Attributes, "company-name"),
					stringAttr(trucker.Attributes, "name"),
				)
			}
		}

		row.TruckScopeID = relationshipIDFromMap(resource.Relationships, "truck-scope")

		rows = append(rows, row)
	}
	return rows
}

func renderBrokerCommitmentsTable(cmd *cobra.Command, rows []brokerCommitmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker commitments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tBROKER\tTRUCKER\tLABEL\tTRUCK SCOPE")
	for _, row := range rows {
		broker := row.Broker
		if broker == "" {
			broker = row.BrokerID
		}
		trucker := row.Trucker
		if trucker == "" {
			trucker = row.TruckerID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(broker, 25),
			truncateString(trucker, 25),
			truncateString(row.Label, 20),
			truncateString(row.TruckScopeID, 12),
		)
	}
	return writer.Flush()
}
