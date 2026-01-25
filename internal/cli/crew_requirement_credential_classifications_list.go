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

type crewRequirementCredentialClassificationsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	CrewRequirement string
}

type crewRequirementCredentialClassificationRow struct {
	ID                           string `json:"id"`
	CrewRequirementType          string `json:"crew_requirement_type,omitempty"`
	CrewRequirementID            string `json:"crew_requirement_id,omitempty"`
	CredentialClassificationType string `json:"credential_classification_type,omitempty"`
	CredentialClassificationID   string `json:"credential_classification_id,omitempty"`
}

func newCrewRequirementCredentialClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List crew requirement credential classifications",
		Long: `List crew requirement credential classifications with filtering and pagination.

These records connect crew requirements with the credential classifications they require.

Output Columns:
  ID                 Link identifier
  CREW REQ TYPE       Crew requirement type
  CREW REQ ID         Crew requirement ID
  CRED TYPE           Credential classification type
  CRED ID             Credential classification ID

Filters:
  --crew-requirement  Filter by crew requirement ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List all crew requirement credential classifications
  xbe view crew-requirement-credential-classifications list

  # Filter by crew requirement
  xbe view crew-requirement-credential-classifications list --crew-requirement 123

  # Output as JSON
  xbe view crew-requirement-credential-classifications list --json`,
		RunE: runCrewRequirementCredentialClassificationsList,
	}
	initCrewRequirementCredentialClassificationsListFlags(cmd)
	return cmd
}

func init() {
	crewRequirementCredentialClassificationsCmd.AddCommand(newCrewRequirementCredentialClassificationsListCmd())
}

func initCrewRequirementCredentialClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("crew-requirement", "", "Filter by crew requirement ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewRequirementCredentialClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCrewRequirementCredentialClassificationsListOptions(cmd)
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
	query.Set("fields[crew-requirement-credential-classifications]", "crew-requirement,credential-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[crew_requirement]", opts.CrewRequirement)

	body, _, err := client.Get(cmd.Context(), "/v1/crew-requirement-credential-classifications", query)
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

	rows := buildCrewRequirementCredentialClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCrewRequirementCredentialClassificationsTable(cmd, rows)
}

func parseCrewRequirementCredentialClassificationsListOptions(cmd *cobra.Command) (crewRequirementCredentialClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	crewRequirement, _ := cmd.Flags().GetString("crew-requirement")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewRequirementCredentialClassificationsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		CrewRequirement: crewRequirement,
	}, nil
}

func buildCrewRequirementCredentialClassificationRows(resp jsonAPIResponse) []crewRequirementCredentialClassificationRow {
	rows := make([]crewRequirementCredentialClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := crewRequirementCredentialClassificationRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["crew-requirement"]; ok && rel.Data != nil {
			row.CrewRequirementType = rel.Data.Type
			row.CrewRequirementID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["credential-classification"]; ok && rel.Data != nil {
			row.CredentialClassificationType = rel.Data.Type
			row.CredentialClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCrewRequirementCredentialClassificationsTable(cmd *cobra.Command, rows []crewRequirementCredentialClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No crew requirement credential classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCREW REQ TYPE\tCREW REQ ID\tCRED TYPE\tCRED ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.CrewRequirementType,
			row.CrewRequirementID,
			row.CredentialClassificationType,
			row.CredentialClassificationID,
		)
	}
	return writer.Flush()
}
