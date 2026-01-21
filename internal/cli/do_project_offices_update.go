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

type doProjectOfficesUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	Abbreviation string
	IsActive     bool
}

func newDoProjectOfficesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project office",
		Long: `Update an existing project office.

Provide the project office ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name         The project office name
  --abbreviation Short code
  --is-active    Whether active`,
		Example: `  # Update name
  xbe do project-offices update 123 --name "Updated Name"

  # Update multiple fields
  xbe do project-offices update 123 --name "New Name" --abbreviation "NEW"

  # Deactivate
  xbe do project-offices update 123 --is-active=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectOfficesUpdate,
	}
	initDoProjectOfficesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectOfficesCmd.AddCommand(newDoProjectOfficesUpdateCmd())
}

func initDoProjectOfficesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Project office name")
	cmd.Flags().String("abbreviation", "", "Short code")
	cmd.Flags().Bool("is-active", true, "Whether active")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectOfficesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectOfficesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --abbreviation, --is-active")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-offices",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-offices/"+opts.ID, jsonBody)
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

	row := buildProjectOfficeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project office %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectOfficesUpdateOptions(cmd *cobra.Command, args []string) (doProjectOfficesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	isActive, _ := cmd.Flags().GetBool("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectOfficesUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Name:         name,
		Abbreviation: abbreviation,
		IsActive:     isActive,
	}, nil
}
