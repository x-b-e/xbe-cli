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

type doProjectEstimateSetsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	Project           string
	Name              string
	CreatedBy         string
	BackupEstimateSet string
}

func newDoProjectEstimateSetsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project estimate set",
		Long: `Create a project estimate set.

Required:
  --project            Project ID

Optional:
  --name               Estimate set name
  --created-by         Created-by user ID
  --backup-estimate-set Backup estimate set ID`,
		Example: `  # Create a project estimate set
  xbe do project-estimate-sets create --project 123 --name "Scenario A"

  # Create with backup estimate set
  xbe do project-estimate-sets create --project 123 --name "Alt" --backup-estimate-set 456

  # Create with created-by
  xbe do project-estimate-sets create --project 123 --name "Alt" --created-by 789`,
		RunE: runDoProjectEstimateSetsCreate,
	}
	initDoProjectEstimateSetsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectEstimateSetsCmd.AddCommand(newDoProjectEstimateSetsCreateCmd())
}

func initDoProjectEstimateSetsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().String("name", "", "Estimate set name")
	cmd.Flags().String("created-by", "", "Created-by user ID")
	cmd.Flags().String("backup-estimate-set", "", "Backup estimate set ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project")
}

func runDoProjectEstimateSetsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectEstimateSetsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}
	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}
	if opts.BackupEstimateSet != "" {
		relationships["backup-estimate-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.BackupEstimateSet,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-estimate-sets",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-estimate-sets", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project estimate set %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectEstimateSetsCreateOptions(cmd *cobra.Command) (doProjectEstimateSetsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	createdBy, _ := cmd.Flags().GetString("created-by")
	backupEstimateSet, _ := cmd.Flags().GetString("backup-estimate-set")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectEstimateSetsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		Project:           project,
		Name:              name,
		CreatedBy:         createdBy,
		BackupEstimateSet: backupEstimateSet,
	}, nil
}
