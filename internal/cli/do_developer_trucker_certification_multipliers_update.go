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

type doDeveloperTruckerCertificationMultipliersUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	Multiplier float64
	Trailer    string
}

func newDoDeveloperTruckerCertificationMultipliersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a developer trucker certification multiplier",
		Long: `Update a developer trucker certification multiplier.

At least one field is required.

Updatable fields:
  --multiplier  Multiplier value between 0 and 1
  --trailer     Trailer ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update multiplier
  xbe do developer-trucker-certification-multipliers update 123 --multiplier 0.9

  # Update trailer
  xbe do developer-trucker-certification-multipliers update 123 --trailer 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDeveloperTruckerCertificationMultipliersUpdate,
	}
	initDoDeveloperTruckerCertificationMultipliersUpdateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperTruckerCertificationMultipliersCmd.AddCommand(newDoDeveloperTruckerCertificationMultipliersUpdateCmd())
}

func initDoDeveloperTruckerCertificationMultipliersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Float64("multiplier", 0, "Multiplier value between 0 and 1")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperTruckerCertificationMultipliersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperTruckerCertificationMultipliersUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("multiplier") {
		attributes["multiplier"] = opts.Multiplier
	}

	relationships := map[string]any{}
	if opts.Trailer != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "developer-trucker-certification-multipliers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/developer-trucker-certification-multipliers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated developer trucker certification multiplier %s\n", details.ID)
	return nil
}

func parseDoDeveloperTruckerCertificationMultipliersUpdateOptions(cmd *cobra.Command, args []string) (doDeveloperTruckerCertificationMultipliersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	multiplier, _ := cmd.Flags().GetFloat64("multiplier")
	trailer, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperTruckerCertificationMultipliersUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		Multiplier: multiplier,
		Trailer:    trailer,
	}, nil
}
