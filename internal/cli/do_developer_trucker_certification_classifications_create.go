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

type doDeveloperTruckerCertificationClassificationsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	DeveloperID string
}

func newDoDeveloperTruckerCertificationClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new developer trucker certification classification",
		Long: `Create a new developer trucker certification classification.

Required flags:
  --name         The classification name (required)
  --developer    Developer ID (required)`,
		Example: `  # Create a developer trucker certification classification
  xbe do developer-trucker-certification-classifications create --name "Safety Training" --developer 123`,
		Args: cobra.NoArgs,
		RunE: runDoDeveloperTruckerCertificationClassificationsCreate,
	}
	initDoDeveloperTruckerCertificationClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationClassificationsCmd.AddCommand(newDoDeveloperTruckerCertificationClassificationsCreateCmd())
}

func initDoDeveloperTruckerCertificationClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name (required)")
	cmd.Flags().String("developer", "", "Developer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperTruckerCertificationClassificationsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.DeveloperID == "" {
		err := fmt.Errorf("--developer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	relationships := map[string]any{
		"developer": map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.DeveloperID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "developer-trucker-certification-classifications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/developer-trucker-certification-classifications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer trucker certification classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDeveloperTruckerCertificationClassificationsCreateOptions(cmd *cobra.Command) (doDeveloperTruckerCertificationClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	developerID, _ := cmd.Flags().GetString("developer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationClassificationsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		DeveloperID: developerID,
	}, nil
}

func buildDeveloperTruckerCertificationClassificationRowFromSingle(resp jsonAPISingleResponse) developerTruckerCertificationClassificationRow {
	attrs := resp.Data.Attributes

	row := developerTruckerCertificationClassificationRow{
		ID:   resp.Data.ID,
		Name: stringAttr(attrs, "name"),
	}

	if rel, ok := resp.Data.Relationships["developer"]; ok && rel.Data != nil {
		row.DeveloperID = rel.Data.ID
	}

	return row
}
