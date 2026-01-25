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

type doProjectTransportLocationEventTypesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ProjectTransportLocation  string
	ProjectTransportEventType string
}

func newDoProjectTransportLocationEventTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport location event type",
		Long: `Create a project transport location event type.

Required flags:
  --project-transport-location   Project transport location ID (required)
  --project-transport-event-type Project transport event type ID (required)`,
		Example: `  # Link a transport event type to a location
  xbe do project-transport-location-event-types create \\
    --project-transport-location 123 \\
    --project-transport-event-type 456

  # JSON output
  xbe do project-transport-location-event-types create \\
    --project-transport-location 123 \\
    --project-transport-event-type 456 \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportLocationEventTypesCreate,
	}
	initDoProjectTransportLocationEventTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportLocationEventTypesCmd.AddCommand(newDoProjectTransportLocationEventTypesCreateCmd())
}

func initDoProjectTransportLocationEventTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID (required)")
	cmd.Flags().String("project-transport-event-type", "", "Project transport event type ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportLocationEventTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportLocationEventTypesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportLocation) == "" {
		err := fmt.Errorf("--project-transport-location is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectTransportEventType) == "" {
		err := fmt.Errorf("--project-transport-event-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"project-transport-location": map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.ProjectTransportLocation,
			},
		},
		"project-transport-event-type": map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.ProjectTransportEventType,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-location-event-types",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-location-event-types", jsonBody)
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

	row := buildProjectTransportLocationEventTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport location event type %s\n", row.ID)
	return nil
}

func parseDoProjectTransportLocationEventTypesCreateOptions(cmd *cobra.Command) (doProjectTransportLocationEventTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	location, _ := cmd.Flags().GetString("project-transport-location")
	eventType, _ := cmd.Flags().GetString("project-transport-event-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportLocationEventTypesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ProjectTransportLocation:  location,
		ProjectTransportEventType: eventType,
	}, nil
}

func buildProjectTransportLocationEventTypeRowFromSingle(resp jsonAPISingleResponse) projectTransportLocationEventTypeRow {
	row := projectTransportLocationEventTypeRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["project-transport-location"]; ok && rel.Data != nil {
		row.ProjectTransportLocationID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		row.ProjectTransportEventTypeID = rel.Data.ID
	}

	return row
}
