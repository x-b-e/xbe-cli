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

type doOneStepGpsVehiclesUpdateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	ID                                       string
	Trailer                                  string
	Tractor                                  string
	SkipTrailerIsNotAlreadyMatchedValidation string
	SkipTractorIsNotAlreadyMatchedValidation string
}

func newDoOneStepGpsVehiclesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update One Step GPS vehicle assignments",
		Long: `Update One Step GPS vehicle assignments.

Optional attributes:
  --skip-trailer-is-not-already-matched-validation  Skip trailer match validation (true/false)
  --skip-tractor-is-not-already-matched-validation  Skip tractor match validation (true/false)

Optional relationships:
  --trailer    Trailer ID
  --tractor    Tractor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update trailer assignment
  xbe do one-step-gps-vehicles update 123 --trailer 456

  # Skip trailer match validation
  xbe do one-step-gps-vehicles update 123 --skip-trailer-is-not-already-matched-validation true`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOneStepGpsVehiclesUpdate,
	}
	initDoOneStepGpsVehiclesUpdateFlags(cmd)
	return cmd
}

func init() {
	doOneStepGpsVehiclesCmd.AddCommand(newDoOneStepGpsVehiclesUpdateCmd())
}

func initDoOneStepGpsVehiclesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("skip-trailer-is-not-already-matched-validation", "", "Skip trailer match validation (true/false)")
	cmd.Flags().String("skip-tractor-is-not-already-matched-validation", "", "Skip tractor match validation (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOneStepGpsVehiclesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOneStepGpsVehiclesUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	setBoolAttrIfPresent(attributes, "skip-trailer-is-not-already-matched-validation", opts.SkipTrailerIsNotAlreadyMatchedValidation)
	setBoolAttrIfPresent(attributes, "skip-tractor-is-not-already-matched-validation", opts.SkipTractorIsNotAlreadyMatchedValidation)

	if cmd.Flags().Changed("trailer") {
		if strings.TrimSpace(opts.Trailer) == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]string{
					"type": "trailers",
					"id":   opts.Trailer,
				},
			}
		}
	}

	if cmd.Flags().Changed("tractor") {
		if strings.TrimSpace(opts.Tractor) == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]string{
					"type": "tractors",
					"id":   opts.Tractor,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "one-step-gps-vehicles",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/one-step-gps-vehicles/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := oneStepGpsVehicleRow{
			ID:                                       resp.Data.ID,
			VehicleID:                                stringAttr(resp.Data.Attributes, "vehicle-id"),
			VehicleNumber:                            stringAttr(resp.Data.Attributes, "vehicle-number"),
			IntegrationIdentifier:                    stringAttr(resp.Data.Attributes, "integration-identifier"),
			TrailerSetAt:                             formatDateTime(stringAttr(resp.Data.Attributes, "trailer-set-at")),
			TractorSetAt:                             formatDateTime(stringAttr(resp.Data.Attributes, "tractor-set-at")),
			SkipTrailerIsNotAlreadyMatchedValidation: boolAttr(resp.Data.Attributes, "skip-trailer-is-not-already-matched-validation"),
			SkipTractorIsNotAlreadyMatchedValidation: boolAttr(resp.Data.Attributes, "skip-tractor-is-not-already-matched-validation"),
			BrokerID:                                 relationshipIDFromMap(resp.Data.Relationships, "broker"),
			TruckerID:                                relationshipIDFromMap(resp.Data.Relationships, "trucker"),
			TrailerID:                                relationshipIDFromMap(resp.Data.Relationships, "trailer"),
			TractorID:                                relationshipIDFromMap(resp.Data.Relationships, "tractor"),
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated One Step GPS vehicle %s\n", resp.Data.ID)
	return nil
}

func parseDoOneStepGpsVehiclesUpdateOptions(cmd *cobra.Command, args []string) (doOneStepGpsVehiclesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	skipTrailerValidation, _ := cmd.Flags().GetString("skip-trailer-is-not-already-matched-validation")
	skipTractorValidation, _ := cmd.Flags().GetString("skip-tractor-is-not-already-matched-validation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOneStepGpsVehiclesUpdateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		ID:                                       args[0],
		Trailer:                                  trailer,
		Tractor:                                  tractor,
		SkipTrailerIsNotAlreadyMatchedValidation: skipTrailerValidation,
		SkipTractorIsNotAlreadyMatchedValidation: skipTractorValidation,
	}, nil
}
