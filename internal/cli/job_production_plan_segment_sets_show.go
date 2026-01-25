package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanSegmentSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSegmentSetDetails struct {
	ID                          string   `json:"id"`
	JobProductionPlanID         string   `json:"job_production_plan_id,omitempty"`
	Name                        string   `json:"name,omitempty"`
	IsDefault                   bool     `json:"is_default"`
	StartOffsetMinutes          int      `json:"start_offset_minutes,omitempty"`
	JobProductionPlanSegmentIDs []string `json:"job_production_plan_segment_ids,omitempty"`
}

func newJobProductionPlanSegmentSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan segment set details",
		Long: `Show the full details of a job production plan segment set.

Output Fields:
  ID                     Segment set identifier
  Job Production Plan     Associated job production plan ID
  Name                   Segment set name
  Is Default             Whether the set is default
  Start Offset Minutes    Start offset in minutes
  Segments               Job production plan segment IDs

Arguments:
  <id>    The segment set ID (required). You can find IDs using the list command.`,
		Example: `  # Show a job production plan segment set
  xbe view job-production-plan-segment-sets show 123

  # Get JSON output
  xbe view job-production-plan-segment-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSegmentSetsShow,
	}
	initJobProductionPlanSegmentSetsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSegmentSetsCmd.AddCommand(newJobProductionPlanSegmentSetsShowCmd())
}

func initJobProductionPlanSegmentSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSegmentSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSegmentSetsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan segment set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "job-production-plan,job-production-plan-segments")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-segment-sets/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanSegmentSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSegmentSetDetails(cmd, details)
}

func parseJobProductionPlanSegmentSetsShowOptions(cmd *cobra.Command) (jobProductionPlanSegmentSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSegmentSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSegmentSetDetails(resp jsonAPISingleResponse) jobProductionPlanSegmentSetDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanSegmentSetDetails{
		ID:                 resource.ID,
		Name:               stringAttr(attrs, "name"),
		IsDefault:          boolAttr(attrs, "is-default"),
		StartOffsetMinutes: intAttr(attrs, "start-offset-minutes"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["job-production-plan-segments"]; ok {
		details.JobProductionPlanSegmentIDs = segmentSetRelationshipIDs(rel)
	}

	return details
}

func segmentSetRelationshipIDs(rel jsonAPIRelationship) []string {
	identifiers := relationshipIDs(rel)
	if len(identifiers) == 0 {
		return nil
	}
	ids := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		if identifier.ID != "" {
			ids = append(ids, identifier.ID)
		}
	}
	return ids
}

func renderJobProductionPlanSegmentSetDetails(cmd *cobra.Command, details jobProductionPlanSegmentSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	fmt.Fprintf(out, "Is Default: %t\n", details.IsDefault)
	fmt.Fprintf(out, "Start Offset Minutes: %d\n", details.StartOffsetMinutes)

	if len(details.JobProductionPlanSegmentIDs) > 0 {
		fmt.Fprintf(out, "Segments: %s\n", strings.Join(details.JobProductionPlanSegmentIDs, ", "))
	}

	return nil
}
