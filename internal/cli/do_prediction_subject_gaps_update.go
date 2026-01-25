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

type doPredictionSubjectGapsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	GapType string
	Status  string
}

func newDoPredictionSubjectGapsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction subject gap",
		Long: `Update a prediction subject gap.

Provide the gap ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Updatable fields:
  --gap-type  Gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus)
  --status    Status (pending, approved)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update gap status
  xbe do prediction-subject-gaps update 123 --status approved

  # Update gap type
  xbe do prediction-subject-gaps update 123 --gap-type actual_vs_consensus

  # Get JSON output
  xbe do prediction-subject-gaps update 123 --status approved --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectGapsUpdate,
	}
	initDoPredictionSubjectGapsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapsCmd.AddCommand(newDoPredictionSubjectGapsUpdateCmd())
}

func initDoPredictionSubjectGapsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("gap-type", "", "Gap type (actual_vs_walk_away, actual_vs_consensus, walk_away_vs_consensus)")
	cmd.Flags().String("status", "", "Status (pending, approved)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectGapsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectGapsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("gap-type") {
		attributes["gap-type"] = opts.GapType
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --gap-type or --status")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-gaps",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-subject-gaps/"+opts.ID, jsonBody)
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

	row := predictionSubjectGapRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	message := fmt.Sprintf("Updated prediction subject gap %s", row.ID)
	details := []string{}
	if row.GapType != "" {
		details = append(details, "type "+row.GapType)
	}
	if row.Status != "" {
		details = append(details, "status "+row.Status)
	}
	if row.GapAmount != nil {
		details = append(details, "gap "+formatPredictionSubjectGapAmount(row.GapAmount))
	}
	if len(details) > 0 {
		message = fmt.Sprintf("%s (%s)", message, strings.Join(details, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionSubjectGapsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionSubjectGapsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	gapType, _ := cmd.Flags().GetString("gap-type")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		GapType: gapType,
		Status:  status,
	}, nil
}
