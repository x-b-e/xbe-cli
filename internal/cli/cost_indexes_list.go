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

type costIndexesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Broker           string
	IsBroker         string
	IsExpired        string
	IsValidForBroker string
	ExpiredAtMin     string
	ExpiredAtMax     string
}

func newCostIndexesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cost indexes",
		Long: `List cost indexes with filtering and pagination.

Cost indexes define pricing indexes that can be used for rate adjustments.

Output Columns:
  ID          Index identifier
  NAME        Index name
  DESCRIPTION Description
  BROKER      Broker organization (if broker-specific)
  EXPIRED     Whether the index has expired

Filters:
  --broker      Filter by broker ID
  --is-broker   Filter by broker presence (true/false)
  --is-expired  Filter by expiration status (true/false)`,
		Example: `  # List all cost indexes
  xbe view cost-indexes list

  # Filter by broker
  xbe view cost-indexes list --broker 123

  # Show only global indexes (no broker)
  xbe view cost-indexes list --is-broker false

  # Show only expired indexes
  xbe view cost-indexes list --is-expired true

  # Output as JSON
  xbe view cost-indexes list --json`,
		RunE: runCostIndexesList,
	}
	initCostIndexesListFlags(cmd)
	return cmd
}

func init() {
	costIndexesCmd.AddCommand(newCostIndexesListCmd())
}

func initCostIndexesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("is-broker", "", "Filter by broker presence (true/false)")
	cmd.Flags().String("is-expired", "", "Filter by expiration status (true/false)")
	cmd.Flags().String("is-valid-for-broker", "", "Filter by validity for broker (true/false)")
	cmd.Flags().String("expired-at-min", "", "Filter by minimum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("expired-at-max", "", "Filter by maximum expiration date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCostIndexesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCostIndexesListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[cost-indexes]", "name,description,url,expired-at,is-expired,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[is-broker]", opts.IsBroker)
	setFilterIfPresent(query, "filter[is-expired]", opts.IsExpired)
	setFilterIfPresent(query, "filter[is-valid-for-broker]", opts.IsValidForBroker)
	setFilterIfPresent(query, "filter[expired-at-min]", opts.ExpiredAtMin)
	setFilterIfPresent(query, "filter[expired-at-max]", opts.ExpiredAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/cost-indexes", query)
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

	rows := buildCostIndexRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCostIndexesTable(cmd, rows)
}

func parseCostIndexesListOptions(cmd *cobra.Command) (costIndexesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	isBroker, _ := cmd.Flags().GetString("is-broker")
	isExpired, _ := cmd.Flags().GetString("is-expired")
	isValidForBroker, _ := cmd.Flags().GetString("is-valid-for-broker")
	expiredAtMin, _ := cmd.Flags().GetString("expired-at-min")
	expiredAtMax, _ := cmd.Flags().GetString("expired-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return costIndexesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Broker:           broker,
		IsBroker:         isBroker,
		IsExpired:        isExpired,
		IsValidForBroker: isValidForBroker,
		ExpiredAtMin:     expiredAtMin,
		ExpiredAtMax:     expiredAtMax,
	}, nil
}

type costIndexRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	ExpiredAt   string `json:"expired_at,omitempty"`
	IsExpired   bool   `json:"is_expired"`
	BrokerID    string `json:"broker_id,omitempty"`
	BrokerName  string `json:"broker_name,omitempty"`
}

func buildCostIndexRows(resp jsonAPIResponse) []costIndexRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]costIndexRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := costIndexRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			Description: stringAttr(resource.Attributes, "description"),
			URL:         stringAttr(resource.Attributes, "url"),
			ExpiredAt:   stringAttr(resource.Attributes, "expired-at"),
			IsExpired:   boolAttr(resource.Attributes, "is-expired"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCostIndexesTable(cmd *cobra.Command, rows []costIndexRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No cost indexes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tBROKER\tEXPIRED")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}
		expired := "No"
		if row.IsExpired {
			expired = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Description, 30),
			truncateString(broker, 20),
			expired,
		)
	}
	return writer.Flush()
}
