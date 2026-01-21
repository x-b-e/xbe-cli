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

type doCertificationsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	EffectiveAt string
	ExpiresAt   string
	Status      string
}

func newDoCertificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a certification",
		Long: `Update a certification.

Optional flags:
  --effective-at    Effective date (YYYY-MM-DD)
  --expires-at      Expiration date (YYYY-MM-DD)
  --status          Certification status`,
		Example: `  # Update effective date
  xbe do certifications update 123 --effective-at 2024-01-15

  # Update expiration date
  xbe do certifications update 123 --expires-at 2025-06-30

  # Update status
  xbe do certifications update 123 --status expired`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCertificationsUpdate,
	}
	initDoCertificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCertificationsCmd.AddCommand(newDoCertificationsUpdateCmd())
}

func initDoCertificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("effective-at", "", "Effective date (YYYY-MM-DD)")
	cmd.Flags().String("expires-at", "", "Expiration date (YYYY-MM-DD)")
	cmd.Flags().String("status", "", "Certification status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("effective-at") {
		attributes["effective-at"] = opts.EffectiveAt
	}
	if cmd.Flags().Changed("expires-at") {
		attributes["expires-at"] = opts.ExpiresAt
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "certifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/certifications/"+opts.ID, jsonBody)
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

	row := buildCertificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated certification %s\n", row.ID)
	return nil
}

func parseDoCertificationsUpdateOptions(cmd *cobra.Command, args []string) (doCertificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	effectiveAt, _ := cmd.Flags().GetString("effective-at")
	expiresAt, _ := cmd.Flags().GetString("expires-at")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		EffectiveAt: effectiveAt,
		ExpiresAt:   expiresAt,
		Status:      status,
	}, nil
}
