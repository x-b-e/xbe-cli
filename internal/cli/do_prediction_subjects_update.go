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

type doPredictionSubjectsUpdateOptions struct {
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
	ParentType           string
	ParentID             string
	BusinessUnit         string
	PredictionConsensus  string
}

func newDoPredictionSubjectsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction subject",
		Long: `Update a prediction subject.

Optional flags:
  --name                  Prediction subject name
  --description           Description
  --status                Status (active, complete, abandoned)
  --kind                  Prediction kind (lowest_losing_bid)
  --actual                Actual value
  --predictions-due-at    Predictions due at (ISO 8601)
  --actual-due-at         Actual due at (ISO 8601)
  --domain-min            Domain minimum value
  --domain-max            Domain maximum value
  --additional-attributes Additional attributes JSON
  --reference-number      Reference number
  --parent-type           Parent type (brokers or projects)
  --parent-id             Parent ID
  --business-unit         Business unit ID
  --prediction-consensus  Prediction consensus ID

Arguments:
  <id>    The prediction subject ID (required)

Notes:
  - Updating parent requires both --parent-type and --parent-id.
  - Updating prediction consensus may require permission via membership.`,
		Example: `  # Update name and status
  xbe do prediction-subjects update 123 --name "Updated" --status active

  # Update actual and set status complete
  xbe do prediction-subjects update 123 --actual 125000 --status complete

  # Update additional attributes
  xbe do prediction-subjects update 123 --additional-attributes '{"source":"cli"}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectsUpdate,
	}
	initDoPredictionSubjectsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectsCmd.AddCommand(newDoPredictionSubjectsUpdateCmd())
}

func initDoPredictionSubjectsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Prediction subject name")
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
	cmd.Flags().String("parent-type", "", "Parent type (brokers or projects)")
	cmd.Flags().String("parent-id", "", "Parent ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("prediction-consensus", "", "Prediction consensus ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prediction subject id is required")
	}

	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
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
	if cmd.Flags().Changed("reference-number") {
		attributes["reference-number"] = opts.ReferenceNumber
	}

	relationships := map[string]any{}
	if opts.ParentType != "" || opts.ParentID != "" {
		if opts.ParentType == "" || opts.ParentID == "" {
			return fmt.Errorf("--parent-type and --parent-id are required together")
		}
		relationships["parent"] = map[string]any{
			"data": map[string]any{
				"type": opts.ParentType,
				"id":   opts.ParentID,
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

	data := map[string]any{
		"id":         id,
		"type":       "prediction-subjects",
		"attributes": attributes,
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-subjects/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prediction subject %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoPredictionSubjectsUpdateOptions(cmd *cobra.Command) (doPredictionSubjectsUpdateOptions, error) {
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
	parentType, _ := cmd.Flags().GetString("parent-type")
	parentID, _ := cmd.Flags().GetString("parent-id")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	predictionConsensus, _ := cmd.Flags().GetString("prediction-consensus")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectsUpdateOptions{
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
		ParentType:           parentType,
		ParentID:             parentID,
		BusinessUnit:         businessUnit,
		PredictionConsensus:  predictionConsensus,
	}, nil
}
