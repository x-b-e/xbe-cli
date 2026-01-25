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

type doUiToursUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	Abbreviation string
	Description  string
}

func newDoUiToursUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a UI tour",
		Long: `Update an existing UI tour.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>  The UI tour ID (required)

Flags:
  --name          Update the UI tour name
  --abbreviation  Update the UI tour abbreviation
  --description   Update the UI tour description

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update UI tour name
  xbe do ui-tours update 123 --name "Updated Name"

  # Update abbreviation
  xbe do ui-tours update 123 --abbreviation "updated-abbrev"

  # Update description
  xbe do ui-tours update 123 --description "New description"

  # Get JSON output
  xbe do ui-tours update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUiToursUpdate,
	}
	initDoUiToursUpdateFlags(cmd)
	return cmd
}

func init() {
	doUiToursCmd.AddCommand(newDoUiToursUpdateCmd())
}

func initDoUiToursUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New UI tour name")
	cmd.Flags().String("abbreviation", "", "New UI tour abbreviation")
	cmd.Flags().String("description", "", "New UI tour description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUiToursUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUiToursUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("abbreviation") {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --abbreviation, --description")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         opts.ID,
			"type":       "ui-tours",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/ui-tours/"+opts.ID, jsonBody)
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

	row := buildUiTourRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated UI tour %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoUiToursUpdateOptions(cmd *cobra.Command, args []string) (doUiToursUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doUiToursUpdateOptions{}, fmt.Errorf("ui tour id is required")
	}

	return doUiToursUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           id,
		Name:         name,
		Abbreviation: abbreviation,
		Description:  description,
	}, nil
}
