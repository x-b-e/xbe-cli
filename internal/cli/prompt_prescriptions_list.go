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

type promptPrescriptionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	EmailAddress string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type promptPrescriptionRow struct {
	ID               string `json:"id"`
	Name             string `json:"name,omitempty"`
	EmailAddress     string `json:"email_address,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	LocationName     string `json:"location_name,omitempty"`
	Role             string `json:"role,omitempty"`
}

func newPromptPrescriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prompt prescriptions",
		Long: `List prompt prescriptions with filtering and pagination.

Output Columns:
  ID            Prompt prescription identifier
  NAME          Contact name
  EMAIL         Contact email address
  ORGANIZATION  Organization name
  LOCATION      Location name
  ROLE          Role or job title

Filters:
  --email-address  Filter by email address
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --is-created-at  Filter by has created-at (true/false)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-updated-at  Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prompt prescriptions
  xbe view prompt-prescriptions list

  # Filter by email address
  xbe view prompt-prescriptions list --email-address "name@example.com"

  # Filter by created-at window
  xbe view prompt-prescriptions list --created-at-min 2025-01-01T00:00:00Z --created-at-max 2025-12-31T23:59:59Z

  # Output as JSON
  xbe view prompt-prescriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runPromptPrescriptionsList,
	}
	initPromptPrescriptionsListFlags(cmd)
	return cmd
}

func init() {
	promptPrescriptionsCmd.AddCommand(newPromptPrescriptionsListCmd())
}

func initPromptPrescriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("email-address", "", "Filter by email address")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPromptPrescriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePromptPrescriptionsListOptions(cmd)
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
	query.Set("fields[prompt-prescriptions]", "name,email-address,organization-name,location-name,role")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[email-address]", opts.EmailAddress)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/prompt-prescriptions", query)
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

	rows := buildPromptPrescriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPromptPrescriptionsTable(cmd, rows)
}

func parsePromptPrescriptionsListOptions(cmd *cobra.Command) (promptPrescriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return promptPrescriptionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		EmailAddress: emailAddress,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildPromptPrescriptionRows(resp jsonAPIResponse) []promptPrescriptionRow {
	rows := make([]promptPrescriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPromptPrescriptionRow(resource))
	}
	return rows
}

func promptPrescriptionRowFromSingle(resp jsonAPISingleResponse) promptPrescriptionRow {
	return buildPromptPrescriptionRow(resp.Data)
}

func buildPromptPrescriptionRow(resource jsonAPIResource) promptPrescriptionRow {
	attrs := resource.Attributes
	return promptPrescriptionRow{
		ID:               resource.ID,
		Name:             strings.TrimSpace(stringAttr(attrs, "name")),
		EmailAddress:     strings.TrimSpace(stringAttr(attrs, "email-address")),
		OrganizationName: strings.TrimSpace(stringAttr(attrs, "organization-name")),
		LocationName:     strings.TrimSpace(stringAttr(attrs, "location-name")),
		Role:             strings.TrimSpace(stringAttr(attrs, "role")),
	}
}

func renderPromptPrescriptionsTable(cmd *cobra.Command, rows []promptPrescriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompt prescriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tEMAIL\tORGANIZATION\tLOCATION\tROLE")
	for _, row := range rows {
		name := row.Name
		if name == "" {
			name = "-"
		}
		email := row.EmailAddress
		if email == "" {
			email = "-"
		}
		organization := row.OrganizationName
		if organization == "" {
			organization = "-"
		}
		location := row.LocationName
		if location == "" {
			location = "-"
		}
		role := row.Role
		if role == "" {
			role = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			name,
			email,
			organization,
			location,
			role,
		)
	}
	return writer.Flush()
}
