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

type doPredictionSubjectsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	Name                 string
	Description          string
	Status               string
	Kind                 string
	Actual               float64
	ActualDueAt          string
	PredictionsDueAt     string
	DomainMin            float64
	DomainMax            float64
	AdditionalAttributes string
	ReferenceNumber      string
	Broker               string
	ParentType           string
	ParentID             string
	BusinessUnit         string
	PredictionConsensus  string
}

func newDoPredictionSubjectsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prediction subject",
		Long: `Create a prediction subject.

Required flags:
  --name          Prediction subject name (required)
  --parent-type   Parent type (brokers or projects) (required)
  --parent-id     Parent ID (required)

Optional flags:
  --status                 Status (active, complete, abandoned)
  --kind                   Prediction kind (lowest_losing_bid)
  --description            Description
  --actual                 Actual value
  --predictions-due-at     Predictions due at (ISO 8601)
  --actual-due-at          Actual due at (ISO 8601)
  --domain-min             Domain minimum value
  --domain-max             Domain maximum value
  --additional-attributes  Additional attributes JSON
  --reference-number       Reference number
  --broker                 Broker ID (create-only)
  --business-unit          Business unit ID
  --prediction-consensus   Prediction consensus ID`,
		Example: `  # Create a prediction subject for a broker
  xbe do prediction-subjects create --name "Lowest losing bid" --parent-type brokers --parent-id 123 --status active

  # Create with dates and domain
  xbe do prediction-subjects create \
    --name "Forecast" \
    --parent-type projects \
    --parent-id 456 \
    --predictions-due-at 2025-02-01 \
    --actual-due-at 2025-02-15 \
    --domain-min 100000 \
    --domain-max 200000

  # Create with additional attributes
  xbe do prediction-subjects create \
    --name "Forecast" \
    --parent-type brokers \
    --parent-id 123 \
    --additional-attributes '{"source":"cli"}' \
    --reference-number "REF-123"`,
		Args: cobra.NoArgs,
		RunE: runDoPredictionSubjectsCreate,
	}
	initDoPredictionSubjectsCreateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectsCmd.AddCommand(newDoPredictionSubjectsCreateCmd())
}

func initDoPredictionSubjectsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Prediction subject name (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("status", "", "Status (active, complete, abandoned)")
	cmd.Flags().String("kind", "", "Prediction kind (lowest_losing_bid)")
	cmd.Flags().Float64("actual", 0, "Actual value")
	cmd.Flags().String("predictions-due-at", "", "Predictions due at (ISO 8601)")
	cmd.Flags().String("actual-due-at", "", "Actual due at (ISO 8601)")
	cmd.Flags().Float64("domain-min", 0, "Domain minimum value")
	cmd.Flags().Float64("domain-max", 0, "Domain maximum value")
	cmd.Flags().String("additional-attributes", "", "Additional attributes JSON")
	cmd.Flags().String("reference-number", "", "Reference number")
	cmd.Flags().String("broker", "", "Broker ID (create-only)")
	cmd.Flags().String("parent-type", "", "Parent type (brokers or projects) (required)")
	cmd.Flags().String("parent-id", "", "Parent ID (required)")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("prediction-consensus", "", "Prediction consensus ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPredictionSubjectsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ParentType == "" {
		err := fmt.Errorf("--parent-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.ParentID == "" {
		err := fmt.Errorf("--parent-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("actual") {
		attributes["actual"] = opts.Actual
	}
	if opts.PredictionsDueAt != "" {
		attributes["predictions-due-at"] = opts.PredictionsDueAt
	}
	if opts.ActualDueAt != "" {
		attributes["actual-due-at"] = opts.ActualDueAt
	}
	if cmd.Flags().Changed("domain-min") {
		attributes["domain-min"] = opts.DomainMin
	}
	if cmd.Flags().Changed("domain-max") {
		attributes["domain-max"] = opts.DomainMax
	}
	if cmd.Flags().Changed("additional-attributes") {
		if strings.TrimSpace(opts.AdditionalAttributes) == "" {
			return fmt.Errorf("--additional-attributes requires valid JSON")
		}
		var parsed any
		if err := json.Unmarshal([]byte(opts.AdditionalAttributes), &parsed); err != nil {
			return fmt.Errorf("invalid additional-attributes JSON: %w", err)
		}
		attributes["additional-attributes"] = parsed
	}
	if opts.ReferenceNumber != "" {
		attributes["reference-number"] = opts.ReferenceNumber
	}

	relationships := map[string]any{
		"parent": map[string]any{
			"data": map[string]any{
				"type": opts.ParentType,
				"id":   opts.ParentID,
			},
		},
	}
	if opts.Broker != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if opts.BusinessUnit != "" {
		relationships["business-unit"] = map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		}
	}
	if opts.PredictionConsensus != "" {
		relationships["prediction-consensus"] = map[string]any{
			"data": map[string]any{
				"type": "predictions",
				"id":   opts.PredictionConsensus,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "prediction-subjects",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prediction-subjects", jsonBody)
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

	row := buildPredictionSubjectRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prediction subject %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoPredictionSubjectsCreateOptions(cmd *cobra.Command) (doPredictionSubjectsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	actual, _ := cmd.Flags().GetFloat64("actual")
	predictionsDueAt, _ := cmd.Flags().GetString("predictions-due-at")
	actualDueAt, _ := cmd.Flags().GetString("actual-due-at")
	domainMin, _ := cmd.Flags().GetFloat64("domain-min")
	domainMax, _ := cmd.Flags().GetFloat64("domain-max")
	additionalAttributes, _ := cmd.Flags().GetString("additional-attributes")
	referenceNumber, _ := cmd.Flags().GetString("reference-number")
	broker, _ := cmd.Flags().GetString("broker")
	parentType, _ := cmd.Flags().GetString("parent-type")
	parentID, _ := cmd.Flags().GetString("parent-id")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	predictionConsensus, _ := cmd.Flags().GetString("prediction-consensus")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		Name:                 name,
		Description:          description,
		Status:               status,
		Kind:                 kind,
		Actual:               actual,
		PredictionsDueAt:     predictionsDueAt,
		ActualDueAt:          actualDueAt,
		DomainMin:            domainMin,
		DomainMax:            domainMax,
		AdditionalAttributes: additionalAttributes,
		ReferenceNumber:      referenceNumber,
		Broker:               broker,
		ParentType:           parentType,
		ParentID:             parentID,
		BusinessUnit:         businessUnit,
		PredictionConsensus:  predictionConsensus,
	}, nil
}
