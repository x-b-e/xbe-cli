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

type driverMovementObservationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverMovementObservationDetails struct {
	ID                       string `json:"id"`
	PlanID                   string `json:"plan_id,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
	UpdatedAt                string `json:"updated_at,omitempty"`
	CyclesByMaterialSite     any    `json:"cycles_by_material_site,omitempty"`
	SegmentMovementFragments any    `json:"segment_movement_fragments,omitempty"`
	JobSiteMinutes           any    `json:"job_site_minutes,omitempty"`
	DrivingMinutes           any    `json:"driving_minutes,omitempty"`
	MaterialSiteMinutes      any    `json:"material_site_minutes,omitempty"`
}

func newDriverMovementObservationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver movement observation details",
		Long: `Show the full details of a driver movement observation.

Output Fields:
  ID
  Plan ID
  Created At
  Updated At
  Cycles By Material Site
  Segment Movement Fragments
  Job Site Minutes
  Driving Minutes
  Material Site Minutes

Arguments:
  <id>    The observation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a driver movement observation
  xbe view driver-movement-observations show 123

  # Get JSON output
  xbe view driver-movement-observations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverMovementObservationsShow,
	}
	initDriverMovementObservationsShowFlags(cmd)
	return cmd
}

func init() {
	driverMovementObservationsCmd.AddCommand(newDriverMovementObservationsShowCmd())
}

func initDriverMovementObservationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementObservationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseDriverMovementObservationsShowOptions(cmd)
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
		return fmt.Errorf("driver movement observation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[driver-movement-observations]", "created-at,updated-at,plan,cycles-by-material-site,segment-movement-fragments,job-site-minutes,driving-minutes,material-site-minutes")

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-observations/"+id, query)
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

	details := buildDriverMovementObservationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverMovementObservationDetails(cmd, details)
}

func parseDriverMovementObservationsShowOptions(cmd *cobra.Command) (driverMovementObservationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementObservationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverMovementObservationDetails(resp jsonAPISingleResponse) driverMovementObservationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := driverMovementObservationDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["plan"]; ok && rel.Data != nil {
		details.PlanID = rel.Data.ID
	}

	if value, ok := attrs["cycles-by-material-site"]; ok {
		details.CyclesByMaterialSite = value
	}
	if value, ok := attrs["segment-movement-fragments"]; ok {
		details.SegmentMovementFragments = value
	}
	if value, ok := attrs["job-site-minutes"]; ok {
		details.JobSiteMinutes = value
	}
	if value, ok := attrs["driving-minutes"]; ok {
		details.DrivingMinutes = value
	}
	if value, ok := attrs["material-site-minutes"]; ok {
		details.MaterialSiteMinutes = value
	}

	return details
}

func renderDriverMovementObservationDetails(cmd *cobra.Command, details driverMovementObservationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PlanID != "" {
		fmt.Fprintf(out, "Plan ID: %s\n", details.PlanID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.CyclesByMaterialSite != nil {
		fmt.Fprintln(out, "Cycles By Material Site:")
		fmt.Fprintln(out, formatDriverMovementObservationJSON(details.CyclesByMaterialSite))
	}
	if details.SegmentMovementFragments != nil {
		fmt.Fprintln(out, "Segment Movement Fragments:")
		fmt.Fprintln(out, formatDriverMovementObservationJSON(details.SegmentMovementFragments))
	}
	if details.JobSiteMinutes != nil {
		fmt.Fprintln(out, "Job Site Minutes:")
		fmt.Fprintln(out, formatDriverMovementObservationJSON(details.JobSiteMinutes))
	}
	if details.DrivingMinutes != nil {
		fmt.Fprintln(out, "Driving Minutes:")
		fmt.Fprintln(out, formatDriverMovementObservationJSON(details.DrivingMinutes))
	}
	if details.MaterialSiteMinutes != nil {
		fmt.Fprintln(out, "Material Site Minutes:")
		fmt.Fprintln(out, formatDriverMovementObservationJSON(details.MaterialSiteMinutes))
	}

	return nil
}

func formatDriverMovementObservationJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
