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

type doDeveloperTruckerCertificationClassificationsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
}

func newDoDeveloperTruckerCertificationClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing developer trucker certification classification",
		Long: `Update an existing developer trucker certification classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name  The classification name`,
		Example: `  # Update name
  xbe do developer-trucker-certification-classifications update 123 --name "Updated Name"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperTruckerCertificationClassificationsUpdate,
	}
	initDoDeveloperTruckerCertificationClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationClassificationsCmd.AddCommand(newDoDeveloperTruckerCertificationClassificationsUpdateCmd())
}

func initDoDeveloperTruckerCertificationClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperTruckerCertificationClassificationsUpdateOptions(cmd, args)
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --name")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "developer-trucker-certification-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-trucker-certification-classifications/"+opts.ID, jsonBody)
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

	row := buildDeveloperTruckerCertificationClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer trucker certification classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDeveloperTruckerCertificationClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperTruckerCertificationClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationClassificationsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
	}, nil
}
