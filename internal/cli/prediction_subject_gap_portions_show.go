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

type predictionSubjectGapPortionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectGapPortionDetails struct {
	ID                     string `json:"id"`
	Name                   string `json:"name,omitempty"`
	Amount                 any    `json:"amount,omitempty"`
	Status                 string `json:"status,omitempty"`
	Description            string `json:"description,omitempty"`
	PredictionSubjectGapID string `json:"prediction_subject_gap_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
}

func newPredictionSubjectGapPortionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject gap portion details",
		Long: `Show the full details of a prediction subject gap portion.

Output Fields:
  ID          Prediction subject gap portion identifier
  NAME        Portion name
  AMOUNT      Portion amount
  STATUS      Portion status
  NOTE        Portion description
  GAP         Prediction subject gap ID
  CREATED BY  Creator user ID

Arguments:
  <id>  The prediction subject gap portion ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction subject gap portion
  xbe view prediction-subject-gap-portions show 123

  # Output as JSON
  xbe view prediction-subject-gap-portions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectGapPortionsShow,
	}
	initPredictionSubjectGapPortionsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectGapPortionsCmd.AddCommand(newPredictionSubjectGapPortionsShowCmd())
}

func initPredictionSubjectGapPortionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectGapPortionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionSubjectGapPortionsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject gap portion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subject-gap-portions]", "name,amount,description,status,prediction-subject-gap,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-gap-portions/"+id, query)
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

	details := buildPredictionSubjectGapPortionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectGapPortionDetails(cmd, details)
}

func parsePredictionSubjectGapPortionsShowOptions(cmd *cobra.Command) (predictionSubjectGapPortionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectGapPortionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectGapPortionDetails(resp jsonAPISingleResponse) predictionSubjectGapPortionDetails {
	attrs := resp.Data.Attributes
	details := predictionSubjectGapPortionDetails{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Amount:      attrs["amount"],
		Status:      stringAttr(attrs, "status"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resp.Data.Relationships["prediction-subject-gap"]; ok && rel.Data != nil {
		details.PredictionSubjectGapID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderPredictionSubjectGapPortionDetails(cmd *cobra.Command, details predictionSubjectGapPortionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Amount != nil {
		fmt.Fprintf(out, "Amount: %s\n", formatAnyValue(details.Amount))
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.PredictionSubjectGapID != "" {
		fmt.Fprintf(out, "Prediction Subject Gap: %s\n", details.PredictionSubjectGapID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}

	return nil
}
