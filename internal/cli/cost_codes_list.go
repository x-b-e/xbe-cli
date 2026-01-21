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

type costCodesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Customer    string
	Trucker     string
	Broker      string
	Code        string
	Q           string
	Description string
	JobNumber   string
}

type costCodeRow struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
}

func newCostCodesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cost codes",
		Long: `List cost codes with filtering and pagination.

Cost codes are used to categorize and track costs for billing and accounting
purposes. They can be associated with specific customers, truckers, or brokers.

Output Columns:
  ID           Cost code identifier
  CODE         Cost code value
  DESCRIPTION  Description of the cost code
  ACTIVE       Whether the cost code is active

Filters:
  --customer     Filter by customer ID
  --trucker      Filter by trucker ID
  --broker       Filter by broker ID
  --code         Filter by code (partial match)
  --q            General search
  --description  Filter by description (partial match)
  --job-number   Filter by job number`,
		Example: `  # List all cost codes
  xbe view cost-codes list

  # Filter by broker
  xbe view cost-codes list --broker 123

  # Search by code
  xbe view cost-codes list --code "MAT-001"

  # General search
  xbe view cost-codes list --q "material"

  # Output as JSON
  xbe view cost-codes list --json`,
		RunE: runCostCodesList,
	}
	initCostCodesListFlags(cmd)
	return cmd
}

func init() {
	costCodesCmd.AddCommand(newCostCodesListCmd())
}

func initCostCodesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("code", "", "Filter by code (partial match)")
	cmd.Flags().String("q", "", "General search")
	cmd.Flags().String("description", "", "Filter by description (partial match)")
	cmd.Flags().String("job-number", "", "Filter by job number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCostCodesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCostCodesListOptions(cmd)
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
	query.Set("sort", "code")
	query.Set("fields[cost-codes]", "code,description,is-active")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[code]", opts.Code)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[description]", opts.Description)
	setFilterIfPresent(query, "filter[job-number]", opts.JobNumber)

	body, _, err := client.Get(cmd.Context(), "/v1/cost-codes", query)
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

	rows := buildCostCodeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCostCodesTable(cmd, rows)
}

func parseCostCodesListOptions(cmd *cobra.Command) (costCodesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	customer, _ := cmd.Flags().GetString("customer")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	code, _ := cmd.Flags().GetString("code")
	q, _ := cmd.Flags().GetString("q")
	description, _ := cmd.Flags().GetString("description")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return costCodesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Customer:    customer,
		Trucker:     trucker,
		Broker:      broker,
		Code:        code,
		Q:           q,
		Description: description,
		JobNumber:   jobNumber,
	}, nil
}

func buildCostCodeRows(resp jsonAPIResponse) []costCodeRow {
	rows := make([]costCodeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := costCodeRow{
			ID:          resource.ID,
			Code:        stringAttr(resource.Attributes, "code"),
			Description: stringAttr(resource.Attributes, "description"),
			IsActive:    boolAttr(resource.Attributes, "is-active"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderCostCodesTable(cmd *cobra.Command, rows []costCodeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No cost codes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCODE\tDESCRIPTION\tACTIVE")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Code, 20),
			truncateString(row.Description, 40),
			active,
		)
	}
	return writer.Flush()
}
