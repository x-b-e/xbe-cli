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

type customerApplicationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	Status       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type customerApplicationRow struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name,omitempty"`
	Status      string `json:"status,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	BrokerName  string `json:"broker_name,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	UserName    string `json:"user_name,omitempty"`
}

func newCustomerApplicationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer applications",
		Long: `List customer applications with filtering and pagination.

Output Columns:
  ID       Application identifier
  COMPANY  Applicant company name
  STATUS   Application status
  BROKER   Broker name
  USER     Applicant user

Filters:
  --status         Filter by status (pending, reviewing, denied, approved)
  --broker         Filter by broker ID (comma-separated for multiple)
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer applications
  xbe view customer-applications list

  # Filter by status
  xbe view customer-applications list --status pending

  # Filter by broker
  xbe view customer-applications list --broker 123

  # Output as JSON
  xbe view customer-applications list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerApplicationsList,
	}
	initCustomerApplicationsListFlags(cmd)
	return cmd
}

func init() {
	customerApplicationsCmd.AddCommand(newCustomerApplicationsListCmd())
}

func initCustomerApplicationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (pending, reviewing, denied, approved)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerApplicationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerApplicationsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-applications]", "company-name,status,broker,user")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name,email-address")
	query.Set("include", "broker,user")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-applications", query)
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

	rows := buildCustomerApplicationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerApplicationsList(cmd, rows)
}

func parseCustomerApplicationsListOptions(cmd *cobra.Command) (customerApplicationsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	updatedAtMin, err := cmd.Flags().GetString("updated-at-min")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	updatedAtMax, err := cmd.Flags().GetString("updated-at-max")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return customerApplicationsListOptions{}, err
	}

	return customerApplicationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Status:       status,
		Broker:       broker,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildCustomerApplicationRows(resp jsonAPIResponse) []customerApplicationRow {
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]customerApplicationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		brokerID := ""
		brokerName := ""
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			brokerID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				brokerName = strings.TrimSpace(stringAttr(attrs, "company-name"))
			}
		}

		userID := ""
		userName := ""
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			userID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				userName = strings.TrimSpace(stringAttr(attrs, "name"))
				if userName == "" {
					userName = strings.TrimSpace(stringAttr(attrs, "email-address"))
				}
			}
		}

		rows = append(rows, customerApplicationRow{
			ID:          resource.ID,
			CompanyName: strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
			Status:      stringAttr(resource.Attributes, "status"),
			BrokerID:    brokerID,
			BrokerName:  brokerName,
			UserID:      userID,
			UserName:    userName,
		})
	}

	return rows
}

func renderCustomerApplicationsList(cmd *cobra.Command, rows []customerApplicationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer applications found.")
		return nil
	}

	const nameMax = 45
	const brokerMax = 30
	const userMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOMPANY\tSTATUS\tBROKER\tUSER")
	for _, row := range rows {
		broker := row.BrokerName
		if broker == "" {
			broker = row.BrokerID
		}
		user := row.UserName
		if user == "" {
			user = row.UserID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.CompanyName, nameMax),
			row.Status,
			truncateString(broker, brokerMax),
			truncateString(user, userMax),
		)
	}
	return writer.Flush()
}
