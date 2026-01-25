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

type doLineupScenarioTrailersCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	LineupScenarioTruckerID string
	TrailerID               string
	LastAssignedOn          string
}

func newDoLineupScenarioTrailersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario trailer",
		Long: `Create a lineup scenario trailer.

Required:
  --lineup-scenario-trucker  Lineup scenario trucker ID
  --trailer                  Trailer ID

Optional:
  --last-assigned-on         Last assigned date (YYYY-MM-DD)`,
		Example: `  # Create a lineup scenario trailer
  xbe do lineup-scenario-trailers create --lineup-scenario-trucker 123 --trailer 456

  # Create with last assigned date
  xbe do lineup-scenario-trailers create --lineup-scenario-trucker 123 --trailer 456 --last-assigned-on 2024-01-01`,
		RunE: runDoLineupScenarioTrailersCreate,
	}
	initDoLineupScenarioTrailersCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTrailersCmd.AddCommand(newDoLineupScenarioTrailersCreateCmd())
}

func initDoLineupScenarioTrailersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario-trucker", "", "Lineup scenario trucker ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("last-assigned-on", "", "Last assigned date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("lineup-scenario-trucker")
	_ = cmd.MarkFlagRequired("trailer")
}

func runDoLineupScenarioTrailersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioTrailersCreateOptions(cmd)
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

	relationships := map[string]any{
		"lineup-scenario-trucker": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenario-truckers",
				"id":   opts.LineupScenarioTruckerID,
			},
		},
		"trailer": map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.TrailerID,
			},
		},
	}

	data := map[string]any{
		"type":          "lineup-scenario-trailers",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-trailers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario trailer %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupScenarioTrailersCreateOptions(cmd *cobra.Command) (doLineupScenarioTrailersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenarioTruckerID, _ := cmd.Flags().GetString("lineup-scenario-trucker")
	trailerID, _ := cmd.Flags().GetString("trailer")
	lastAssignedOn, _ := cmd.Flags().GetString("last-assigned-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTrailersCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		LineupScenarioTruckerID: lineupScenarioTruckerID,
		TrailerID:               trailerID,
		LastAssignedOn:          lastAssignedOn,
	}, nil
}
