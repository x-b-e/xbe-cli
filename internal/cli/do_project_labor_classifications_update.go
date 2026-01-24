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

type doProjectLaborClassificationsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	BasicHourlyRate  string
	FringeHourlyRate string
}

func newDoProjectLaborClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project labor classification",
		Long: `Update an existing project labor classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --basic-hourly-rate   Basic hourly rate
  --fringe-hourly-rate  Fringe hourly rate

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update the basic hourly rate
  xbe do project-labor-classifications update 123 --basic-hourly-rate 50

  # Update both rates
  xbe do project-labor-classifications update 123 --basic-hourly-rate 50 --fringe-hourly-rate 12

  # JSON output
  xbe do project-labor-classifications update 123 --basic-hourly-rate 50 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectLaborClassificationsUpdate,
	}
	initDoProjectLaborClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectLaborClassificationsCmd.AddCommand(newDoProjectLaborClassificationsUpdateCmd())
}

func initDoProjectLaborClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("basic-hourly-rate", "", "Basic hourly rate")
	cmd.Flags().String("fringe-hourly-rate", "", "Fringe hourly rate")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectLaborClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectLaborClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("basic-hourly-rate") {
		attributes["basic-hourly-rate"] = opts.BasicHourlyRate
	}
	if cmd.Flags().Changed("fringe-hourly-rate") {
		attributes["fringe-hourly-rate"] = opts.FringeHourlyRate
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --basic-hourly-rate, --fringe-hourly-rate")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-labor-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-labor-classifications/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project labor classification %s\n", row.ID)
	return nil
}

func parseDoProjectLaborClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectLaborClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	basicHourlyRate, _ := cmd.Flags().GetString("basic-hourly-rate")
	fringeHourlyRate, _ := cmd.Flags().GetString("fringe-hourly-rate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectLaborClassificationsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		BasicHourlyRate:  basicHourlyRate,
		FringeHourlyRate: fringeHourlyRate,
	}, nil
}
