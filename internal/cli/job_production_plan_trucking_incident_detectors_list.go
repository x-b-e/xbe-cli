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

type jobProductionPlanTruckingIncidentDetectorsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
	IsPerformed         string
}

type jobProductionPlanTruckingIncidentDetectorRow struct {
	ID                   string `json:"id"`
	JobProductionPlanID  string `json:"job_production_plan_id,omitempty"`
	AsOf                 string `json:"as_of,omitempty"`
	PersistChanges       bool   `json:"persist_changes"`
	IsPerformed          bool   `json:"is_performed"`
	DetectedIncidentsCnt int    `json:"detected_incidents_count,omitempty"`
}

func newJobProductionPlanTruckingIncidentDetectorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan trucking incident detectors",
		Long: `List job production plan trucking incident detectors.

Output Columns:
  ID        Detector identifier
  PLAN      Job production plan ID
  AS OF     As-of timestamp for detection
  PERSIST   Whether detector persists incident changes
  PERFORMED Whether detection has been performed
  DETECTED  Count of detected incidents

Filters:
  --job-production-plan  Filter by job production plan ID
  --is-performed         Filter by performed status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucking incident detectors
  xbe view job-production-plan-trucking-incident-detectors list

  # Filter by job production plan
  xbe view job-production-plan-trucking-incident-detectors list --job-production-plan 123

  # Filter by performed status
  xbe view job-production-plan-trucking-incident-detectors list --is-performed true

  # Output as JSON
  xbe view job-production-plan-trucking-incident-detectors list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanTruckingIncidentDetectorsList,
	}
	initJobProductionPlanTruckingIncidentDetectorsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTruckingIncidentDetectorsCmd.AddCommand(newJobProductionPlanTruckingIncidentDetectorsListCmd())
}

func initJobProductionPlanTruckingIncidentDetectorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("is-performed", "", "Filter by performed status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTruckingIncidentDetectorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanTruckingIncidentDetectorsListOptions(cmd)
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
	query.Set("fields[job-production-plan-trucking-incident-detectors]", "as-of,persist-changes,is-performed,detected-incidents")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlanID)
	setFilterIfPresent(query, "filter[is-performed]", opts.IsPerformed)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-trucking-incident-detectors", query)
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

	rows := buildJobProductionPlanTruckingIncidentDetectorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanTruckingIncidentDetectorsTable(cmd, rows)
}

func parseJobProductionPlanTruckingIncidentDetectorsListOptions(cmd *cobra.Command) (jobProductionPlanTruckingIncidentDetectorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	isPerformed, _ := cmd.Flags().GetString("is-performed")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTruckingIncidentDetectorsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
		IsPerformed:         isPerformed,
	}, nil
}

func buildJobProductionPlanTruckingIncidentDetectorRows(resp jsonAPIResponse) []jobProductionPlanTruckingIncidentDetectorRow {
	rows := make([]jobProductionPlanTruckingIncidentDetectorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanTruckingIncidentDetectorRow{
			ID:                   resource.ID,
			AsOf:                 formatDateTime(stringAttr(attrs, "as-of")),
			PersistChanges:       boolAttr(attrs, "persist-changes"),
			IsPerformed:          boolAttr(attrs, "is-performed"),
			DetectedIncidentsCnt: sliceLenAttr(attrs, "detected-incidents"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanTruckingIncidentDetectorsTable(cmd *cobra.Command, rows []jobProductionPlanTruckingIncidentDetectorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan trucking incident detectors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tAS OF\tPERSIST\tPERFORMED\tDETECTED")
	for _, row := range rows {
		persist := ""
		performed := ""
		detected := ""
		if row.PersistChanges {
			persist = "yes"
		}
		if row.IsPerformed {
			performed = "yes"
		}
		if row.DetectedIncidentsCnt > 0 {
			detected = strconv.Itoa(row.DetectedIncidentsCnt)
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.AsOf,
			persist,
			performed,
			detected,
		)
	}
	return writer.Flush()
}

func sliceLenAttr(attrs map[string]any, key string) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}
