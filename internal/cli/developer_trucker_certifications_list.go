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

type developerTruckerCertificationsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Developer              string
	Trucker                string
	Classification         string
	TenderJobScheduleShift string
}

type developerTruckerCertificationRow struct {
	ID                 string `json:"id"`
	StartOn            string `json:"start_on,omitempty"`
	EndOn              string `json:"end_on,omitempty"`
	DefaultMultiplier  string `json:"default_multiplier,omitempty"`
	DeveloperID        string `json:"developer_id,omitempty"`
	DeveloperName      string `json:"developer_name,omitempty"`
	TruckerID          string `json:"trucker_id,omitempty"`
	TruckerName        string `json:"trucker_name,omitempty"`
	ClassificationID   string `json:"classification_id,omitempty"`
	ClassificationName string `json:"classification_name,omitempty"`
}

func newDeveloperTruckerCertificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List developer trucker certifications",
		Long: `List developer trucker certifications with filtering and pagination.

Output Columns:
  ID              Certification identifier
  DEVELOPER       Developer name or ID
  TRUCKER         Trucker name or ID
  CLASSIFICATION  Certification classification name or ID
  START           Start date
  END             End date
  DEFAULT         Default multiplier

Filters:
  --developer                 Filter by developer ID
  --trucker                   Filter by trucker ID
  --classification            Filter by classification ID
  --tender-job-schedule-shift Filter by tender job schedule shift ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List developer trucker certifications
  xbe view developer-trucker-certifications list

  # Filter by developer
  xbe view developer-trucker-certifications list --developer 123

  # Filter by trucker
  xbe view developer-trucker-certifications list --trucker 456

  # Filter by classification
  xbe view developer-trucker-certifications list --classification 789

  # Filter by tender job schedule shift
  xbe view developer-trucker-certifications list --tender-job-schedule-shift 321

  # Output as JSON
  xbe view developer-trucker-certifications list --json`,
		RunE: runDeveloperTruckerCertificationsList,
	}
	initDeveloperTruckerCertificationsListFlags(cmd)
	return cmd
}

func init() {
	developerTruckerCertificationsCmd.AddCommand(newDeveloperTruckerCertificationsListCmd())
}

func initDeveloperTruckerCertificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("developer", "", "Filter by developer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("classification", "", "Filter by classification ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeveloperTruckerCertificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeveloperTruckerCertificationsListOptions(cmd)
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
	query.Set("fields[developer-trucker-certifications]", "start-on,end-on,default-multiplier,developer,trucker,classification")
	query.Set("include", "developer,trucker,classification")
	query.Set("fields[developers]", "name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developer-trucker-certification-classifications]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[classification]", opts.Classification)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShift)

	body, _, err := client.Get(cmd.Context(), "/v1/developer-trucker-certifications", query)
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

	rows := buildDeveloperTruckerCertificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeveloperTruckerCertificationsTable(cmd, rows)
}

func parseDeveloperTruckerCertificationsListOptions(cmd *cobra.Command) (developerTruckerCertificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	developer, _ := cmd.Flags().GetString("developer")
	trucker, _ := cmd.Flags().GetString("trucker")
	classification, _ := cmd.Flags().GetString("classification")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return developerTruckerCertificationsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Developer:              developer,
		Trucker:                trucker,
		Classification:         classification,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}

func buildDeveloperTruckerCertificationRows(resp jsonAPIResponse) []developerTruckerCertificationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]developerTruckerCertificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDeveloperTruckerCertificationRow(resource, included))
	}
	return rows
}

func developerTruckerCertificationRowFromSingle(resp jsonAPISingleResponse) developerTruckerCertificationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildDeveloperTruckerCertificationRow(resp.Data, included)
}

func buildDeveloperTruckerCertificationRow(resource jsonAPIResource, included map[string]jsonAPIResource) developerTruckerCertificationRow {
	row := developerTruckerCertificationRow{
		ID:                resource.ID,
		StartOn:           stringAttr(resource.Attributes, "start-on"),
		EndOn:             stringAttr(resource.Attributes, "end-on"),
		DefaultMultiplier: stringAttr(resource.Attributes, "default-multiplier"),
	}

	if rel, ok := resource.Relationships["developer"]; ok && rel.Data != nil {
		row.DeveloperID = rel.Data.ID
		if developer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.DeveloperName = stringAttr(developer.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["classification"]; ok && rel.Data != nil {
		row.ClassificationID = rel.Data.ID
		if classification, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ClassificationName = stringAttr(classification.Attributes, "name")
		}
	}

	return row
}

func renderDeveloperTruckerCertificationsTable(cmd *cobra.Command, rows []developerTruckerCertificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No developer trucker certifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDEVELOPER\tTRUCKER\tCLASSIFICATION\tSTART\tEND\tDEFAULT")
	for _, row := range rows {
		developer := firstNonEmpty(row.DeveloperName, row.DeveloperID)
		trucker := firstNonEmpty(row.TruckerName, row.TruckerID)
		classification := firstNonEmpty(row.ClassificationName, row.ClassificationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(developer, 28),
			truncateString(trucker, 28),
			truncateString(classification, 28),
			row.StartOn,
			row.EndOn,
			row.DefaultMultiplier,
		)
	}
	return writer.Flush()
}
