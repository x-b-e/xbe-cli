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

type doProjectTransportPlanSegmentTractorsCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ProjectTransportPlanSegment string
	Tractor                     string
}

func newDoProjectTransportPlanSegmentTractorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan segment tractor",
		Long: `Create a project transport plan segment tractor.

Required flags:
  --project-transport-plan-segment  Project transport plan segment ID
  --tractor                         Tractor ID

Notes:
  - The tractor must belong to the same trucker as the segment.
  - Each segment can only have one tractor relationship.`,
		Example: `  # Create a segment tractor assignment
  xbe do project-transport-plan-segment-tractors create --project-transport-plan-segment 123 --tractor 456

  # Get JSON output
  xbe do project-transport-plan-segment-tractors create --project-transport-plan-segment 123 --tractor 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanSegmentTractorsCreate,
	}
	initDoProjectTransportPlanSegmentTractorsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentTractorsCmd.AddCommand(newDoProjectTransportPlanSegmentTractorsCreateCmd())
}

func initDoProjectTransportPlanSegmentTractorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-segment", "", "Project transport plan segment ID (required)")
	cmd.Flags().String("tractor", "", "Tractor ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentTractorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanSegmentTractorsCreateOptions(cmd)
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

	segmentID := strings.TrimSpace(opts.ProjectTransportPlanSegment)
	if segmentID == "" {
		err := fmt.Errorf("--project-transport-plan-segment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	tractorID := strings.TrimSpace(opts.Tractor)
	if tractorID == "" {
		err := fmt.Errorf("--tractor is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-plan-segment": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   segmentID,
			},
		},
		"tractor": map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   tractorID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-segment-tractors",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-segment-tractors", jsonBody)
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

	details := buildProjectTransportPlanSegmentTractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan segment tractor %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentTractorsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanSegmentTractorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	segmentID, _ := cmd.Flags().GetString("project-transport-plan-segment")
	tractorID, _ := cmd.Flags().GetString("tractor")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentTractorsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ProjectTransportPlanSegment: segmentID,
		Tractor:                     tractorID,
	}, nil
}
