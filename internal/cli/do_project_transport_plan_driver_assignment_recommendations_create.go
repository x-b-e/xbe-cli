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

type doProjectTransportPlanDriverAssignmentRecommendationsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ProjectTransportPlanDriver string
}

func newDoProjectTransportPlanDriverAssignmentRecommendationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate driver assignment recommendations",
		Long: `Generate driver assignment recommendations for a project transport plan driver.

Required flags:
  --project-transport-plan-driver  Project transport plan driver ID`,
		Example: `  # Generate recommendations for a project transport plan driver
  xbe do project-transport-plan-driver-assignment-recommendations create --project-transport-plan-driver 123

  # JSON output
  xbe do project-transport-plan-driver-assignment-recommendations create --project-transport-plan-driver 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanDriverAssignmentRecommendationsCreate,
	}
	initDoProjectTransportPlanDriverAssignmentRecommendationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanDriverAssignmentRecommendationsCmd.AddCommand(newDoProjectTransportPlanDriverAssignmentRecommendationsCreateCmd())
}

func initDoProjectTransportPlanDriverAssignmentRecommendationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-driver", "", "Project transport plan driver ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanDriverAssignmentRecommendationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanDriverAssignmentRecommendationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlanDriver) == "" {
		err := fmt.Errorf("--project-transport-plan-driver is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-plan-driver": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-drivers",
				"id":   opts.ProjectTransportPlanDriver,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-driver-assignment-recommendations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-driver-assignment-recommendations", jsonBody)
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

	details := buildProjectTransportPlanDriverAssignmentRecommendationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanDriverAssignmentRecommendationDetails(cmd, details)
}

func parseDoProjectTransportPlanDriverAssignmentRecommendationsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanDriverAssignmentRecommendationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanDriver, _ := cmd.Flags().GetString("project-transport-plan-driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanDriverAssignmentRecommendationsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ProjectTransportPlanDriver: projectTransportPlanDriver,
	}, nil
}
