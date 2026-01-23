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

type doResourceUnavailabilitiesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ResourceType string
	ResourceID   string
	StartAt      string
	EndAt        string
	Description  string
}

func newDoResourceUnavailabilitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource unavailability",
		Long: `Create a resource unavailability.

Required:
  --resource-type   Resource type (User, Equipment, Trailer, Tractor)
  --resource-id     Resource ID

Optional:
  --start-at        Start timestamp (ISO 8601)
  --end-at          End timestamp (ISO 8601)
  --description     Description/note

Notes:
  If both start-at and end-at are provided, end-at must be after start-at.`,
		Example: `  # Create a user unavailability
  xbe do resource-unavailabilities create \
    --resource-type User \
    --resource-id 123 \
    --start-at 2025-01-01T08:00:00Z \
    --end-at 2025-01-01T17:00:00Z \
    --description "PTO"`,
		Args: cobra.NoArgs,
		RunE: runDoResourceUnavailabilitiesCreate,
	}
	initDoResourceUnavailabilitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doResourceUnavailabilitiesCmd.AddCommand(newDoResourceUnavailabilitiesCreateCmd())
}

func initDoResourceUnavailabilitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("resource-type", "", "Resource type (User, Equipment, Trailer, Tractor)")
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Description/note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoResourceUnavailabilitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoResourceUnavailabilitiesCreateOptions(cmd)
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

	if opts.ResourceType == "" || opts.ResourceID == "" {
		err := fmt.Errorf("--resource-type and --resource-id are required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	resourceType, err := parseResourceUnavailabilityType(opts.ResourceType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	relationships := map[string]any{
		"resource": map[string]any{
			"data": map[string]any{
				"type": resourceType,
				"id":   opts.ResourceID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "resource-unavailabilities",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/resource-unavailabilities", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created resource unavailability %s\n", row.ID)
	return nil
}

func parseDoResourceUnavailabilitiesCreateOptions(cmd *cobra.Command) (doResourceUnavailabilitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	resourceType, _ := cmd.Flags().GetString("resource-type")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doResourceUnavailabilitiesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		StartAt:      startAt,
		EndAt:        endAt,
		Description:  description,
	}, nil
}
