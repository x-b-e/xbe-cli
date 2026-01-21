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

type doCultureValuesUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Description string
	Position    string
}

func newDoCultureValuesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a culture value",
		Long: `Update an existing culture value.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The culture value ID (required)

Flags:
  --name         Update the name
  --description  Update the description
  --position     Update the sequence position (0-based index within organization)`,
		Example: `  # Update just the name
  xbe do culture-values update 456 --name "New Name"

  # Update the position (move to position 0, i.e., first)
  xbe do culture-values update 456 --position 0

  # Update multiple fields
  xbe do culture-values update 456 --name "Excellence" --description "Strive for excellence"

  # Get JSON output
  xbe do culture-values update 456 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCultureValuesUpdate,
	}
	initDoCultureValuesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCultureValuesCmd.AddCommand(newDoCultureValuesUpdateCmd())
}

func initDoCultureValuesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("position", "", "New sequence position (0-based index)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCultureValuesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCultureValuesUpdateOptions(cmd)
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
		return fmt.Errorf("culture value id is required")
	}

	// Require at least one field to update
	if opts.Name == "" && opts.Description == "" && opts.Position == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Position != "" {
		attributes["sequence-position"] = opts.Position
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "culture-values",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/culture-values/"+id, jsonBody)
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

	row := buildCultureValueRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated culture value %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCultureValuesUpdateOptions(cmd *cobra.Command) (doCultureValuesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	position, _ := cmd.Flags().GetString("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCultureValuesUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Description: description,
		Position:    position,
	}, nil
}
