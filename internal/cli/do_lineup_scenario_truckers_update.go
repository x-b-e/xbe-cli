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

type doLineupScenarioTruckersUpdateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	ID                               string
	MinimumAssignmentCount           int
	MaximumAssignmentCount           int
	MaximumMinutesToStartSite        int
	MaterialTypeConstraints          string
	TrailerClassificationConstraints string
}

func newDoLineupScenarioTruckersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup scenario trucker",
		Long: `Update a lineup scenario trucker.

Optional flags:
  --minimum-assignment-count         Minimum assignments for the trucker
  --maximum-assignment-count         Maximum assignments for the trucker
  --maximum-minutes-to-start-site    Maximum minutes to the start site
  --material-type-constraints        JSON array of constraints (snake_case keys)
  --trailer-classification-constraints  JSON array of constraints (snake_case keys)

Notes:
  - minimum-assignment-count must be less than or equal to maximum-assignment-count when both are set.`,
		Example: `  # Update assignment limits
  xbe do lineup-scenario-truckers update 123 --minimum-assignment-count 1 --maximum-assignment-count 3

  # Update constraints
  xbe do lineup-scenario-truckers update 123 \\
    --material-type-constraints '[{\"material_type_base_qualified_name\":\"soil\",\"maximum_assignment_count\":2}]' \\
    --trailer-classification-constraints '[{\"trailer_classification_id\":\"789\",\"minimum_assignment_count\":1,\"maximum_assignment_count\":2}]'

  # Output as JSON
  xbe do lineup-scenario-truckers update 123 --maximum-minutes-to-start-site 45 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenarioTruckersUpdate,
	}
	initDoLineupScenarioTruckersUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTruckersCmd.AddCommand(newDoLineupScenarioTruckersUpdateCmd())
}

func initDoLineupScenarioTruckersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("minimum-assignment-count", 0, "Minimum assignments for the trucker")
	cmd.Flags().Int("maximum-assignment-count", 0, "Maximum assignments for the trucker")
	cmd.Flags().Int("maximum-minutes-to-start-site", 0, "Maximum minutes to the start site")
	cmd.Flags().String("material-type-constraints", "", "JSON array of material type constraints")
	cmd.Flags().String("trailer-classification-constraints", "", "JSON array of trailer classification constraints")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioTruckersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenarioTruckersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("minimum-assignment-count") {
		attributes["minimum-assignment-count"] = opts.MinimumAssignmentCount
	}
	if cmd.Flags().Changed("maximum-assignment-count") {
		attributes["maximum-assignment-count"] = opts.MaximumAssignmentCount
	}
	if cmd.Flags().Changed("maximum-minutes-to-start-site") {
		attributes["maximum-minutes-to-start-site"] = opts.MaximumMinutesToStartSite
	}
	if cmd.Flags().Changed("material-type-constraints") {
		constraints, err := parseConstraintArray(opts.MaterialTypeConstraints, "material-type-constraints")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["material-type-constraints"] = constraints
	}
	if cmd.Flags().Changed("trailer-classification-constraints") {
		constraints, err := parseConstraintArray(opts.TrailerClassificationConstraints, "trailer-classification-constraints")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["trailer-classification-constraints"] = constraints
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lineup-scenario-truckers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-scenario-truckers/"+opts.ID, jsonBody)
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
		row := lineupScenarioTruckerRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup scenario trucker %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupScenarioTruckersUpdateOptions(cmd *cobra.Command, args []string) (doLineupScenarioTruckersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	minimumAssignmentCount, _ := cmd.Flags().GetInt("minimum-assignment-count")
	maximumAssignmentCount, _ := cmd.Flags().GetInt("maximum-assignment-count")
	maximumMinutesToStartSite, _ := cmd.Flags().GetInt("maximum-minutes-to-start-site")
	materialTypeConstraints, _ := cmd.Flags().GetString("material-type-constraints")
	trailerClassificationConstraints, _ := cmd.Flags().GetString("trailer-classification-constraints")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTruckersUpdateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		ID:                               args[0],
		MinimumAssignmentCount:           minimumAssignmentCount,
		MaximumAssignmentCount:           maximumAssignmentCount,
		MaximumMinutesToStartSite:        maximumMinutesToStartSite,
		MaterialTypeConstraints:          materialTypeConstraints,
		TrailerClassificationConstraints: trailerClassificationConstraints,
	}, nil
}
