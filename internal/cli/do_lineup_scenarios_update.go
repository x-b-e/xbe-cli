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

type doLineupScenariosUpdateOptions struct {
	BaseURL                                string
	Token                                  string
	JSON                                   bool
	ID                                     string
	Name                                   string
	IncludeTruckerAssignmentsAsConstraints bool
}

func newDoLineupScenariosUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup scenario",
		Long: `Update a lineup scenario.

Provide the scenario ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --name                                       Scenario name
  --include-trucker-assignments-as-constraints Include trucker assignments as constraints`,
		Example: `  # Update the scenario name
  xbe do lineup-scenarios update 123 --name "Updated scenario"

  # Update constraints flag
  xbe do lineup-scenarios update 123 --include-trucker-assignments-as-constraints=false

  # JSON output
  xbe do lineup-scenarios update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenariosUpdate,
	}
	initDoLineupScenariosUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenariosCmd.AddCommand(newDoLineupScenariosUpdateCmd())
}

func initDoLineupScenariosUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Scenario name")
	cmd.Flags().Bool("include-trucker-assignments-as-constraints", false, "Include trucker assignments as constraints")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenariosUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenariosUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("include-trucker-assignments-as-constraints") {
		attributes["include-trucker-assignments-as-constraints"] = opts.IncludeTruckerAssignmentsAsConstraints
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --include-trucker-assignments-as-constraints")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lineup-scenarios",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-scenarios/"+opts.ID, jsonBody)
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

	row := buildLineupScenarioRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup scenario %s\n", row.ID)
	return nil
}

func parseDoLineupScenariosUpdateOptions(cmd *cobra.Command, args []string) (doLineupScenariosUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	includeAssignments, _ := cmd.Flags().GetBool("include-trucker-assignments-as-constraints")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenariosUpdateOptions{
		BaseURL:                                baseURL,
		Token:                                  token,
		JSON:                                   jsonOut,
		ID:                                     args[0],
		Name:                                   name,
		IncludeTruckerAssignmentsAsConstraints: includeAssignments,
	}, nil
}
