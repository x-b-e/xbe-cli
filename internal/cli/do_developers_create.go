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

type doDevelopersCreateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	Name                            string
	Broker                          string
	WeigherSealLabel                string
	IsPrevailingWageExplicit        bool
	IsCertificationRequiredExplicit bool
}

func newDoDevelopersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new developer",
		Long: `Create a new developer.

Required flags:
  --name      The developer name (required)
  --broker    The broker ID (required)

Optional flags:
  --weigher-seal-label                  Weigher seal label
  --is-prevailing-wage-explicit         Explicitly set prevailing wage requirement
  --is-certification-required-explicit  Explicitly set certification requirement`,
		Example: `  # Create a developer
  xbe do developers create --name "Acme Development" --broker 123

  # Create with all options
  xbe do developers create --name "Acme Development" --broker 123 \
    --weigher-seal-label "ACME" \
    --is-prevailing-wage-explicit \
    --is-certification-required-explicit

  # Get JSON output
  xbe do developers create --name "New Developer" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDevelopersCreate,
	}
	initDoDevelopersCreateFlags(cmd)
	return cmd
}

func init() {
	doDevelopersCmd.AddCommand(newDoDevelopersCreateCmd())
}

func initDoDevelopersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Developer name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("weigher-seal-label", "", "Weigher seal label")
	cmd.Flags().Bool("is-prevailing-wage-explicit", false, "Explicitly set prevailing wage requirement")
	cmd.Flags().Bool("is-certification-required-explicit", false, "Explicitly set certification requirement")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDevelopersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDevelopersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require broker
	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name": opts.Name,
	}
	if opts.WeigherSealLabel != "" {
		attributes["weigher-seal-label"] = opts.WeigherSealLabel
	}
	if cmd.Flags().Changed("is-prevailing-wage-explicit") {
		attributes["is-prevailing-wage-explicit"] = opts.IsPrevailingWageExplicit
	}
	if cmd.Flags().Changed("is-certification-required-explicit") {
		attributes["is-certification-required-explicit"] = opts.IsCertificationRequiredExplicit
	}

	// Build relationships
	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]string{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "developers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/developers", jsonBody)
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

	row := developerRow{
		ID:   resp.Data.ID,
		Name: stringAttr(resp.Data.Attributes, "name"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDevelopersCreateOptions(cmd *cobra.Command) (doDevelopersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	weigherSealLabel, _ := cmd.Flags().GetString("weigher-seal-label")
	isPrevailingWageExplicit, _ := cmd.Flags().GetBool("is-prevailing-wage-explicit")
	isCertificationRequiredExplicit, _ := cmd.Flags().GetBool("is-certification-required-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDevelopersCreateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		Name:                            name,
		Broker:                          broker,
		WeigherSealLabel:                weigherSealLabel,
		IsPrevailingWageExplicit:        isPrevailingWageExplicit,
		IsCertificationRequiredExplicit: isCertificationRequiredExplicit,
	}, nil
}
