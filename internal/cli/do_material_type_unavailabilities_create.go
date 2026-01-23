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

type doMaterialTypeUnavailabilitiesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	MaterialType string
	StartAt      string
	EndAt        string
	Description  string
}

func newDoMaterialTypeUnavailabilitiesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material type unavailability",
		Long: `Create a material type unavailability.

Required:
  --material-type   Material type ID

At least one time bound is required:
  --start-at        Start timestamp (ISO 8601)
  --end-at          End timestamp (ISO 8601)

Optional:
  --description     Description`,
		Example: `  # Create a material type unavailability with a start time
  xbe do material-type-unavailabilities create --material-type 123 --start-at 2025-01-01T00:00:00Z

  # Create with a full window and description
  xbe do material-type-unavailabilities create --material-type 123 --start-at 2025-01-01T00:00:00Z --end-at 2025-01-02T00:00:00Z --description "Plant maintenance"`,
		RunE: runDoMaterialTypeUnavailabilitiesCreate,
	}
	initDoMaterialTypeUnavailabilitiesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeUnavailabilitiesCmd.AddCommand(newDoMaterialTypeUnavailabilitiesCreateCmd())
}

func initDoMaterialTypeUnavailabilitiesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO 8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO 8601)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-type")
}

func runDoMaterialTypeUnavailabilitiesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTypeUnavailabilitiesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.StartAt) == "" && strings.TrimSpace(opts.EndAt) == "" {
		err := fmt.Errorf("at least one of --start-at or --end-at is required")
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
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-type-unavailabilities",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-type-unavailabilities", jsonBody)
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
		row := materialTypeUnavailabilityRow{
			ID:          resp.Data.ID,
			StartAt:     formatDateTime(stringAttr(resp.Data.Attributes, "start-at")),
			EndAt:       formatDateTime(stringAttr(resp.Data.Attributes, "end-at")),
			Description: stringAttr(resp.Data.Attributes, "description"),
		}
		if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material type unavailability %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialTypeUnavailabilitiesCreateOptions(cmd *cobra.Command) (doMaterialTypeUnavailabilitiesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeUnavailabilitiesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		MaterialType: materialType,
		StartAt:      startAt,
		EndAt:        endAt,
		Description:  description,
	}, nil
}
