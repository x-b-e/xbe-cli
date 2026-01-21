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

type doQualityControlClassificationsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Name        string
	Description string
}

func newDoQualityControlClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing quality control classification",
		Long: `Update an existing quality control classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name         The classification name
  --description  Description of the classification

Note: The broker cannot be changed after creation.`,
		Example: `  # Update name
  xbe do quality-control-classifications update 123 --name "Updated Name"

  # Update description
  xbe do quality-control-classifications update 123 --description "New description"

  # Update multiple fields
  xbe do quality-control-classifications update 123 --name "Density Test" --description "Check material density"

  # Get JSON output
  xbe do quality-control-classifications update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoQualityControlClassificationsUpdate,
	}
	initDoQualityControlClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doQualityControlClassificationsCmd.AddCommand(newDoQualityControlClassificationsUpdateCmd())
}

func initDoQualityControlClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoQualityControlClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoQualityControlClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --description")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "quality-control-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/quality-control-classifications/"+opts.ID, jsonBody)
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

	row := buildQualityControlClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated quality control classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoQualityControlClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doQualityControlClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doQualityControlClassificationsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Name:        name,
		Description: description,
	}, nil
}
