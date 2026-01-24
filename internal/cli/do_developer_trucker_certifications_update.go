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

type doDeveloperTruckerCertificationsUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	Classification    string
	StartOn           string
	EndOn             string
	DefaultMultiplier string
}

func newDoDeveloperTruckerCertificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a developer trucker certification",
		Long: `Update a developer trucker certification.

Only the fields you specify will be updated. Developer and trucker cannot be
changed after creation.

Arguments:
  <id>    The developer trucker certification ID (required)

Optional flags:
  --classification     Classification ID
  --start-on           Start date (YYYY-MM-DD)
  --end-on             End date (YYYY-MM-DD)
  --default-multiplier Default multiplier (numeric)`,
		Example: `  # Update dates and default multiplier
  xbe do developer-trucker-certifications update 123 --start-on 2024-02-01 --end-on 2024-12-31 --default-multiplier 1.25

  # Update classification
  xbe do developer-trucker-certifications update 123 --classification 456

  # Output as JSON
  xbe do developer-trucker-certifications update 123 --default-multiplier 1.3 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperTruckerCertificationsUpdate,
	}
	initDoDeveloperTruckerCertificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationsCmd.AddCommand(newDoDeveloperTruckerCertificationsUpdateCmd())
}

func initDoDeveloperTruckerCertificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("classification", "", "Classification ID")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("default-multiplier", "", "Default multiplier (numeric)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperTruckerCertificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("default-multiplier") {
		attributes["default-multiplier"] = opts.DefaultMultiplier
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("classification") {
		relationships["classification"] = map[string]any{
			"data": map[string]any{
				"type": "developer-trucker-certification-classifications",
				"id":   opts.Classification,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "developer-trucker-certifications",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-trucker-certifications/"+opts.ID, jsonBody)
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

	row := developerTruckerCertificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer trucker certification %s\n", row.ID)
	return nil
}

func parseDoDeveloperTruckerCertificationsUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperTruckerCertificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	classification, _ := cmd.Flags().GetString("classification")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	defaultMultiplier, _ := cmd.Flags().GetString("default-multiplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	id := strings.TrimSpace(args[0])
	if id == "" {
		return doDeveloperTruckerCertificationsUpdateOptions{}, fmt.Errorf("developer trucker certification id is required")
	}

	return doDeveloperTruckerCertificationsUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                id,
		Classification:    classification,
		StartOn:           startOn,
		EndOn:             endOn,
		DefaultMultiplier: defaultMultiplier,
	}, nil
}
