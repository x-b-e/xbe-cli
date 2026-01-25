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

type doDigitalFleetTrucksUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	TractorID string
	TrailerID string
}

func newDoDigitalFleetTrucksUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update digital fleet truck assignments",
		Long: `Update digital fleet truck trailer or tractor assignments.

Optional flags:
  --tractor  Tractor ID (use empty string to clear)
  --trailer  Trailer ID (use empty string to clear)`,
		Example: `  # Update tractor assignment
  xbe do digital-fleet-trucks update 123 --tractor 456

  # Update trailer assignment
  xbe do digital-fleet-trucks update 123 --trailer 789

  # Clear trailer assignment
  xbe do digital-fleet-trucks update 123 --trailer ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDigitalFleetTrucksUpdate,
	}
	initDoDigitalFleetTrucksUpdateFlags(cmd)
	return cmd
}

func init() {
	doDigitalFleetTrucksCmd.AddCommand(newDoDigitalFleetTrucksUpdateCmd())
}

func initDoDigitalFleetTrucksUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDigitalFleetTrucksUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDigitalFleetTrucksUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	var hasChanges bool

	if cmd.Flags().Changed("tractor") {
		if strings.TrimSpace(opts.TractorID) == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]any{
					"type": "tractors",
					"id":   opts.TractorID,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("trailer") {
		if strings.TrimSpace(opts.TrailerID) == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.TrailerID,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify --tractor or --trailer")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "digital-fleet-trucks",
		"id":   opts.ID,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/digital-fleet-trucks/"+opts.ID, jsonBody)
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

	row := digitalFleetTruckRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated digital fleet truck %s\n", row.ID)
	return nil
}

func parseDoDigitalFleetTrucksUpdateOptions(cmd *cobra.Command, args []string) (doDigitalFleetTrucksUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tractorID, _ := cmd.Flags().GetString("tractor")
	trailerID, _ := cmd.Flags().GetString("trailer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDigitalFleetTrucksUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		TractorID: tractorID,
		TrailerID: trailerID,
	}, nil
}
