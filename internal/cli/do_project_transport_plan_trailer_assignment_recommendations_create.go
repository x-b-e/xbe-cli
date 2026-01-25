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

type doProjectTransportPlanTrailerAssignmentRecommendationsCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ProjectTransportPlanTrailer string
}

func newDoProjectTransportPlanTrailerAssignmentRecommendationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate trailer assignment recommendations",
		Long: `Generate trailer assignment recommendations for a project transport plan trailer.

Required flags:
  --project-transport-plan-trailer  Project transport plan trailer ID`,
		Example: `  # Generate recommendations for a project transport plan trailer
  xbe do project-transport-plan-trailer-assignment-recommendations create --project-transport-plan-trailer 123

  # JSON output
  xbe do project-transport-plan-trailer-assignment-recommendations create --project-transport-plan-trailer 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanTrailerAssignmentRecommendationsCreate,
	}
	initDoProjectTransportPlanTrailerAssignmentRecommendationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanTrailerAssignmentRecommendationsCmd.AddCommand(newDoProjectTransportPlanTrailerAssignmentRecommendationsCreateCmd())
}

func initDoProjectTransportPlanTrailerAssignmentRecommendationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-trailer", "", "Project transport plan trailer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanTrailerAssignmentRecommendationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanTrailerAssignmentRecommendationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlanTrailer) == "" {
		err := fmt.Errorf("--project-transport-plan-trailer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-plan-trailer": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-trailers",
				"id":   opts.ProjectTransportPlanTrailer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-trailer-assignment-recommendations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-trailer-assignment-recommendations", jsonBody)
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

	details := buildProjectTransportPlanTrailerAssignmentRecommendationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanTrailerAssignmentRecommendationDetails(cmd, details)
}

func parseDoProjectTransportPlanTrailerAssignmentRecommendationsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanTrailerAssignmentRecommendationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanTrailer, _ := cmd.Flags().GetString("project-transport-plan-trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanTrailerAssignmentRecommendationsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ProjectTransportPlanTrailer: projectTransportPlanTrailer,
	}, nil
}
