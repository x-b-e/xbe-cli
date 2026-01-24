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

type doProjectPhaseDatesEstimatesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ProjectPhase       string
	ProjectEstimateSet string
	CreatedBy          string
	StartDate          string
	EndDate            string
}

func newDoProjectPhaseDatesEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase dates estimate",
		Long: `Create a project phase dates estimate.

Required:
  --project-phase        Project phase ID
  --project-estimate-set Project estimate set ID
  --start-date           Estimated start date (YYYY-MM-DD)
  --end-date             Estimated end date (YYYY-MM-DD)

Optional:
  --created-by            Creator user ID`,
		Example: `  # Create a dates estimate
  xbe do project-phase-dates-estimates create \
    --project-phase 123 \
    --project-estimate-set 456 \
    --start-date 2025-01-01 \
    --end-date 2025-01-15

  # Output as JSON
  xbe do project-phase-dates-estimates create \
    --project-phase 123 \
    --project-estimate-set 456 \
    --start-date 2025-01-01 \
    --end-date 2025-01-15 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseDatesEstimatesCreate,
	}
	initDoProjectPhaseDatesEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseDatesEstimatesCmd.AddCommand(newDoProjectPhaseDatesEstimatesCreateCmd())
}

func initDoProjectPhaseDatesEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase", "", "Project phase ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID (required)")
	cmd.Flags().String("start-date", "", "Estimated start date (required, YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "Estimated end date (required, YYYY-MM-DD)")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseDatesEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseDatesEstimatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectPhase) == "" {
		err := fmt.Errorf("--project-phase is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ProjectEstimateSet) == "" {
		err := fmt.Errorf("--project-estimate-set is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartDate) == "" {
		err := fmt.Errorf("--start-date is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.EndDate) == "" {
		err := fmt.Errorf("--end-date is required")
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

	attributes := map[string]any{
		"start-date": opts.StartDate,
		"end-date":   opts.EndDate,
	}

	relationships := map[string]any{
		"project-phase": map[string]any{
			"data": map[string]any{
				"type": "project-phases",
				"id":   opts.ProjectPhase,
			},
		},
		"project-estimate-set": map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.ProjectEstimateSet,
			},
		},
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	data := map[string]any{
		"type":          "project-phase-dates-estimates",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-dates-estimates", jsonBody)
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

	row := projectPhaseDatesEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase dates estimate %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseDatesEstimatesCreateOptions(cmd *cobra.Command) (doProjectPhaseDatesEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseDatesEstimatesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ProjectPhase:       projectPhase,
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
		StartDate:          startDate,
		EndDate:            endDate,
	}, nil
}
