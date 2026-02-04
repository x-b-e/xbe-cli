package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type predictionAgentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionAgentDetails struct {
	ID                  string `json:"id"`
	PredictionSubjectID string `json:"prediction_subject_id,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	PredictionID        string `json:"prediction_id,omitempty"`
	CustomInstructions  string `json:"custom_instructions,omitempty"`
	Messages            any    `json:"messages,omitempty"`
}

func newPredictionAgentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction agent details",
		Long: `Show the full details of a prediction agent.

Output Fields:
  ID                    Prediction agent identifier
  Prediction Subject    Prediction subject ID
  Created By            Creator user ID
  Prediction            Prediction ID (if present)
  Custom Instructions   Custom instructions (if present)
  Messages              Messages JSON (if present)

Arguments:
  <id>  The prediction agent ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show prediction agent details
  xbe view prediction-agents show 123

  # Output as JSON
  xbe view prediction-agents show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionAgentsShow,
	}
	initPredictionAgentsShowFlags(cmd)
	return cmd
}

func init() {
	predictionAgentsCmd.AddCommand(newPredictionAgentsShowCmd())
}

func initPredictionAgentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionAgentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePredictionAgentsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prediction agent id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-agents]", "prediction-subject,created-by,prediction,custom-instructions,messages")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-agents/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPredictionAgentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionAgentDetails(cmd, details)
}

func parsePredictionAgentsShowOptions(cmd *cobra.Command) (predictionAgentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionAgentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionAgentDetails(resp jsonAPISingleResponse) predictionAgentDetails {
	attrs := resp.Data.Attributes
	details := predictionAgentDetails{
		ID:                 resp.Data.ID,
		CustomInstructions: stringAttr(attrs, "custom-instructions"),
		Messages:           anyAttr(attrs, "messages"),
	}

	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["prediction"]; ok && rel.Data != nil {
		details.PredictionID = rel.Data.ID
	}

	return details
}

func renderPredictionAgentDetails(cmd *cobra.Command, details predictionAgentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", details.PredictionSubjectID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.PredictionID != "" {
		fmt.Fprintf(out, "Prediction: %s\n", details.PredictionID)
	}
	if details.CustomInstructions != "" {
		fmt.Fprintf(out, "Custom Instructions: %s\n", details.CustomInstructions)
	}
	if details.Messages != nil {
		fmt.Fprintf(out, "Messages: %d\n", countConstraintItems(details.Messages))
		if formatted := formatAnyJSON(details.Messages); formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Messages JSON:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
