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

type doProjectProjectCostClassificationsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	NameOverride string
}

func newDoProjectProjectCostClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project project cost classification",
		Long: `Update an existing project project cost classification.

Provide the project project cost classification ID as an argument, then use flags
to specify which fields to update. Only specified fields will be modified.

Updatable fields:
  --name-override  Override the classification name for this project`,
		Example: `  # Update name override
  xbe do project-project-cost-classifications update 123 --name-override "Custom Name"

  # Get JSON output
  xbe do project-project-cost-classifications update 123 --name-override "Custom Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectProjectCostClassificationsUpdate,
	}
	initDoProjectProjectCostClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectProjectCostClassificationsCmd.AddCommand(newDoProjectProjectCostClassificationsUpdateCmd())
}

func initDoProjectProjectCostClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name-override", "", "Override the classification name for this project")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectProjectCostClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectProjectCostClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name-override") {
		attributes["name-override"] = opts.NameOverride
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --name-override")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-project-cost-classifications",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-project-cost-classifications/"+opts.ID, jsonBody)
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

	row := buildProjectProjectCostClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project project cost classification %s\n", row.ID)
	return nil
}

func parseDoProjectProjectCostClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectProjectCostClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	nameOverride, _ := cmd.Flags().GetString("name-override")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectProjectCostClassificationsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		NameOverride: nameOverride,
	}, nil
}
