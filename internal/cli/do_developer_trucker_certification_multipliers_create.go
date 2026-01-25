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

type doDeveloperTruckerCertificationMultipliersCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	DeveloperTruckerCertification string
	Trailer                       string
	Multiplier                    float64
}

func newDoDeveloperTruckerCertificationMultipliersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a developer trucker certification multiplier",
		Long: `Create a developer trucker certification multiplier.

Required flags:
  --developer-trucker-certification  Developer trucker certification ID (required)
  --trailer                          Trailer ID (required)
  --multiplier                       Multiplier value between 0 and 1 (required)

Notes:
  The trailer must belong to the same trucker as the certification, and
  each trailer can only appear once per certification.`,
		Example: `  # Create a multiplier for a trailer
  xbe do developer-trucker-certification-multipliers create \
    --developer-trucker-certification 123 \
    --trailer 456 \
    --multiplier 0.85

  # JSON output
  xbe do developer-trucker-certification-multipliers create \
    --developer-trucker-certification 123 \
    --trailer 456 \
    --multiplier 0.85 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoDeveloperTruckerCertificationMultipliersCreate,
	}
	initDoDeveloperTruckerCertificationMultipliersCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationMultipliersCmd.AddCommand(newDoDeveloperTruckerCertificationMultipliersCreateCmd())
}

func initDoDeveloperTruckerCertificationMultipliersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("developer-trucker-certification", "", "Developer trucker certification ID (required)")
	cmd.Flags().String("trailer", "", "Trailer ID (required)")
	cmd.Flags().Float64("multiplier", 0, "Multiplier value between 0 and 1 (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationMultipliersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeveloperTruckerCertificationMultipliersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.DeveloperTruckerCertification) == "" {
		err := fmt.Errorf("--developer-trucker-certification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Trailer) == "" {
		err := fmt.Errorf("--trailer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if !cmd.Flags().Changed("multiplier") {
		err := fmt.Errorf("--multiplier is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"multiplier": opts.Multiplier,
	}

	relationships := map[string]any{
		"developer-trucker-certification": map[string]any{
			"data": map[string]any{
				"type": "developer-trucker-certifications",
				"id":   opts.DeveloperTruckerCertification,
			},
		},
		"trailer": map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "developer-trucker-certification-multipliers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/developer-trucker-certification-multipliers", jsonBody)
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

	details := buildDeveloperTruckerCertificationMultiplierDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer trucker certification multiplier %s\n", details.ID)
	return nil
}

func parseDoDeveloperTruckerCertificationMultipliersCreateOptions(cmd *cobra.Command) (doDeveloperTruckerCertificationMultipliersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	developerTruckerCertification, _ := cmd.Flags().GetString("developer-trucker-certification")
	trailer, _ := cmd.Flags().GetString("trailer")
	multiplier, _ := cmd.Flags().GetFloat64("multiplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationMultipliersCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		DeveloperTruckerCertification: developerTruckerCertification,
		Trailer:                       trailer,
		Multiplier:                    multiplier,
	}, nil
}
