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

type doProjectEstimateSetsUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	Name              string
	CreatedBy         string
	BackupEstimateSet string
}

func newDoProjectEstimateSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project estimate set",
		Long: `Update a project estimate set.

Optional:
  --name               New estimate set name
  --created-by         Created-by user ID
  --backup-estimate-set Backup estimate set ID`,
		Example: `  # Update the estimate set name
  xbe do project-estimate-sets update 123 --name "Updated"

  # Update backup estimate set
  xbe do project-estimate-sets update 123 --backup-estimate-set 456

  # Update created-by
  xbe do project-estimate-sets update 123 --created-by 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectEstimateSetsUpdate,
	}
	initDoProjectEstimateSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectEstimateSetsCmd.AddCommand(newDoProjectEstimateSetsUpdateCmd())
}

func initDoProjectEstimateSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Estimate set name")
	cmd.Flags().String("created-by", "", "Created-by user ID")
	cmd.Flags().String("backup-estimate-set", "", "Backup estimate set ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectEstimateSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectEstimateSetsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("created-by") {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}
	if cmd.Flags().Changed("backup-estimate-set") {
		relationships["backup-estimate-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.BackupEstimateSet,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-estimate-sets",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-estimate-sets/"+opts.ID, jsonBody)
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
		row := buildProjectEstimateSetRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project estimate set %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectEstimateSetsUpdateOptions(cmd *cobra.Command, args []string) (doProjectEstimateSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	createdBy, _ := cmd.Flags().GetString("created-by")
	backupEstimateSet, _ := cmd.Flags().GetString("backup-estimate-set")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectEstimateSetsUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		Name:              name,
		CreatedBy:         createdBy,
		BackupEstimateSet: backupEstimateSet,
	}, nil
}
