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

type doProjectLaborClassificationsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Project             string
	LaborClassification string
	BasicHourlyRate     string
	FringeHourlyRate    string
}

func newDoProjectLaborClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project labor classification",
		Long: `Create a new project labor classification.

Required flags:
  --project               Project ID (required)
  --labor-classification  Labor classification ID (required)

Optional flags:
  --basic-hourly-rate     Basic hourly rate
  --fringe-hourly-rate    Fringe hourly rate

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project labor classification
  xbe do project-labor-classifications create --project 123 --labor-classification 456

  # Create with rates
  xbe do project-labor-classifications create --project 123 --labor-classification 456 --basic-hourly-rate 45 --fringe-hourly-rate 12

  # JSON output
  xbe do project-labor-classifications create --project 123 --labor-classification 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectLaborClassificationsCreate,
	}
	initDoProjectLaborClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectLaborClassificationsCmd.AddCommand(newDoProjectLaborClassificationsCreateCmd())
}

func initDoProjectLaborClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("labor-classification", "", "Labor classification ID (required)")
	cmd.Flags().String("basic-hourly-rate", "", "Basic hourly rate")
	cmd.Flags().String("fringe-hourly-rate", "", "Fringe hourly rate")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectLaborClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectLaborClassificationsCreateOptions(cmd)
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

	if opts.Project == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.LaborClassification == "" {
		err := fmt.Errorf("--labor-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("basic-hourly-rate") {
		attributes["basic-hourly-rate"] = opts.BasicHourlyRate
	}
	if cmd.Flags().Changed("fringe-hourly-rate") {
		attributes["fringe-hourly-rate"] = opts.FringeHourlyRate
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
		"labor-classification": map[string]any{
			"data": map[string]any{
				"type": "labor-classifications",
				"id":   opts.LaborClassification,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-labor-classifications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-labor-classifications", jsonBody)
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

	row := buildProjectLaborClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project labor classification %s\n", row.ID)
	return nil
}

func parseDoProjectLaborClassificationsCreateOptions(cmd *cobra.Command) (doProjectLaborClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	laborClassification, _ := cmd.Flags().GetString("labor-classification")
	basicHourlyRate, _ := cmd.Flags().GetString("basic-hourly-rate")
	fringeHourlyRate, _ := cmd.Flags().GetString("fringe-hourly-rate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectLaborClassificationsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Project:             project,
		LaborClassification: laborClassification,
		BasicHourlyRate:     basicHourlyRate,
		FringeHourlyRate:    fringeHourlyRate,
	}, nil
}
