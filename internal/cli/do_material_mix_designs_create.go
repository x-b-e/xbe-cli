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

type doMaterialMixDesignsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	MaterialType string
	Description  string
	Mix          string
	StartAt      string
	EndAt        string
	Notes        string
}

func newDoMaterialMixDesignsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material mix design",
		Long: `Create a material mix design.

Required:
  --material-type    Material type ID

Optional:
  --description      Description
  --mix              Mix identifier
  --start-at         Start time (ISO 8601)
  --end-at           End time (ISO 8601)
  --notes            Notes`,
		Example: `  # Create a material mix design
  xbe do material-mix-designs create --material-type 123

  # Create with description and mix
  xbe do material-mix-designs create --material-type 123 --description "Concrete Mix A" --mix "MIX-001"

  # Create with date range
  xbe do material-mix-designs create --material-type 123 --start-at "2024-01-01" --end-at "2024-12-31"`,
		RunE: runDoMaterialMixDesignsCreate,
	}
	initDoMaterialMixDesignsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialMixDesignsCmd.AddCommand(newDoMaterialMixDesignsCreateCmd())
}

func initDoMaterialMixDesignsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("mix", "", "Mix identifier")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-type")
}

func runDoMaterialMixDesignsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialMixDesignsCreateOptions(cmd)
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

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Mix != "" {
		attributes["mix"] = opts.Mix
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.EndAt != "" {
		attributes["end-at"] = opts.EndAt
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
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
			"type":          "material-mix-designs",
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

	body, _, err := client.Post(cmd.Context(), "/v1/material-mix-designs", jsonBody)
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
		row := materialMixDesignRow{
			ID:          resp.Data.ID,
			Description: stringAttr(resp.Data.Attributes, "description"),
			Mix:         stringAttr(resp.Data.Attributes, "mix"),
			StartAt:     stringAttr(resp.Data.Attributes, "start-at"),
			EndAt:       stringAttr(resp.Data.Attributes, "end-at"),
			Notes:       stringAttr(resp.Data.Attributes, "notes"),
		}
		if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material mix design %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialMixDesignsCreateOptions(cmd *cobra.Command) (doMaterialMixDesignsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	description, _ := cmd.Flags().GetString("description")
	mix, _ := cmd.Flags().GetString("mix")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialMixDesignsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		MaterialType: materialType,
		Description:  description,
		Mix:          mix,
		StartAt:      startAt,
		EndAt:        endAt,
		Notes:        notes,
	}, nil
}
