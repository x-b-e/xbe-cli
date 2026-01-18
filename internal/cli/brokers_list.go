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

type brokersListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	CompanyName string
	IsActive    string
}

type brokerRow struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
}

func newBrokersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List brokers",
		Long:  "List brokers.",
		RunE:  runBrokersList,
	}
	initBrokersListFlags(cmd)
	return cmd
}

func init() {
	brokersCmd.AddCommand(newBrokersListCmd())
}

func initBrokersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("company-name", "", "Filter by company name (partial match)")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokersListOptions(cmd)
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
	query.Set("sort", "company-name")
	query.Set("fields[brokers]", "company-name")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[company-name]", opts.CompanyName)
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)

	body, _, err := client.Get(cmd.Context(), "/v1/brokers", query)
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

	rows := buildBrokerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokersTable(cmd, rows)
}

func parseBrokersListOptions(cmd *cobra.Command) (brokersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return brokersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return brokersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return brokersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return brokersListOptions{}, err
	}
	companyName, err := cmd.Flags().GetString("company-name")
	if err != nil {
		return brokersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetString("is-active")
	if err != nil {
		return brokersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return brokersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return brokersListOptions{}, err
	}

	return brokersListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		CompanyName: companyName,
		IsActive:    isActive,
	}, nil
}

func buildBrokerRows(resp jsonAPIResponse) []brokerRow {
	rows := make([]brokerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, brokerRow{
			ID:          resource.ID,
			CompanyName: strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
		})
	}

	return rows
}

func renderBrokersTable(cmd *cobra.Command, rows []brokerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No brokers found.")
		return nil
	}

	const tableCompanyMax = 80

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tCOMPANY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\n", row.ID, truncateString(row.CompanyName, tableCompanyMax))
	}
	return writer.Flush()
}
