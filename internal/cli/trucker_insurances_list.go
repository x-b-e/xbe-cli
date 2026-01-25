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

type truckerInsurancesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Trucker string
}

type truckerInsuranceRow struct {
	ID                   string `json:"id"`
	CompanyName          string `json:"company_name,omitempty"`
	ContactName          string `json:"contact_name,omitempty"`
	PhoneNumber          string `json:"phone_number,omitempty"`
	PhoneNumberFormatted string `json:"phone_number_formatted,omitempty"`
	TruckerID            string `json:"trucker_id,omitempty"`
}

func newTruckerInsurancesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker insurances",
		Long: `List trucker insurances.

Output Columns:
  ID              Trucker insurance identifier
  COMPANY         Insurance company name
  CONTACT         Contact name
  PHONE           Phone number
  TRUCKER         Trucker ID

Filters:
  --trucker       Filter by trucker ID`,
		Example: `  # List all trucker insurances
  xbe view trucker-insurances list

  # Filter by trucker
  xbe view trucker-insurances list --trucker 123

  # Output as JSON
  xbe view trucker-insurances list --json`,
		RunE: runTruckerInsurancesList,
	}
	initTruckerInsurancesListFlags(cmd)
	return cmd
}

func init() {
	truckerInsurancesCmd.AddCommand(newTruckerInsurancesListCmd())
}

func initTruckerInsurancesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerInsurancesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerInsurancesListOptions(cmd)
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
	query.Set("include", "trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-insurances", query)
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

	rows := buildTruckerInsuranceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerInsurancesTable(cmd, rows)
}

func parseTruckerInsurancesListOptions(cmd *cobra.Command) (truckerInsurancesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerInsurancesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Trucker: trucker,
	}, nil
}

func buildTruckerInsuranceRows(resp jsonAPIResponse) []truckerInsuranceRow {
	rows := make([]truckerInsuranceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := truckerInsuranceRow{
			ID:                   resource.ID,
			CompanyName:          stringAttr(resource.Attributes, "company-name"),
			ContactName:          stringAttr(resource.Attributes, "contact-name"),
			PhoneNumber:          stringAttr(resource.Attributes, "phone-number"),
			PhoneNumberFormatted: stringAttr(resource.Attributes, "phone-number-formatted"),
		}

		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTruckerInsurancesTable(cmd *cobra.Command, rows []truckerInsuranceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker insurances found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOMPANY\tCONTACT\tPHONE\tTRUCKER")
	for _, row := range rows {
		phone := row.PhoneNumberFormatted
		if phone == "" {
			phone = row.PhoneNumber
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.CompanyName, 25),
			truncateString(row.ContactName, 20),
			phone,
			row.TruckerID,
		)
	}
	return writer.Flush()
}
