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

type doProjectTransportPlanSegmentDriversCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ProjectTransportPlanSegment string
	Driver                      string
}

func newDoProjectTransportPlanSegmentDriversCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan segment driver",
		Long: `Create a project transport plan segment driver.

Required flags:
  --project-transport-plan-segment  Project transport plan segment ID
  --driver                          Driver (user) ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project transport plan segment driver
  xbe do project-transport-plan-segment-drivers create \
    --project-transport-plan-segment 123 \
    --driver 456`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanSegmentDriversCreate,
	}
	initDoProjectTransportPlanSegmentDriversCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanSegmentDriversCmd.AddCommand(newDoProjectTransportPlanSegmentDriversCreateCmd())
}

func initDoProjectTransportPlanSegmentDriversCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-segment", "", "Project transport plan segment ID")
	cmd.Flags().String("driver", "", "Driver (user) ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanSegmentDriversCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanSegmentDriversCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlanSegment) == "" {
		err := fmt.Errorf("--project-transport-plan-segment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Driver) == "" {
		err := fmt.Errorf("--driver is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-plan-segment": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.ProjectTransportPlanSegment,
			},
		},
		"driver": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Driver,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-segment-drivers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-segment-drivers", jsonBody)
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

	if opts.JSON {
		row := buildProjectTransportPlanSegmentDriverRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(row) > 0 {
			return writeJSON(cmd.OutOrStdout(), row[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan segment driver %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectTransportPlanSegmentDriversCreateOptions(cmd *cobra.Command) (doProjectTransportPlanSegmentDriversCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	segment, _ := cmd.Flags().GetString("project-transport-plan-segment")
	driver, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanSegmentDriversCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ProjectTransportPlanSegment: segment,
		Driver:                      driver,
	}, nil
}
