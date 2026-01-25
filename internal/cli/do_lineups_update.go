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

type doLineupsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	Name       string
	StartAtMin string
	StartAtMax string
}

func newDoLineupsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup",
		Long: `Update a lineup.

Optional flags:
  --name          Lineup name
  --start-at-min  Earliest start time (ISO 8601)
  --start-at-max  Latest start time (ISO 8601)`,
		Example: `  # Update lineup name
  xbe do lineups update 123 --name "Afternoon"

  # Update lineup time window
  xbe do lineups update 123 --start-at-min 2026-01-02T06:00:00Z --start-at-max 2026-01-02T18:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupsUpdate,
	}
	initDoLineupsUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupsCmd.AddCommand(newDoLineupsUpdateCmd())
}

func initDoLineupsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Lineup name")
	cmd.Flags().String("start-at-min", "", "Earliest start time (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Latest start time (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("start-at-min") {
		attributes["start-at-min"] = opts.StartAtMin
	}
	if cmd.Flags().Changed("start-at-max") {
		attributes["start-at-max"] = opts.StartAtMax
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lineups",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/lineups/"+opts.ID, jsonBody)
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

	details := buildLineupDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup %s\n", details.ID)
	return nil
}

func parseDoLineupsUpdateOptions(cmd *cobra.Command, args []string) (doLineupsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		Name:       name,
		StartAtMin: startAtMin,
		StartAtMax: startAtMax,
	}, nil
}
