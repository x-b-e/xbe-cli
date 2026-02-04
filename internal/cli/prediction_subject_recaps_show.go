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

type predictionSubjectRecapsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectRecapDetails struct {
	ID                  string `json:"id"`
	PredictionSubjectID string `json:"prediction_subject_id,omitempty"`
	Markdown            string `json:"markdown,omitempty"`
	Data                any    `json:"data,omitempty"`
}

func newPredictionSubjectRecapsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject recap details",
		Long: `Show the full details of a prediction subject recap.

Output Fields:
  ID       Prediction subject recap identifier
  SUBJECT  Prediction subject ID
  MARKDOWN Recap markdown content
  DATA     Recap data payload

Arguments:
  <id>  The prediction subject recap ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction subject recap
  xbe view prediction-subject-recaps show 123

  # Output as JSON
  xbe view prediction-subject-recaps show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectRecapsShow,
	}
	initPredictionSubjectRecapsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectRecapsCmd.AddCommand(newPredictionSubjectRecapsShowCmd())
}

func initPredictionSubjectRecapsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectRecapsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parsePredictionSubjectRecapsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject recap id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subject-recaps]", "prediction-subject,data,markdown")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-recaps/"+id, query)
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

	details := buildPredictionSubjectRecapDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectRecapDetails(cmd, details)
}

func parsePredictionSubjectRecapsShowOptions(cmd *cobra.Command) (predictionSubjectRecapsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectRecapsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectRecapDetails(resp jsonAPISingleResponse) predictionSubjectRecapDetails {
	attrs := resp.Data.Attributes
	details := predictionSubjectRecapDetails{
		ID:       resp.Data.ID,
		Markdown: stringAttr(attrs, "markdown"),
		Data:     attrs["data"],
	}

	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
	}

	return details
}

func renderPredictionSubjectRecapDetails(cmd *cobra.Command, details predictionSubjectRecapDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", details.PredictionSubjectID)
	}
	if details.Markdown != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Markdown:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Markdown)
	}
	if details.Data != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Data:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatAnyValue(details.Data))
	}

	return nil
}
