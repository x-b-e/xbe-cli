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

type doCrewRequirementCredentialClassificationsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	CrewRequirementType          string
	CrewRequirementID            string
	CredentialClassificationType string
	CredentialClassificationID   string
}

func newDoCrewRequirementCredentialClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a crew requirement credential classification",
		Long: `Create a crew requirement credential classification.

Required flags:
  --crew-requirement-type           Crew requirement type (e.g., labor-requirements) (required)
  --crew-requirement                Crew requirement ID (required)
  --credential-classification-type  Credential classification type (e.g., user-credential-classifications) (required)
  --credential-classification       Credential classification ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a crew requirement credential classification
  xbe do crew-requirement-credential-classifications create \
    --crew-requirement-type labor-requirements \
    --crew-requirement 123 \
    --credential-classification-type user-credential-classifications \
    --credential-classification 456`,
		Args: cobra.NoArgs,
		RunE: runDoCrewRequirementCredentialClassificationsCreate,
	}
	initDoCrewRequirementCredentialClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doCrewRequirementCredentialClassificationsCmd.AddCommand(newDoCrewRequirementCredentialClassificationsCreateCmd())
}

func initDoCrewRequirementCredentialClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("crew-requirement-type", "", "Crew requirement type (e.g., labor-requirements) (required)")
	cmd.Flags().String("crew-requirement", "", "Crew requirement ID (required)")
	cmd.Flags().String("credential-classification-type", "", "Credential classification type (e.g., user-credential-classifications) (required)")
	cmd.Flags().String("credential-classification", "", "Credential classification ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCrewRequirementCredentialClassificationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCrewRequirementCredentialClassificationsCreateOptions(cmd)
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

	if opts.CrewRequirementType == "" {
		err := fmt.Errorf("--crew-requirement-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CrewRequirementID == "" {
		err := fmt.Errorf("--crew-requirement is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CredentialClassificationType == "" {
		err := fmt.Errorf("--credential-classification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.CredentialClassificationID == "" {
		err := fmt.Errorf("--credential-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"crew-requirement": map[string]any{
			"data": map[string]any{
				"type": opts.CrewRequirementType,
				"id":   opts.CrewRequirementID,
			},
		},
		"credential-classification": map[string]any{
			"data": map[string]any{
				"type": opts.CredentialClassificationType,
				"id":   opts.CredentialClassificationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "crew-requirement-credential-classifications",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/crew-requirement-credential-classifications", jsonBody)
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

	row := buildCrewRequirementCredentialClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created crew requirement credential classification %s\n", row.ID)
	return nil
}

func parseDoCrewRequirementCredentialClassificationsCreateOptions(cmd *cobra.Command) (doCrewRequirementCredentialClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	crewRequirementType, _ := cmd.Flags().GetString("crew-requirement-type")
	crewRequirementID, _ := cmd.Flags().GetString("crew-requirement")
	credentialClassificationType, _ := cmd.Flags().GetString("credential-classification-type")
	credentialClassificationID, _ := cmd.Flags().GetString("credential-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCrewRequirementCredentialClassificationsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		CrewRequirementType:          crewRequirementType,
		CrewRequirementID:            crewRequirementID,
		CredentialClassificationType: credentialClassificationType,
		CredentialClassificationID:   credentialClassificationID,
	}, nil
}

func buildCrewRequirementCredentialClassificationRowFromSingle(resp jsonAPISingleResponse) crewRequirementCredentialClassificationRow {
	row := crewRequirementCredentialClassificationRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["crew-requirement"]; ok && rel.Data != nil {
		row.CrewRequirementType = rel.Data.Type
		row.CrewRequirementID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["credential-classification"]; ok && rel.Data != nil {
		row.CredentialClassificationType = rel.Data.Type
		row.CredentialClassificationID = rel.Data.ID
	}

	return row
}
