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

type crewRequirementCredentialClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type crewRequirementCredentialClassificationDetails struct {
	ID                           string `json:"id"`
	CrewRequirementType          string `json:"crew_requirement_type,omitempty"`
	CrewRequirementID            string `json:"crew_requirement_id,omitempty"`
	CredentialClassificationType string `json:"credential_classification_type,omitempty"`
	CredentialClassificationID   string `json:"credential_classification_id,omitempty"`
}

func newCrewRequirementCredentialClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show crew requirement credential classification details",
		Long: `Show the full details of a crew requirement credential classification.

Output Fields:
  ID
  Crew Requirement Type
  Crew Requirement ID
  Credential Classification Type
  Credential Classification ID

Arguments:
  <id>    The link ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a link
  xbe view crew-requirement-credential-classifications show 123

  # Get JSON output
  xbe view crew-requirement-credential-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCrewRequirementCredentialClassificationsShow,
	}
	initCrewRequirementCredentialClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	crewRequirementCredentialClassificationsCmd.AddCommand(newCrewRequirementCredentialClassificationsShowCmd())
}

func initCrewRequirementCredentialClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCrewRequirementCredentialClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCrewRequirementCredentialClassificationsShowOptions(cmd)
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
		return fmt.Errorf("crew requirement credential classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[crew-requirement-credential-classifications]", "crew-requirement,credential-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/crew-requirement-credential-classifications/"+id, query)
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

	details := buildCrewRequirementCredentialClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCrewRequirementCredentialClassificationDetails(cmd, details)
}

func parseCrewRequirementCredentialClassificationsShowOptions(cmd *cobra.Command) (crewRequirementCredentialClassificationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return crewRequirementCredentialClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCrewRequirementCredentialClassificationDetails(resp jsonAPISingleResponse) crewRequirementCredentialClassificationDetails {
	details := crewRequirementCredentialClassificationDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["crew-requirement"]; ok && rel.Data != nil {
		details.CrewRequirementType = rel.Data.Type
		details.CrewRequirementID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["credential-classification"]; ok && rel.Data != nil {
		details.CredentialClassificationType = rel.Data.Type
		details.CredentialClassificationID = rel.Data.ID
	}

	return details
}

func renderCrewRequirementCredentialClassificationDetails(cmd *cobra.Command, details crewRequirementCredentialClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CrewRequirementType != "" {
		fmt.Fprintf(out, "Crew Requirement Type: %s\n", details.CrewRequirementType)
	}
	if details.CrewRequirementID != "" {
		fmt.Fprintf(out, "Crew Requirement ID: %s\n", details.CrewRequirementID)
	}
	if details.CredentialClassificationType != "" {
		fmt.Fprintf(out, "Credential Classification Type: %s\n", details.CredentialClassificationType)
	}
	if details.CredentialClassificationID != "" {
		fmt.Fprintf(out, "Credential Classification ID: %s\n", details.CredentialClassificationID)
	}

	return nil
}
