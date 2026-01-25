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

type doTenderRejectionsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Tender     string
	TenderType string
	Comment    string
}

func newDoTenderRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reject a tender",
		Long: `Reject a tender.

Tenders must be in offered status before rejection.

Required flags:
  --tender       Tender ID (or Type|ID)
  --tender-type  Tender type (e.g., broker-tenders, customer-tenders) when --tender is an ID

Optional flags:
  --comment      Rejection comment

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Reject a customer tender with a comment
  xbe do tender-rejections create \
    --tender-type customer-tenders \
    --tender 123 \
    --comment "No capacity available"

  # Reject a broker tender using Type|ID format
  xbe do tender-rejections create --tender broker-tenders|456`,
		Args: cobra.NoArgs,
		RunE: runDoTenderRejectionsCreate,
	}
	initDoTenderRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderRejectionsCmd.AddCommand(newDoTenderRejectionsCreateCmd())
}

func initDoTenderRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender", "", "Tender ID or Type|ID (required)")
	cmd.Flags().String("tender-type", "", "Tender type (e.g., broker-tenders, customer-tenders)")
	cmd.Flags().String("comment", "", "Rejection comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderRejectionsCreateOptions(cmd)
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

	tenderType, tenderID, err := parseTenderReference(opts.Tender, opts.TenderType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": tenderType,
				"id":   tenderID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-rejections",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tender-rejections", jsonBody)
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

	row := buildTenderRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender rejection %s\n", row.ID)
	return nil
}

func parseDoTenderRejectionsCreateOptions(cmd *cobra.Command) (doTenderRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tender, _ := cmd.Flags().GetString("tender")
	tenderType, _ := cmd.Flags().GetString("tender-type")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderRejectionsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Tender:     tender,
		TenderType: tenderType,
		Comment:    comment,
	}, nil
}

func parseTenderReference(tenderValue, tenderType string) (string, string, error) {
	tenderValue = strings.TrimSpace(tenderValue)
	tenderType = strings.TrimSpace(tenderType)

	if tenderValue == "" {
		return "", "", fmt.Errorf("--tender is required")
	}

	if strings.Contains(tenderValue, "|") {
		if tenderType != "" {
			return "", "", fmt.Errorf("use --tender Type|ID or --tender with --tender-type, not both")
		}
		parts := strings.SplitN(tenderValue, "|", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("--tender must be in Type|ID format (e.g. broker-tenders|123)")
		}
		typePart := strings.TrimSpace(parts[0])
		idPart := strings.TrimSpace(parts[1])
		if typePart == "" || idPart == "" {
			return "", "", fmt.Errorf("--tender must be in Type|ID format (e.g. broker-tenders|123)")
		}
		return typePart, idPart, nil
	}

	if tenderType == "" {
		return "", "", fmt.Errorf("--tender-type is required when --tender is an ID (e.g. broker-tenders)")
	}

	return tenderType, tenderValue, nil
}
