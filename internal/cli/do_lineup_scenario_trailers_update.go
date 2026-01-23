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

type doLineupScenarioTrailersUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	LastAssignedOn string
}

func newDoLineupScenarioTrailersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup scenario trailer",
		Long: `Update a lineup scenario trailer.

Optional:
  --last-assigned-on  Last assigned date (YYYY-MM-DD)`,
		Example: `  # Update last assigned date
  xbe do lineup-scenario-trailers update 123 --last-assigned-on 2024-01-01`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenarioTrailersUpdate,
	}
	initDoLineupScenarioTrailersUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTrailersCmd.AddCommand(newDoLineupScenarioTrailersUpdateCmd())
}

func initDoLineupScenarioTrailersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("last-assigned-on", "", "Last assigned date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioTrailersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenarioTrailersUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("last-assigned-on") {
		attributes["last-assigned-on"] = opts.LastAssignedOn
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lineup-scenario-trailers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-scenario-trailers/"+opts.ID, jsonBody)
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
		row := lineupScenarioTrailerRow{
			ID:             resp.Data.ID,
			LastAssignedOn: formatDate(stringAttr(resp.Data.Attributes, "last-assigned-on")),
		}
		if rel, ok := resp.Data.Relationships["lineup-scenario-trucker"]; ok && rel.Data != nil {
			row.LineupScenarioTruckerID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup scenario trailer %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupScenarioTrailersUpdateOptions(cmd *cobra.Command, args []string) (doLineupScenarioTrailersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lastAssignedOn, _ := cmd.Flags().GetString("last-assigned-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTrailersUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		LastAssignedOn: lastAssignedOn,
	}, nil
}
