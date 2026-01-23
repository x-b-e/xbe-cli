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

type doResourceUnavailabilitiesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	ResourceType string
	ResourceID   string
	StartAt      string
	EndAt        string
	Description  string
}

func newDoResourceUnavailabilitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a resource unavailability",
		Long: `Update a resource unavailability.

Optional:
  --resource-type   Resource type (User, Equipment, Trailer, Tractor)
  --resource-id     Resource ID (requires --resource-type)
  --start-at        Start timestamp (ISO 8601)
  --end-at          End timestamp (ISO 8601)
  --description     Description/note`,
		Example: `  # Update time window
  xbe do resource-unavailabilities update 123 \
    --start-at 2025-02-01T08:00:00Z \
    --end-at 2025-02-01T17:00:00Z

  # Update description
  xbe do resource-unavailabilities update 123 --description "Updated note"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoResourceUnavailabilitiesUpdate,
	}
	initDoResourceUnavailabilitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doResourceUnavailabilitiesCmd.AddCommand(newDoResourceUnavailabilitiesUpdateCmd())
}

func initDoResourceUnavailabilitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("resource-type", "", "Resource type (User, Equipment, Trailer, Tractor)")
	cmd.Flags().String("resource-id", "", "Resource ID (requires --resource-type)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Description/note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoResourceUnavailabilitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoResourceUnavailabilitiesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	resourceTypeChanged := cmd.Flags().Changed("resource-type")
	resourceIDChanged := cmd.Flags().Changed("resource-id")
	if resourceTypeChanged || resourceIDChanged {
		if opts.ResourceType == "" || opts.ResourceID == "" {
			err := fmt.Errorf("--resource-type and --resource-id must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		resourceType, err := parseResourceUnavailabilityType(opts.ResourceType)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["resource"] = map[string]any{
			"data": map[string]any{
				"type": resourceType,
				"id":   opts.ResourceID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "resource-unavailabilities",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/resource-unavailabilities/"+opts.ID, jsonBody)
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

	row := buildResourceUnavailabilityRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated resource unavailability %s\n", row.ID)
	return nil
}

func parseDoResourceUnavailabilitiesUpdateOptions(cmd *cobra.Command, args []string) (doResourceUnavailabilitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doResourceUnavailabilitiesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		ResourceType: resourceType,
		ResourceID:   resourceID,
		StartAt:      startAt,
		EndAt:        endAt,
		Description:  description,
	}, nil
}
