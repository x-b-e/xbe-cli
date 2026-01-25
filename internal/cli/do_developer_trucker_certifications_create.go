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

type doDeveloperTruckerCertificationsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	Developer         string
	Trucker           string
	Classification    string
	StartOn           string
	EndOn             string
	DefaultMultiplier string
}

func newDoDeveloperTruckerCertificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a developer trucker certification",
		Long: `Create a developer trucker certification.

Required flags:
  --developer       Developer ID (required)
  --trucker         Trucker ID (required)
  --classification  Classification ID (required)

Optional flags:
  --start-on           Start date (YYYY-MM-DD)
  --end-on             End date (YYYY-MM-DD)
  --default-multiplier Default multiplier (numeric)`,
		Example: `  # Create a developer trucker certification
  xbe do developer-trucker-certifications create --developer 123 --trucker 456 --classification 789 --start-on 2024-01-01 --end-on 2024-12-31 --default-multiplier 1.2

  # Create without dates
  xbe do developer-trucker-certifications create --developer 123 --trucker 456 --classification 789

  # Output as JSON
  xbe do developer-trucker-certifications create --developer 123 --trucker 456 --classification 789 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDeveloperTruckerCertificationsCreate,
	}
	initDoDeveloperTruckerCertificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationsCmd.AddCommand(newDoDeveloperTruckerCertificationsCreateCmd())
}

func initDoDeveloperTruckerCertificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("developer", "", "Developer ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("classification", "", "Classification ID (required)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("default-multiplier", "", "Default multiplier (numeric)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeveloperTruckerCertificationsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Developer) == "" {
		err := fmt.Errorf("--developer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Trucker) == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Classification) == "" {
		err := fmt.Errorf("--classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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

	relationships := map[string]any{
		"developer": map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.Developer,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
		"classification": map[string]any{
			"data": map[string]any{
				"type": "developer-trucker-certification-classifications",
				"id":   opts.Classification,
			},
		},
	}

	data := map[string]any{
		"type":          "developer-trucker-certifications",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/developer-trucker-certifications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer trucker certification %s\n", row.ID)
	return nil
}

func parseDoDeveloperTruckerCertificationsCreateOptions(cmd *cobra.Command) (doDeveloperTruckerCertificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	developer, _ := cmd.Flags().GetString("developer")
	trucker, _ := cmd.Flags().GetString("trucker")
	classification, _ := cmd.Flags().GetString("classification")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	defaultMultiplier, _ := cmd.Flags().GetString("default-multiplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		Developer:         developer,
		Trucker:           trucker,
		Classification:    classification,
		StartOn:           startOn,
		EndOn:             endOn,
		DefaultMultiplier: defaultMultiplier,
	}, nil
}
