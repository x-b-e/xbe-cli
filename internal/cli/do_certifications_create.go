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

type doCertificationsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	CertificationTypeID string
	CertifiesType       string
	CertifiesID         string
	EffectiveAt         string
	ExpiresAt           string
	Status              string
}

func newDoCertificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new certification",
		Long: `Create a new certification.

Required flags:
  --certification-type    Certification type ID (required)
  --certifies-type        Type of entity to certify (e.g., truckers) (required)
  --certifies-id          ID of entity to certify (required)

Optional flags:
  --effective-at    Effective date (YYYY-MM-DD)
  --expires-at      Expiration date (YYYY-MM-DD)
  --status          Certification status`,
		Example: `  # Create a certification for a trucker
  xbe do certifications create --certification-type 123 --certifies-type truckers --certifies-id 456

  # Create with dates
  xbe do certifications create --certification-type 123 --certifies-type truckers --certifies-id 456 --effective-at 2024-01-01 --expires-at 2025-01-01`,
		Args: cobra.NoArgs,
		RunE: runDoCertificationsCreate,
	}
	initDoCertificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doCertificationsCmd.AddCommand(newDoCertificationsCreateCmd())
}

func initDoCertificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("certification-type", "", "Certification type ID (required)")
	cmd.Flags().String("certifies-type", "", "Type of entity to certify (e.g., truckers) (required)")
	cmd.Flags().String("certifies-id", "", "ID of entity to certify (required)")
	cmd.Flags().String("effective-at", "", "Effective date (YYYY-MM-DD)")
	cmd.Flags().String("expires-at", "", "Expiration date (YYYY-MM-DD)")
	cmd.Flags().String("status", "", "Certification status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationsCreateOptions(cmd)
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

	if opts.CertificationTypeID == "" {
		err := fmt.Errorf("--certification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.CertifiesType == "" {
		err := fmt.Errorf("--certifies-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.CertifiesID == "" {
		err := fmt.Errorf("--certifies-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.EffectiveAt != "" {
		attributes["effective-at"] = opts.EffectiveAt
	}
	if opts.ExpiresAt != "" {
		attributes["expires-at"] = opts.ExpiresAt
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}

	relationships := map[string]any{
		"certification-type": map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		},
		"certifies": map[string]any{
			"data": map[string]any{
				"type": opts.CertifiesType,
				"id":   opts.CertifiesID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "certifications",
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

	body, _, err := client.Post(cmd.Context(), "/v1/certifications", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created certification %s\n", row.ID)
	return nil
}

func parseDoCertificationsCreateOptions(cmd *cobra.Command) (doCertificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	certifiesType, _ := cmd.Flags().GetString("certifies-type")
	certifiesID, _ := cmd.Flags().GetString("certifies-id")
	effectiveAt, _ := cmd.Flags().GetString("effective-at")
	expiresAt, _ := cmd.Flags().GetString("expires-at")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		CertificationTypeID: certificationTypeID,
		CertifiesType:       certifiesType,
		CertifiesID:         certifiesID,
		EffectiveAt:         effectiveAt,
		ExpiresAt:           expiresAt,
		Status:              status,
	}, nil
}

func buildCertificationRowFromSingle(resp jsonAPISingleResponse) certificationRow {
	attrs := resp.Data.Attributes

	row := certificationRow{
		ID:          resp.Data.ID,
		Status:      stringAttr(attrs, "status"),
		EffectiveAt: stringAttr(attrs, "effective-at"),
		ExpiresAt:   stringAttr(attrs, "expires-at"),
	}

	if rel, ok := resp.Data.Relationships["certifies"]; ok && rel.Data != nil {
		row.CertifiesType = rel.Data.Type
		row.CertifiesID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["certification-type"]; ok && rel.Data != nil {
		row.CertificationTypeID = rel.Data.ID
	}

	return row
}
