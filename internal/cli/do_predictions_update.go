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

type doPredictionsUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	Status                  string
	ProbabilityDistribution string
}

func newDoPredictionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction",
		Long: `Update a prediction.

Optional flags:
  --status                  Status (draft, submitted, abandoned)
  --probability-distribution  JSON probability distribution payload (includes class_name)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update prediction status
  xbe do predictions update 123 --status submitted

  # Update probability distribution
  xbe do predictions update 123 \
    --probability-distribution '{"class_name":"TriangularDistribution","minimum":105,"mode":125,"maximum":145}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionsUpdate,
	}
	initDoPredictionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionsCmd.AddCommand(newDoPredictionsUpdateCmd())
}

func initDoPredictionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Status (draft, submitted, abandoned)")
	cmd.Flags().String("probability-distribution", "", "JSON probability distribution payload")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("probability-distribution") {
		if opts.ProbabilityDistribution != "" {
			var distribution any
			if err := json.Unmarshal([]byte(opts.ProbabilityDistribution), &distribution); err != nil {
				err = fmt.Errorf("invalid probability-distribution JSON: %w", err)
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["probability-distribution"] = distribution
		} else {
			attributes["probability-distribution"] = nil
		}
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "predictions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/predictions/"+opts.ID, jsonBody)
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

	row := predictionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	message := fmt.Sprintf("Updated prediction %s", row.ID)
	parts := []string{}
	if row.Status != "" {
		parts = append(parts, "status "+row.Status)
	}
	if row.PredictionSubjectID != "" {
		parts = append(parts, "subject "+row.PredictionSubjectID)
	}
	if len(parts) > 0 {
		message = fmt.Sprintf("%s (%s)", message, strings.Join(parts, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout(), message)
	return nil
}

func parseDoPredictionsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	probabilityDistribution, _ := cmd.Flags().GetString("probability-distribution")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionsUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		Status:                  status,
		ProbabilityDistribution: probabilityDistribution,
	}, nil
}
