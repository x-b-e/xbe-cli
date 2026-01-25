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

type doProjectTrailerClassificationsUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
	ProjectLaborClassification string
}

func newDoProjectTrailerClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project trailer classification",
		Long: `Update an existing project trailer classification.

Only the project labor classification can be updated.

Arguments:
  <id>    The project trailer classification ID (required)

Flags:
  --project-labor-classification  Project labor classification ID (use empty string to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Set a project labor classification
  xbe do project-trailer-classifications update 123 --project-labor-classification 789

  # Clear the project labor classification
  xbe do project-trailer-classifications update 123 --project-labor-classification ""

  # Get JSON output
  xbe do project-trailer-classifications update 123 --project-labor-classification 789 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTrailerClassificationsUpdate,
	}
	initDoProjectTrailerClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTrailerClassificationsCmd.AddCommand(newDoProjectTrailerClassificationsUpdateCmd())
}

func initDoProjectTrailerClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-labor-classification", "", "Project labor classification ID (use empty string to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTrailerClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTrailerClassificationsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("project-labor-classification") {
		if opts.ProjectLaborClassification == "" {
			relationships["project-labor-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-labor-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-labor-classifications",
					"id":   opts.ProjectLaborClassification,
				},
			}
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify --project-labor-classification")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "project-trailer-classifications",
		"id":            opts.ID,
		"relationships": relationships,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-trailer-classifications/"+opts.ID, jsonBody)
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

	details := buildProjectTrailerClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project trailer classification %s\n", details.ID)
	return nil
}

func parseDoProjectTrailerClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTrailerClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectLaborClassification, _ := cmd.Flags().GetString("project-labor-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTrailerClassificationsUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
		ProjectLaborClassification: projectLaborClassification,
	}, nil
}
