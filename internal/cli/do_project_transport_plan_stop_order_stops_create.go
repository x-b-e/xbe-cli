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

type doProjectTransportPlanStopOrderStopsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ProjectTransportPlanStop string
	TransportOrderStop       string
}

func newDoProjectTransportPlanStopOrderStopsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan stop order stop",
		Long: `Create a project transport plan stop order stop.

Required flags:
  --project-transport-plan-stop  Project transport plan stop ID (required)
  --transport-order-stop         Transport order stop ID (required)`,
		Example: `  # Link a plan stop to an order stop
  xbe do project-transport-plan-stop-order-stops create --project-transport-plan-stop 123 --transport-order-stop 456`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanStopOrderStopsCreate,
	}
	initDoProjectTransportPlanStopOrderStopsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanStopOrderStopsCmd.AddCommand(newDoProjectTransportPlanStopOrderStopsCreateCmd())
}

func initDoProjectTransportPlanStopOrderStopsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-stop", "", "Project transport plan stop ID (required)")
	cmd.Flags().String("transport-order-stop", "", "Transport order stop ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanStopOrderStopsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanStopOrderStopsCreateOptions(cmd)
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

	if opts.ProjectTransportPlanStop == "" {
		err := fmt.Errorf("--project-transport-plan-stop is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TransportOrderStop == "" {
		err := fmt.Errorf("--transport-order-stop is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-plan-stop": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-stops",
				"id":   opts.ProjectTransportPlanStop,
			},
		},
		"transport-order-stop": map[string]any{
			"data": map[string]any{
				"type": "transport-order-stops",
				"id":   opts.TransportOrderStop,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-stop-order-stops",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-stop-order-stops", jsonBody)
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

	row := projectTransportPlanStopOrderStopRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan stop order stop %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanStopOrderStopsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanStopOrderStopsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanStop, _ := cmd.Flags().GetString("project-transport-plan-stop")
	transportOrderStop, _ := cmd.Flags().GetString("transport-order-stop")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanStopOrderStopsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ProjectTransportPlanStop: projectTransportPlanStop,
		TransportOrderStop:       transportOrderStop,
	}, nil
}
