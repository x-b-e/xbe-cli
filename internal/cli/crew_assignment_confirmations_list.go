package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type crewAssignmentConfirmationsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	CrewRequirement string
	ResourceType    string
	ResourceID      string
	ConfirmedBy     string
	ConfirmedAtMin  string
	ConfirmedAtMax  string
	CreatedAtMin    string
	CreatedAtMax    string
	UpdatedAtMin    string
	UpdatedAtMax    string
}

type crewAssignmentConfirmationRow struct {
	ID                         string `json:"id"`
	AssignmentConfirmationUUID string `json:"assignment_confirmation_uuid,omitempty"`
	CrewRequirementID          string `json:"crew_requirement_id,omitempty"`
	ResourceType               string `json:"resource_type,omitempty"`
	ResourceID                 string `json:"resource_id,omitempty"`
	ConfirmedByID              string `json:"confirmed_by_id,omitempty"`
	StartAt                    string `json:"start_at,omitempty"`
	ConfirmedAt                string `json:"confirmed_at,omitempty"`
	Note                       string `json:"note,omitempty"`
	IsExplicit                 bool   `json:"is_explicit"`
}

func newCrewAssignmentConfirmationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List crew assignment confirmations",
		Long: `List crew assignment confirmations.

Output Columns:
  ID            Confirmation identifier
  UUID          Assignment confirmation UUID
  CREW REQ      Crew requirement ID
  RESOURCE      Resource type and ID
  START AT      Assignment start time
  CONFIRMED AT  Confirmation timestamp
  EXPLICIT      Whether confirmation was explicit
  CONFIRMED BY  User who confirmed

Filters:
  --crew-requirement   Filter by crew requirement ID
  --resource-type      Filter by resource type (class name like Laborer or JSON API type like laborers)
  --resource-id        Filter by resource ID (used with --resource-type)
  --confirmed-by       Filter by confirmed-by user ID
  --confirmed-at-min   Filter by confirmed-at on/after (ISO 8601)
  --confirmed-at-max   Filter by confirmed-at on/before (ISO 8601)
  --created-at-min     Filter by created-at on/after (ISO 8601)
  --created-at-max     Filter by created-at on/before (ISO 8601)
  --updated-at-min     Filter by updated-at on/after (ISO 8601)
  --updated-at-max     Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List confirmations
  xbe view crew-assignment-confirmations list

  # Filter by crew requirement
  xbe view crew-assignment-confirmations list --crew-requirement 123

  # Filter by resource
  xbe view crew-assignment-confirmations list --resource-type laborers --resource-id 456

  # Filter by confirmed date
  xbe view crew-assignment-confirmations list --confirmed-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view crew-assignment-confirmations list --json`,
		Args: cobra.NoArgs,
		RunE: runCrewAssignmentConfirmationsList,
	}
	initCrewAssignmentConfirmationsListFlags(cmd)
	return cmd
}

func init() {
	crewAssignmentConfirmationsCmd.AddCommand(newCrewAssignmentConfirmationsListCmd())
}

func initCrewAssignmentConfirmationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("crew-requirement", "", "Filter by crew requirement ID")
	cmd.Flags().String("resource-type", "", "Filter by resource type (class name like Laborer or JSON API type like laborers)")
	cmd.Flags().String("resource-id", "", "Filter by resource ID (used with --resource-type)")
	cmd.Flags().String("confirmed-by", "", "Filter by confirmed-by user ID")
	cmd.Flags().String("confirmed-at-min", "", "Filter by confirmed-at on/after (ISO 8601)")
	cmd.Flags().String("confirmed-at-max", "", "Filter by confirmed-at on/before (ISO 8601)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewAssignmentConfirmationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCrewAssignmentConfirmationsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[crew-requirement]", opts.CrewRequirement)
	setFilterIfPresent(query, "filter[confirmed-by]", opts.ConfirmedBy)
	setFilterIfPresent(query, "filter[confirmed-at-min]", opts.ConfirmedAtMin)
	setFilterIfPresent(query, "filter[confirmed-at-max]", opts.ConfirmedAtMax)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	if opts.ResourceType != "" && opts.ResourceID != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
		query.Set("filter[resource]", resourceType+"|"+opts.ResourceID)
	} else if opts.ResourceType != "" {
		resourceType := normalizeResourceTypeForFilter(opts.ResourceType)
		query.Set("filter[resource-type]", resourceType)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/crew-assignment-confirmations", query)
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

	rows := buildCrewAssignmentConfirmationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCrewAssignmentConfirmationsTable(cmd, rows)
}

func parseCrewAssignmentConfirmationsListOptions(cmd *cobra.Command) (crewAssignmentConfirmationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	crewRequirement, _ := cmd.Flags().GetString("crew-requirement")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	confirmedBy, _ := cmd.Flags().GetString("confirmed-by")
	confirmedAtMin, _ := cmd.Flags().GetString("confirmed-at-min")
	confirmedAtMax, _ := cmd.Flags().GetString("confirmed-at-max")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewAssignmentConfirmationsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		CrewRequirement: crewRequirement,
		ResourceType:    resourceType,
		ResourceID:      resourceID,
		ConfirmedBy:     confirmedBy,
		ConfirmedAtMin:  confirmedAtMin,
		ConfirmedAtMax:  confirmedAtMax,
		CreatedAtMin:    createdAtMin,
		CreatedAtMax:    createdAtMax,
		UpdatedAtMin:    updatedAtMin,
		UpdatedAtMax:    updatedAtMax,
	}, nil
}

func buildCrewAssignmentConfirmationRows(resp jsonAPIResponse) []crewAssignmentConfirmationRow {
	rows := make([]crewAssignmentConfirmationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := crewAssignmentConfirmationRow{
			ID:                         resource.ID,
			AssignmentConfirmationUUID: stringAttr(attrs, "assignment-confirmation-uuid"),
			StartAt:                    formatDateTime(stringAttr(attrs, "start-at")),
			ConfirmedAt:                formatDateTime(stringAttr(attrs, "confirmed-at")),
			Note:                       stringAttr(attrs, "note"),
			IsExplicit:                 boolAttr(attrs, "is-explicit"),
		}

		if rel, ok := resource.Relationships["crew-requirement"]; ok && rel.Data != nil {
			row.CrewRequirementID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
			row.ResourceType = rel.Data.Type
			row.ResourceID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["confirmed-by"]; ok && rel.Data != nil {
			row.ConfirmedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildCrewAssignmentConfirmationRowFromSingle(resp jsonAPISingleResponse) crewAssignmentConfirmationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := crewAssignmentConfirmationRow{
		ID:                         resource.ID,
		AssignmentConfirmationUUID: stringAttr(attrs, "assignment-confirmation-uuid"),
		StartAt:                    formatDateTime(stringAttr(attrs, "start-at")),
		ConfirmedAt:                formatDateTime(stringAttr(attrs, "confirmed-at")),
		Note:                       stringAttr(attrs, "note"),
		IsExplicit:                 boolAttr(attrs, "is-explicit"),
	}

	if rel, ok := resource.Relationships["crew-requirement"]; ok && rel.Data != nil {
		row.CrewRequirementID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["resource"]; ok && rel.Data != nil {
		row.ResourceType = rel.Data.Type
		row.ResourceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["confirmed-by"]; ok && rel.Data != nil {
		row.ConfirmedByID = rel.Data.ID
	}

	return row
}

func renderCrewAssignmentConfirmationsTable(cmd *cobra.Command, rows []crewAssignmentConfirmationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No crew assignment confirmations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUUID\tCREW REQ\tRESOURCE\tSTART AT\tCONFIRMED AT\tEXPLICIT\tCONFIRMED BY")
	for _, row := range rows {
		resource := ""
		if row.ResourceType != "" && row.ResourceID != "" {
			resource = row.ResourceType + "/" + row.ResourceID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%t\t%s\n",
			row.ID,
			truncateString(row.AssignmentConfirmationUUID, 24),
			row.CrewRequirementID,
			truncateString(resource, 32),
			row.StartAt,
			row.ConfirmedAt,
			row.IsExplicit,
			row.ConfirmedByID,
		)
	}
	return writer.Flush()
}

func normalizeResourceTypeForFilter(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, r := range value {
		if unicode.IsUpper(r) {
			return value
		}
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == len(parts)-1 && strings.HasSuffix(part, "s") && len(part) > 1 {
			part = strings.TrimSuffix(part, "s")
		}
		part = strings.ToLower(part)
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}
