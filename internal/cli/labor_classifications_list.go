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

type laborClassificationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Name         string
	Abbreviation string
}

func newLaborClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labor classifications",
		Long: `List labor classifications with filtering and pagination.

Labor classifications define worker types with their capabilities and permissions.

Output Columns:
  ID             Labor classification identifier
  NAME           Classification name (e.g., Raker, Foreman)
  ABBREVIATION   Short code
  MOBILIZATION   Mobilization method
  APPROVER       Can approve time cards
  MANAGER        Is a manager role
  PROJECTS       Can manage projects

Filters:
  --name          Filter by name (partial match, case-insensitive)
  --abbreviation  Filter by abbreviation (partial match, case-insensitive)`,
		Example: `  # List all labor classifications
  xbe view labor-classifications list

  # Filter by name
  xbe view labor-classifications list --name "foreman"

  # Filter by abbreviation
  xbe view labor-classifications list --abbreviation "fmn"

  # Output as JSON
  xbe view labor-classifications list --json`,
		RunE: runLaborClassificationsList,
	}
	initLaborClassificationsListFlags(cmd)
	return cmd
}

func init() {
	laborClassificationsCmd.AddCommand(newLaborClassificationsListCmd())
}

func initLaborClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("abbreviation", "", "Filter by abbreviation (partial match)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLaborClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLaborClassificationsListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[labor-classifications]", "name,abbreviation,mobilization-method,is-time-card-approver,is-manager,can-manage-projects")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[abbreviation]", opts.Abbreviation)

	body, _, err := client.Get(cmd.Context(), "/v1/labor-classifications", query)
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

	rows := buildLaborClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLaborClassificationsTable(cmd, rows)
}

func parseLaborClassificationsListOptions(cmd *cobra.Command) (laborClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return laborClassificationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Name:         name,
		Abbreviation: abbreviation,
	}, nil
}

func buildLaborClassificationRows(resp jsonAPIResponse) []laborClassificationRow {
	rows := make([]laborClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := laborClassificationRow{
			ID:                 resource.ID,
			Name:               stringAttr(resource.Attributes, "name"),
			Abbreviation:       stringAttr(resource.Attributes, "abbreviation"),
			MobilizationMethod: stringAttr(resource.Attributes, "mobilization-method"),
			IsTimeCardApprover: boolAttr(resource.Attributes, "is-time-card-approver"),
			IsManager:          boolAttr(resource.Attributes, "is-manager"),
			CanManageProjects:  boolAttr(resource.Attributes, "can-manage-projects"),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderLaborClassificationsTable(cmd *cobra.Command, rows []laborClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No labor classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tABBREVIATION\tMOBILIZATION\tAPPROVER\tMANAGER\tPROJECTS")
	for _, row := range rows {
		approver := "no"
		if row.IsTimeCardApprover {
			approver = "yes"
		}
		manager := "no"
		if row.IsManager {
			manager = "yes"
		}
		projects := "no"
		if row.CanManageProjects {
			projects = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 25),
			truncateString(row.Abbreviation, 10),
			truncateString(row.MobilizationMethod, 12),
			approver,
			manager,
			projects,
		)
	}
	return writer.Flush()
}
