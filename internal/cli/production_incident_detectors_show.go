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

type productionIncidentDetectorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type productionIncidentDetectorDetails struct {
	ID                string `json:"id"`
	JobProductionPlan string `json:"job_production_plan_id,omitempty"`
	LookaheadOffset   int    `json:"lookahead_offset,omitempty"`
	MinutesThreshold  int    `json:"minutes_threshold,omitempty"`
	QuantityThreshold int    `json:"quantity_threshold,omitempty"`
	Incidents         any    `json:"incidents,omitempty"`
}

func newProductionIncidentDetectorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show production incident detector details",
		Long: `Show the full details of a production incident detector run.

Output Fields:
  ID            Detector run identifier
  Job Plan      Job production plan ID
  Lookahead     Lookahead offset (minutes)
  Minutes       Minutes threshold (minutes)
  Quantity      Quantity threshold (units)
  Incidents     Detected incident details

Arguments:
  <id>  The detector run ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show detector run details
  xbe view production-incident-detectors show 123

  # Output as JSON
  xbe view production-incident-detectors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProductionIncidentDetectorsShow,
	}
	initProductionIncidentDetectorsShowFlags(cmd)
	return cmd
}

func init() {
	productionIncidentDetectorsCmd.AddCommand(newProductionIncidentDetectorsShowCmd())
}

func initProductionIncidentDetectorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionIncidentDetectorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProductionIncidentDetectorsShowOptions(cmd)
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
		return fmt.Errorf("production incident detector id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[production-incident-detectors]", "job-production-plan,lookahead-offset,minutes-threshold,quantity-threshold,incidents")

	body, _, err := client.Get(cmd.Context(), "/v1/production-incident-detectors/"+id, query)
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

	details := buildProductionIncidentDetectorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProductionIncidentDetectorDetails(cmd, details)
}

func parseProductionIncidentDetectorsShowOptions(cmd *cobra.Command) (productionIncidentDetectorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionIncidentDetectorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProductionIncidentDetectorDetails(resp jsonAPISingleResponse) productionIncidentDetectorDetails {
	attrs := resp.Data.Attributes
	details := productionIncidentDetectorDetails{
		ID:                resp.Data.ID,
		LookaheadOffset:   intAttr(attrs, "lookahead-offset"),
		MinutesThreshold:  intAttr(attrs, "minutes-threshold"),
		QuantityThreshold: intAttr(attrs, "quantity-threshold"),
		Incidents:         anyAttr(attrs, "incidents"),
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlan = rel.Data.ID
	}

	return details
}

func renderProductionIncidentDetectorDetails(cmd *cobra.Command, details productionIncidentDetectorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlan)
	}
	if details.LookaheadOffset != 0 {
		fmt.Fprintf(out, "Lookahead Offset (minutes): %d\n", details.LookaheadOffset)
	}
	if details.MinutesThreshold != 0 {
		fmt.Fprintf(out, "Minutes Threshold (minutes): %d\n", details.MinutesThreshold)
	}
	if details.QuantityThreshold != 0 {
		fmt.Fprintf(out, "Quantity Threshold (units): %d\n", details.QuantityThreshold)
	}
	if details.Incidents != nil {
		count := countConstraintItems(details.Incidents)
		fmt.Fprintf(out, "Incidents: %d\n", count)
		if formatted := formatAnyJSON(details.Incidents); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Incident Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
