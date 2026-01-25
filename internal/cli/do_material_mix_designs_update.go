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

type doMaterialMixDesignsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Description string
	Mix         string
	StartAt     string
	EndAt       string
	Notes       string
}

func newDoMaterialMixDesignsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material mix design",
		Long: `Update a material mix design.

Optional:
  --description      Description
  --mix              Mix identifier
  --start-at         Start time (ISO 8601)
  --end-at           End time (ISO 8601)
  --notes            Notes`,
		Example: `  # Update description
  xbe do material-mix-designs update 123 --description "Updated Mix Design"

  # Update date range
  xbe do material-mix-designs update 123 --start-at "2024-01-01" --end-at "2024-12-31"

  # Update notes
  xbe do material-mix-designs update 123 --notes "New notes"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialMixDesignsUpdate,
	}
	initDoMaterialMixDesignsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialMixDesignsCmd.AddCommand(newDoMaterialMixDesignsUpdateCmd())
}

func initDoMaterialMixDesignsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("mix", "", "Mix identifier")
	cmd.Flags().String("start-at", "", "Start time (ISO 8601)")
	cmd.Flags().String("end-at", "", "End time (ISO 8601)")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialMixDesignsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialMixDesignsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("mix") {
		attributes["mix"] = opts.Mix
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
	}
	if cmd.Flags().Changed("end-at") {
		attributes["end-at"] = opts.EndAt
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "material-mix-designs",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-mix-designs/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material mix design %s\n", resp.Data.ID)
	return nil
}

func parseDoMaterialMixDesignsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialMixDesignsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	mix, _ := cmd.Flags().GetString("mix")
	startAt, _ := cmd.Flags().GetString("start-at")
	endAt, _ := cmd.Flags().GetString("end-at")
	notes, _ := cmd.Flags().GetString("notes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialMixDesignsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Description: description,
		Mix:         mix,
		StartAt:     startAt,
		EndAt:       endAt,
		Notes:       notes,
	}, nil
}
