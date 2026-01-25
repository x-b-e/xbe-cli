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

type incidentParticipantsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Incident     string
	User         string
	MobileNumber string
	EmailAddress string
}

type incidentParticipantRow struct {
	ID                    string `json:"id"`
	Name                  string `json:"name,omitempty"`
	EmailAddress          string `json:"email_address,omitempty"`
	MobileNumber          string `json:"mobile_number,omitempty"`
	MobileNumberFormatted string `json:"mobile_number_formatted,omitempty"`
	Involvement           string `json:"involvement,omitempty"`
	IncidentID            string `json:"incident_id,omitempty"`
	UserID                string `json:"user_id,omitempty"`
}

func newIncidentParticipantsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident participants",
		Long: `List incident participants with filtering and pagination.

Output Columns:
  ID           Participant identifier
  NAME         Participant name
  EMAIL        Email address
  MOBILE       Mobile number (formatted when available)
  INVOLVEMENT  Involvement summary
  INCIDENT     Incident ID
  USER         Matched user ID

Filters:
  --incident      Filter by incident ID (comma-separated for multiple)
  --user          Filter by user ID (comma-separated for multiple)
  --mobile-number Filter by mobile number (exact match)
  --email-address Filter by email address (exact match)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident participants
  xbe view incident-participants list

  # Filter by incident
  xbe view incident-participants list --incident 123

  # Filter by email address
  xbe view incident-participants list --email-address "user@example.com"

  # Output as JSON
  xbe view incident-participants list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentParticipantsList,
	}
	initIncidentParticipantsListFlags(cmd)
	return cmd
}

func init() {
	incidentParticipantsCmd.AddCommand(newIncidentParticipantsListCmd())
}

func initIncidentParticipantsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("incident", "", "Filter by incident ID (comma-separated for multiple)")
	cmd.Flags().String("user", "", "Filter by user ID (comma-separated for multiple)")
	cmd.Flags().String("mobile-number", "", "Filter by mobile number (exact match)")
	cmd.Flags().String("email-address", "", "Filter by email address (exact match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentParticipantsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentParticipantsListOptions(cmd)
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
	query.Set("fields[incident-participants]", "name,email-address,mobile-number,mobile-number-formatted,involvement,incident,user")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[incident]", opts.Incident)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[mobile-number]", opts.MobileNumber)
	setFilterIfPresent(query, "filter[email-address]", opts.EmailAddress)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-participants", query)
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

	rows := buildIncidentParticipantRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentParticipantsTable(cmd, rows)
}

func parseIncidentParticipantsListOptions(cmd *cobra.Command) (incidentParticipantsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	incident, _ := cmd.Flags().GetString("incident")
	user, _ := cmd.Flags().GetString("user")
	mobileNumber, _ := cmd.Flags().GetString("mobile-number")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentParticipantsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Incident:     incident,
		User:         user,
		MobileNumber: mobileNumber,
		EmailAddress: emailAddress,
	}, nil
}

func buildIncidentParticipantRows(resp jsonAPIResponse) []incidentParticipantRow {
	rows := make([]incidentParticipantRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := incidentParticipantRow{
			ID:                    resource.ID,
			Name:                  stringAttr(attrs, "name"),
			EmailAddress:          stringAttr(attrs, "email-address"),
			MobileNumber:          stringAttr(attrs, "mobile-number"),
			MobileNumberFormatted: stringAttr(attrs, "mobile-number-formatted"),
			Involvement:           stringAttr(attrs, "involvement"),
			IncidentID:            relationshipIDFromMap(resource.Relationships, "incident"),
			UserID:                relationshipIDFromMap(resource.Relationships, "user"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderIncidentParticipantsTable(cmd *cobra.Command, rows []incidentParticipantRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident participants found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tEMAIL\tMOBILE\tINVOLVEMENT\tINCIDENT\tUSER")
	for _, row := range rows {
		mobile := row.MobileNumberFormatted
		if mobile == "" {
			mobile = row.MobileNumber
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.EmailAddress, 25),
			truncateString(mobile, 18),
			truncateString(row.Involvement, 20),
			row.IncidentID,
			row.UserID,
		)
	}
	return writer.Flush()
}
