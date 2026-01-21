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

type doProjectRevenueClassificationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
	Parent  string
}

func newDoProjectRevenueClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project revenue classification",
		Long: `Update an existing project revenue classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name    The classification name
  --parent  Parent classification ID`,
		Example: `  # Update name
  xbe do project-revenue-classifications update 123 --name "Updated Sales"

  # Set parent
  xbe do project-revenue-classifications update 123 --parent 456

  # Get JSON output
  xbe do project-revenue-classifications update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectRevenueClassificationsUpdate,
	}
	initDoProjectRevenueClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueClassificationsCmd.AddCommand(newDoProjectRevenueClassificationsUpdateCmd())
}

func initDoProjectRevenueClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name")
	cmd.Flags().String("parent", "", "Parent classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectRevenueClassificationsUpdateOptions(cmd, args)
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

	var relationships map[string]any
	if cmd.Flags().Changed("parent") {
		relationships = map[string]any{}
		if opts.Parent == "" {
			relationships["parent"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["parent"] = map[string]any{
				"data": map[string]any{
					"type": "project-revenue-classifications",
					"id":   opts.Parent,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --parent")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-revenue-classifications",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-revenue-classifications/"+opts.ID, jsonBody)
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

	row := buildProjectRevenueClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project revenue classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectRevenueClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectRevenueClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	parent, _ := cmd.Flags().GetString("parent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueClassificationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
		Parent:  parent,
	}, nil
}
