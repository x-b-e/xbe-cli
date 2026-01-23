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

type workOrderServiceCodesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Broker  string
}

type workOrderServiceCodeRow struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	BrokerName  string `json:"broker_name,omitempty"`
}

func newWorkOrderServiceCodesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work order service codes",
		Long: `List work order service codes with filtering and pagination.

Work order service codes describe the service categories used on work orders
and are scoped to brokers.

Output Columns:
  ID           Service code identifier
  CODE         Service code value
  DESCRIPTION  Description of the service code
  BROKER       Broker organization

Filters:
  --broker  Filter by broker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List work order service codes
  xbe view work-order-service-codes list

  # Filter by broker
  xbe view work-order-service-codes list --broker 123

  # Output as JSON
  xbe view work-order-service-codes list --json`,
		Args: cobra.NoArgs,
		RunE: runWorkOrderServiceCodesList,
	}
	initWorkOrderServiceCodesListFlags(cmd)
	return cmd
}

func init() {
	workOrderServiceCodesCmd.AddCommand(newWorkOrderServiceCodesListCmd())
}

func initWorkOrderServiceCodesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runWorkOrderServiceCodesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseWorkOrderServiceCodesListOptions(cmd)
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
	query.Set("fields[work-order-service-codes]", "code,description,broker")
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "code")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/work-order-service-codes", query)
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

	rows := buildWorkOrderServiceCodeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderWorkOrderServiceCodesTable(cmd, rows)
}

func parseWorkOrderServiceCodesListOptions(cmd *cobra.Command) (workOrderServiceCodesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return workOrderServiceCodesListOptions{
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

func buildWorkOrderServiceCodeRows(resp jsonAPIResponse) []workOrderServiceCodeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]workOrderServiceCodeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := workOrderServiceCodeRow{
			ID:          resource.ID,
			Code:        stringAttr(resource.Attributes, "code"),
			Description: stringAttr(resource.Attributes, "description"),
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

func renderWorkOrderServiceCodesTable(cmd *cobra.Command, rows []workOrderServiceCodeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No work order service codes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCODE\tDESCRIPTION\tBROKER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" && row.BrokerID != "" {
			broker = row.BrokerID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Code, 20),
			truncateString(row.Description, 40),
			truncateString(broker, 25),
		)
	}
	return writer.Flush()
}
