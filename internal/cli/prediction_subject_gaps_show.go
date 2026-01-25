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

type predictionSubjectGapsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type predictionSubjectGapPortion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
}

type predictionSubjectGapDetails struct {
	ID                               string                        `json:"id"`
	GapType                          string                        `json:"gap_type,omitempty"`
	Status                           string                        `json:"status,omitempty"`
	PrimaryAmount                    *float64                      `json:"primary_amount,omitempty"`
	SecondaryAmount                  *float64                      `json:"secondary_amount,omitempty"`
	GapAmount                        *float64                      `json:"gap_amount,omitempty"`
	ExplainedGapAmount               *float64                      `json:"explained_gap_amount,omitempty"`
	PredictionSubjectID              string                        `json:"prediction_subject_id,omitempty"`
	PredictionSubjectName            string                        `json:"prediction_subject_name,omitempty"`
	PredictionSubjectReferenceNumber string                        `json:"prediction_subject_reference_number,omitempty"`
	PortionIDs                       []string                      `json:"portion_ids,omitempty"`
	Portions                         []predictionSubjectGapPortion `json:"portions,omitempty"`
}

func newPredictionSubjectGapsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prediction subject gap details",
		Long: `Show the full details of a prediction subject gap.

Output Fields:
  ID           Prediction subject gap identifier
  Type         Gap type
  Status       Gap status
  Primary      Primary amount
  Secondary    Secondary amount
  Gap          Gap amount
  Explained    Explained gap amount
  Subject      Prediction subject name or ID
  Portions     Related gap portion details (if available)

Arguments:
  <id>    The prediction subject gap ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a prediction subject gap
  xbe view prediction-subject-gaps show 123

  # JSON output
  xbe view prediction-subject-gaps show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPredictionSubjectGapsShow,
	}
	initPredictionSubjectGapsShowFlags(cmd)
	return cmd
}

func init() {
	predictionSubjectGapsCmd.AddCommand(newPredictionSubjectGapsShowCmd())
}

func initPredictionSubjectGapsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionSubjectGapsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePredictionSubjectGapsShowOptions(cmd)
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
		return fmt.Errorf("prediction subject gap id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prediction-subject-gaps]", "gap-type,status,primary-amount,secondary-amount,gap-amount,explained-gap-amount,prediction-subject,portions")
	query.Set("include", "prediction-subject,portions")
	query.Set("fields[prediction-subjects]", "name,reference-number")
	query.Set("fields[prediction-subject-gap-portions]", "name,amount,description,status")

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-subject-gaps/"+id, query)
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

	details := buildPredictionSubjectGapDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPredictionSubjectGapDetails(cmd, details)
}

func parsePredictionSubjectGapsShowOptions(cmd *cobra.Command) (predictionSubjectGapsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionSubjectGapsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPredictionSubjectGapDetails(resp jsonAPISingleResponse) predictionSubjectGapDetails {
	attrs := resp.Data.Attributes
	details := predictionSubjectGapDetails{ID: resp.Data.ID}

	details.GapType = stringAttr(attrs, "gap-type")
	details.Status = stringAttr(attrs, "status")
	if amount, ok := floatAttrValue(attrs, "primary-amount"); ok {
		details.PrimaryAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "secondary-amount"); ok {
		details.SecondaryAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "gap-amount"); ok {
		details.GapAmount = &amount
	}
	if amount, ok := floatAttrValue(attrs, "explained-gap-amount"); ok {
		details.ExplainedGapAmount = &amount
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["prediction-subject"]; ok && rel.Data != nil {
		details.PredictionSubjectID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.PredictionSubjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			details.PredictionSubjectReferenceNumber = strings.TrimSpace(stringAttr(inc.Attributes, "reference-number"))
		}
	}

	if rel, ok := resp.Data.Relationships["portions"]; ok {
		details.PortionIDs = relationshipIDStrings(rel)
		for _, ref := range relationshipIDs(rel) {
			portion := predictionSubjectGapPortion{ID: ref.ID}
			if inc, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
				portion.Name = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				portion.Status = stringAttr(inc.Attributes, "status")
				portion.Description = strings.TrimSpace(stringAttr(inc.Attributes, "description"))
				if amount, ok := floatAttrValue(inc.Attributes, "amount"); ok {
					portion.Amount = &amount
				}
			}
			details.Portions = append(details.Portions, portion)
		}
	}

	return details
}

func renderPredictionSubjectGapDetails(cmd *cobra.Command, details predictionSubjectGapDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.GapType != "" {
		fmt.Fprintf(out, "Gap Type: %s\n", details.GapType)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.PrimaryAmount != nil {
		fmt.Fprintf(out, "Primary Amount: %s\n", formatPredictionSubjectGapAmount(details.PrimaryAmount))
	}
	if details.SecondaryAmount != nil {
		fmt.Fprintf(out, "Secondary Amount: %s\n", formatPredictionSubjectGapAmount(details.SecondaryAmount))
	}
	if details.GapAmount != nil {
		fmt.Fprintf(out, "Gap Amount: %s\n", formatPredictionSubjectGapAmount(details.GapAmount))
	}
	if details.ExplainedGapAmount != nil {
		fmt.Fprintf(out, "Explained Gap Amount: %s\n", formatPredictionSubjectGapAmount(details.ExplainedGapAmount))
	}

	subjectLabel := firstNonEmpty(details.PredictionSubjectName, details.PredictionSubjectReferenceNumber)
	if subjectLabel != "" {
		fmt.Fprintf(out, "Prediction Subject: %s\n", subjectLabel)
	}
	if details.PredictionSubjectID != "" {
		fmt.Fprintf(out, "Prediction Subject ID: %s\n", details.PredictionSubjectID)
	}

	if len(details.Portions) > 0 {
		fmt.Fprintf(out, "Portions (%d):\n", len(details.Portions))
		for _, portion := range details.Portions {
			label := firstNonEmpty(portion.Name, portion.ID)
			parts := []string{}
			if portion.Amount != nil {
				parts = append(parts, "amount "+formatPredictionSubjectGapAmount(portion.Amount))
			}
			if portion.Status != "" {
				parts = append(parts, "status "+portion.Status)
			}
			if portion.Description != "" {
				parts = append(parts, portion.Description)
			}
			if len(parts) > 0 {
				fmt.Fprintf(out, "  - %s (%s)\n", label, strings.Join(parts, ", "))
			} else {
				fmt.Fprintf(out, "  - %s\n", label)
			}
		}
	} else if len(details.PortionIDs) > 0 {
		fmt.Fprintf(out, "Portion IDs: %s\n", strings.Join(details.PortionIDs, ", "))
	}

	return nil
}
