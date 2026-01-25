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

type projectCostCodesListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Project     string
	Customer    string
	CostCode    string
	Code        string
	Query       string
	Description string
}

type projectCostCodeRow struct {
	ID                          string `json:"id"`
	ExplicitCostCodeDescription string `json:"explicit_cost_code_description,omitempty"`
	CostCodeDescription         string `json:"cost_code_description,omitempty"`
	ProjectCustomerID           string `json:"project_customer_id,omitempty"`
	CostCodeID                  string `json:"cost_code_id,omitempty"`
	ProjectID                   string `json:"project_id,omitempty"`
	CustomerID                  string `json:"customer_id,omitempty"`
}

func newProjectCostCodesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project cost codes",
		Long: `List project cost codes.

Output Columns:
  ID              Project cost code identifier
  DESCRIPTION     Cost code description
  PROJECT         Project ID
  COST CODE       Cost code ID

Filters:
  --project       Filter by project ID
  --customer      Filter by customer ID
  --cost-code     Filter by cost code ID
  --code          Filter by code
  --query         Search query
  --description   Filter by description`,
		Example: `  # List all project cost codes
  xbe view project-cost-codes list

  # Filter by project
  xbe view project-cost-codes list --project 123

  # Filter by customer
  xbe view project-cost-codes list --customer 456

  # Search
  xbe view project-cost-codes list --query "labor"

  # Output as JSON
  xbe view project-cost-codes list --json`,
		RunE: runProjectCostCodesList,
	}
	initProjectCostCodesListFlags(cmd)
	return cmd
}

func init() {
	projectCostCodesCmd.AddCommand(newProjectCostCodesListCmd())
}

func initProjectCostCodesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("cost-code", "", "Filter by cost code ID")
	cmd.Flags().String("code", "", "Filter by code")
	cmd.Flags().String("query", "", "Search query")
	cmd.Flags().String("description", "", "Filter by description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectCostCodesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectCostCodesListOptions(cmd)
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
	query.Set("include", "project,cost-code")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[cost_code]", opts.CostCode)
	setFilterIfPresent(query, "filter[code]", opts.Code)
	setFilterIfPresent(query, "filter[q]", opts.Query)
	setFilterIfPresent(query, "filter[description]", opts.Description)

	body, _, err := client.Get(cmd.Context(), "/v1/project-cost-codes", query)
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

	rows := buildProjectCostCodeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectCostCodesTable(cmd, rows)
}

func parseProjectCostCodesListOptions(cmd *cobra.Command) (projectCostCodesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	project, _ := cmd.Flags().GetString("project")
	customer, _ := cmd.Flags().GetString("customer")
	costCode, _ := cmd.Flags().GetString("cost-code")
	code, _ := cmd.Flags().GetString("code")
	queryStr, _ := cmd.Flags().GetString("query")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectCostCodesListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Project:     project,
		Customer:    customer,
		CostCode:    costCode,
		Code:        code,
		Query:       queryStr,
		Description: description,
	}, nil
}

func buildProjectCostCodeRows(resp jsonAPIResponse) []projectCostCodeRow {
	rows := make([]projectCostCodeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectCostCodeRow{
			ID:                          resource.ID,
			ExplicitCostCodeDescription: stringAttr(resource.Attributes, "explicit-cost-code-description"),
			CostCodeDescription:         stringAttr(resource.Attributes, "cost-code-description"),
		}

		if rel, ok := resource.Relationships["project-customer"]; ok && rel.Data != nil {
			row.ProjectCustomerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["cost-code"]; ok && rel.Data != nil {
			row.CostCodeID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectCostCodesTable(cmd *cobra.Command, rows []projectCostCodeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project cost codes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tPROJECT\tCOST CODE")
	for _, row := range rows {
		description := row.CostCodeDescription
		if row.ExplicitCostCodeDescription != "" {
			description = row.ExplicitCostCodeDescription
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(description, 40),
			row.ProjectID,
			row.CostCodeID,
		)
	}
	return writer.Flush()
}
