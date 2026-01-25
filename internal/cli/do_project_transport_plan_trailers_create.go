package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doProjectTransportPlanTrailersCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ProjectTransportPlan string
	SegmentStart         string
	SegmentEnd           string
	Trailer              string
	Status               string
}

func newDoProjectTransportPlanTrailersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan trailer assignment",
		Long: `Create a project transport plan trailer assignment.

Required:
  --project-transport-plan  Project transport plan ID
  --segment-start           Segment start ID
  --segment-end             Segment end ID

Optional:
  --trailer                 Trailer ID
  --status                  Assignment status (editing, active). Defaults to editing unless provided.

Notes:
  - Status "active" requires a trailer assignment.`,
		Example: `  # Create a trailer assignment (defaults to editing)
  xbe do project-transport-plan-trailers create \
    --project-transport-plan 123 \
    --segment-start 456 \
    --segment-end 789

  # Create an active trailer assignment
  xbe do project-transport-plan-trailers create \
    --project-transport-plan 123 \
    --segment-start 456 \
    --segment-end 789 \
    --trailer 555 \
    --status active`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanTrailersCreate,
	}
	initDoProjectTransportPlanTrailersCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanTrailersCmd.AddCommand(newDoProjectTransportPlanTrailersCreateCmd())
}

func initDoProjectTransportPlanTrailersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID (required)")
	cmd.Flags().String("segment-start", "", "Segment start ID (required)")
	cmd.Flags().String("segment-end", "", "Segment end ID (required)")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("status", "", "Assignment status (editing, active)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanTrailersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanTrailersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.ProjectTransportPlan) == "" {
		err := fmt.Errorf("--project-transport-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentStart) == "" {
		err := fmt.Errorf("--segment-start is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentEnd) == "" {
		err := fmt.Errorf("--segment-end is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		},
		"segment-start": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentStart,
			},
		},
		"segment-end": map[string]any{
			"data": map[string]string{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentEnd,
			},
		},
	}
	if opts.Trailer != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]string{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		}
	}

	data := map[string]any{
		"type":          "project-transport-plan-trailers",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-trailers", jsonBody)
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

	details := buildProjectTransportPlanTrailerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan trailer %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanTrailersCreateOptions(cmd *cobra.Command) (doProjectTransportPlanTrailersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	trailer, _ := cmd.Flags().GetString("trailer")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanTrailersCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ProjectTransportPlan: projectTransportPlan,
		SegmentStart:         segmentStart,
		SegmentEnd:           segmentEnd,
		Trailer:              trailer,
		Status:               status,
	}, nil
}
