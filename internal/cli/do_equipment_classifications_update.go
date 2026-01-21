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

type doEquipmentClassificationsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	Abbreviation       string
	MobilizationMethod string
	Parent             string
}

func newDoEquipmentClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment classification",
		Long: `Update an existing equipment classification.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The equipment classification ID (required)

Flags:
  --name                  Update the name
  --abbreviation          Update the abbreviation
  --mobilization-method   Update how equipment is mobilized
  --parent                Update the parent classification ID`,
		Example: `  # Update just the name
  xbe do equipment-classifications update 456 --name "Updated Name"

  # Update mobilization method
  xbe do equipment-classifications update 456 --mobilization-method trailer

  # Change parent
  xbe do equipment-classifications update 456 --parent 123

  # Get JSON output
  xbe do equipment-classifications update 456 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentClassificationsUpdate,
	}
	initDoEquipmentClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentClassificationsCmd.AddCommand(newDoEquipmentClassificationsUpdateCmd())
}

func initDoEquipmentClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("abbreviation", "", "New abbreviation")
	cmd.Flags().String("mobilization-method", "", "New mobilization method")
	cmd.Flags().String("parent", "", "New parent classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentClassificationsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("equipment classification id is required")
	}

	// Check if at least one field is being updated
	hasUpdate := opts.Name != "" || opts.Abbreviation != "" ||
		opts.MobilizationMethod != "" || cmd.Flags().Changed("parent")

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}

	// Build request data
	data := map[string]any{
		"id":         id,
		"type":       "equipment-classifications",
		"attributes": attributes,
	}

	// Add parent relationship if specified
	if cmd.Flags().Changed("parent") {
		if opts.Parent != "" {
			data["relationships"] = map[string]any{
				"parent": map[string]any{
					"data": map[string]string{
						"type": "equipment-classifications",
						"id":   opts.Parent,
					},
				},
			}
		} else {
			// Clear parent by setting to null
			data["relationships"] = map[string]any{
				"parent": map[string]any{
					"data": nil,
				},
			}
		}
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-classifications/"+id, jsonBody)
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

	row := buildEquipmentClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoEquipmentClassificationsUpdateOptions(cmd *cobra.Command) (doEquipmentClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentClassificationsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		Abbreviation:       abbreviation,
		MobilizationMethod: mobilizationMethod,
		Parent:             parent,
	}, nil
}
