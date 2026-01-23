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

type equipmentMovementStopRequirementsListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	Stop        string
	Requirement string
	Kind        string
}

type equipmentMovementStopRequirementRow struct {
	ID            string `json:"id"`
	StopID        string `json:"stop_id,omitempty"`
	RequirementID string `json:"requirement_id,omitempty"`
	Kind          string `json:"kind,omitempty"`
	RequirementAt string `json:"requirement_at,omitempty"`
}

func newEquipmentMovementStopRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement stop requirements",
		Long: `List equipment movement stop requirements with filtering and pagination.

These records link equipment movement stops to movement requirements, indicating
whether the stop satisfies the origin or destination requirement.

Output Columns:
  ID               Stop requirement identifier
  STOP ID          Equipment movement stop ID
  REQUIREMENT ID   Equipment movement requirement ID
  KIND             Requirement kind (origin/destination)
  REQUIREMENT AT   Requirement timestamp

Filters:
  --stop          Filter by equipment movement stop ID
  --requirement   Filter by equipment movement requirement ID
  --kind          Filter by kind (origin/destination)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List stop requirements
  xbe view equipment-movement-stop-requirements list

  # Filter by stop
  xbe view equipment-movement-stop-requirements list --stop 123

  # Filter by requirement
  xbe view equipment-movement-stop-requirements list --requirement 456

  # Filter by kind
  xbe view equipment-movement-stop-requirements list --kind destination

  # Output as JSON
  xbe view equipment-movement-stop-requirements list --json`,
		RunE: runEquipmentMovementStopRequirementsList,
	}
	initEquipmentMovementStopRequirementsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopRequirementsCmd.AddCommand(newEquipmentMovementStopRequirementsListCmd())
}

func initEquipmentMovementStopRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("stop", "", "Filter by equipment movement stop ID")
	cmd.Flags().String("requirement", "", "Filter by equipment movement requirement ID")
	cmd.Flags().String("kind", "", "Filter by requirement kind (origin/destination)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementStopRequirementsListOptions(cmd)
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
	query.Set("fields[equipment-movement-stop-requirements]", "kind,requirement-at,stop,requirement")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[stop]", opts.Stop)
	setFilterIfPresent(query, "filter[requirement]", opts.Requirement)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stop-requirements", query)
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

	rows := buildEquipmentMovementStopRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementStopRequirementsTable(cmd, rows)
}

func parseEquipmentMovementStopRequirementsListOptions(cmd *cobra.Command) (equipmentMovementStopRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	stop, _ := cmd.Flags().GetString("stop")
	requirement, _ := cmd.Flags().GetString("requirement")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopRequirementsListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		Stop:        stop,
		Requirement: requirement,
		Kind:        kind,
	}, nil
}

func buildEquipmentMovementStopRequirementRows(resp jsonAPIResponse) []equipmentMovementStopRequirementRow {
	rows := make([]equipmentMovementStopRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentMovementStopRequirementRow{
			ID:            resource.ID,
			Kind:          stringAttr(resource.Attributes, "kind"),
			RequirementAt: formatDateTime(stringAttr(resource.Attributes, "requirement-at")),
		}

		if rel, ok := resource.Relationships["stop"]; ok && rel.Data != nil {
			row.StopID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["requirement"]; ok && rel.Data != nil {
			row.RequirementID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentMovementStopRequirementsTable(cmd *cobra.Command, rows []equipmentMovementStopRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement stop requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTOP ID\tREQUIREMENT ID\tKIND\tREQUIREMENT AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.StopID,
			row.RequirementID,
			row.Kind,
			row.RequirementAt,
		)
	}
	return writer.Flush()
}
