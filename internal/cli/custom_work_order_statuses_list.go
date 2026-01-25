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

type customWorkOrderStatusesListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Broker        string
	PrimaryStatus string
}

func newCustomWorkOrderStatusesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom work order statuses",
		Long: `List custom work order statuses with filtering and pagination.

Custom work order statuses allow brokers to define their own status labels
with colors that map to primary statuses.

Output Columns:
  ID              Status identifier
  LABEL           Display label
  PRIMARY STATUS  Underlying status (pending, in_progress, completed, etc.)
  COLOR           Hex color code
  BROKER          Broker organization

Filters:
  --broker          Filter by broker ID
  --primary-status  Filter by primary status`,
		Example: `  # List all custom work order statuses
  xbe view custom-work-order-statuses list

  # Filter by broker
  xbe view custom-work-order-statuses list --broker 123

  # Filter by primary status
  xbe view custom-work-order-statuses list --primary-status pending

  # Output as JSON
  xbe view custom-work-order-statuses list --json`,
		RunE: runCustomWorkOrderStatusesList,
	}
	initCustomWorkOrderStatusesListFlags(cmd)
	return cmd
}

func init() {
	customWorkOrderStatusesCmd.AddCommand(newCustomWorkOrderStatusesListCmd())
}

func initCustomWorkOrderStatusesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("primary-status", "", "Filter by primary status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomWorkOrderStatusesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomWorkOrderStatusesListOptions(cmd)
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
	query.Set("sort", "label")
	query.Set("fields[custom-work-order-statuses]", "label,description,color-hex,primary-status,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[primary-status]", opts.PrimaryStatus)

	body, _, err := client.Get(cmd.Context(), "/v1/custom-work-order-statuses", query)
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

	rows := buildCustomWorkOrderStatusRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomWorkOrderStatusesTable(cmd, rows)
}

func parseCustomWorkOrderStatusesListOptions(cmd *cobra.Command) (customWorkOrderStatusesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	primaryStatus, _ := cmd.Flags().GetString("primary-status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customWorkOrderStatusesListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Broker:        broker,
		PrimaryStatus: primaryStatus,
	}, nil
}

type customWorkOrderStatusRow struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Description   string `json:"description,omitempty"`
	ColorHex      string `json:"color_hex,omitempty"`
	PrimaryStatus string `json:"primary_status"`
	BrokerID      string `json:"broker_id,omitempty"`
	BrokerName    string `json:"broker_name,omitempty"`
}

func buildCustomWorkOrderStatusRows(resp jsonAPIResponse) []customWorkOrderStatusRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]customWorkOrderStatusRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := customWorkOrderStatusRow{
			ID:            resource.ID,
			Label:         stringAttr(resource.Attributes, "label"),
			Description:   stringAttr(resource.Attributes, "description"),
			ColorHex:      stringAttr(resource.Attributes, "color-hex"),
			PrimaryStatus: stringAttr(resource.Attributes, "primary-status"),
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

func renderCustomWorkOrderStatusesTable(cmd *cobra.Command, rows []customWorkOrderStatusRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No custom work order statuses found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLABEL\tPRIMARY STATUS\tCOLOR\tBROKER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Label, 25),
			row.PrimaryStatus,
			row.ColorHex,
			truncateString(broker, 25),
		)
	}
	return writer.Flush()
}
