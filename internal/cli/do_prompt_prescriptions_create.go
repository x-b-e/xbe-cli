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

type doPromptPrescriptionsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	EmailAddress     string
	Name             string
	OrganizationName string
	LocationName     string
	Role             string
	Symptoms         string
}

func newDoPromptPrescriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prompt prescription",
		Long: `Create a prompt prescription request and generate tailored prompts.

Required flags:
  --email-address      Contact email address (required)
  --name               Contact name (required)
  --organization-name  Organization name (required)
  --location-name      Location name (required)
  --role               Role or job title (required)

Optional flags:
  --symptoms            Challenges to incorporate into the prompt suggestions

Note: Prompt generation runs asynchronously. Use the show command to retrieve
the generated prompts once they're available.`,
		Example: `  # Create a prompt prescription
  xbe do prompt-prescriptions create \\
    --email-address "name@example.com" \\
    --name "Alex Builder" \\
    --organization-name "Concrete Co" \\
    --location-name "Austin, TX" \\
    --role "Operations Manager" \\
    --symptoms "Rising costs, scheduling delays"

  # Output as JSON
  xbe do prompt-prescriptions create \\
    --email-address "name@example.com" \\
    --name "Alex Builder" \\
    --organization-name "Concrete Co" \\
    --location-name "Austin, TX" \\
    --role "Operations Manager" \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoPromptPrescriptionsCreate,
	}
	initDoPromptPrescriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doPromptPrescriptionsCmd.AddCommand(newDoPromptPrescriptionsCreateCmd())
}

func initDoPromptPrescriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("email-address", "", "Contact email address (required)")
	cmd.Flags().String("name", "", "Contact name (required)")
	cmd.Flags().String("organization-name", "", "Organization name (required)")
	cmd.Flags().String("location-name", "", "Location name (required)")
	cmd.Flags().String("role", "", "Role or job title (required)")
	cmd.Flags().String("symptoms", "", "Challenges to incorporate into prompts")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("email-address")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("organization-name")
	cmd.MarkFlagRequired("location-name")
	cmd.MarkFlagRequired("role")
}

func runDoPromptPrescriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPromptPrescriptionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	opts.EmailAddress = strings.TrimSpace(opts.EmailAddress)
	opts.Name = strings.TrimSpace(opts.Name)
	opts.OrganizationName = strings.TrimSpace(opts.OrganizationName)
	opts.LocationName = strings.TrimSpace(opts.LocationName)
	opts.Role = strings.TrimSpace(opts.Role)
	opts.Symptoms = strings.TrimSpace(opts.Symptoms)

	if opts.EmailAddress == "" || opts.Name == "" || opts.OrganizationName == "" || opts.LocationName == "" || opts.Role == "" {
		err := fmt.Errorf("--email-address, --name, --organization-name, --location-name, and --role are required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"email-address":     opts.EmailAddress,
		"name":              opts.Name,
		"organization-name": opts.OrganizationName,
		"location-name":     opts.LocationName,
		"role":              opts.Role,
	}
	if opts.Symptoms != "" {
		attributes["symptoms"] = opts.Symptoms
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prompt-prescriptions",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/prompt-prescriptions", jsonBody)
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

	row := promptPrescriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created prompt prescription %s\n", row.ID)
	return nil
}

func parseDoPromptPrescriptionsCreateOptions(cmd *cobra.Command) (doPromptPrescriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	name, _ := cmd.Flags().GetString("name")
	organizationName, _ := cmd.Flags().GetString("organization-name")
	locationName, _ := cmd.Flags().GetString("location-name")
	role, _ := cmd.Flags().GetString("role")
	symptoms, _ := cmd.Flags().GetString("symptoms")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPromptPrescriptionsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		EmailAddress:     emailAddress,
		Name:             name,
		OrganizationName: organizationName,
		LocationName:     locationName,
		Role:             role,
		Symptoms:         symptoms,
	}, nil
}
