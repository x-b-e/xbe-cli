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

type doBusinessUnitsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Name       string
	ExternalID string
	Parent     string
}

func newDoBusinessUnitsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a business unit",
		Long: `Update an existing business unit.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The business unit ID (required)

Flags:
  --name          Update the name
  --external-id   Update the external identifier
  --parent        Update the parent business unit ID`,
		Example: `  # Update the name
  xbe do business-units update 123 --name "New Division Name"

  # Update external ID
  xbe do business-units update 123 --external-id "BU-002"

  # Update parent
  xbe do business-units update 123 --parent 456

  # Get JSON output
  xbe do business-units update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBusinessUnitsUpdate,
	}
	initDoBusinessUnitsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitsCmd.AddCommand(newDoBusinessUnitsUpdateCmd())
}

func initDoBusinessUnitsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("external-id", "", "New external identifier")
	cmd.Flags().String("parent", "", "New parent business unit ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBusinessUnitsUpdateOptions(cmd)
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
		return fmt.Errorf("business unit id is required")
	}

	// Require at least one field to update
	if opts.Name == "" && opts.ExternalID == "" && opts.Parent == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["company-name"] = opts.Name
	}
	if opts.ExternalID != "" {
		attributes["external-id"] = opts.ExternalID
	}

	data := map[string]any{
		"id":         id,
		"type":       "business-units",
		"attributes": attributes,
	}

	// Build relationships if parent is provided
	if opts.Parent != "" {
		data["relationships"] = map[string]any{
			"parent": map[string]any{
				"data": map[string]string{
					"type": "business-units",
					"id":   opts.Parent,
				},
			},
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

	body, _, err := client.Patch(cmd.Context(), "/v1/business-units/"+id, jsonBody)
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

	row := businessUnitRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "company-name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated business unit %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoBusinessUnitsUpdateOptions(cmd *cobra.Command) (doBusinessUnitsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	externalID, _ := cmd.Flags().GetString("external-id")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Name:       name,
		ExternalID: externalID,
		Parent:     parent,
	}, nil
}
