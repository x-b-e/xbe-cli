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

type doMaterialSiteMergersCreateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	Orphan   string
	Survivor string
}

type materialSiteMergerRow struct {
	ID         string `json:"id"`
	OrphanID   string `json:"orphan_id,omitempty"`
	SurvivorID string `json:"survivor_id,omitempty"`
}

func newDoMaterialSiteMergersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Merge a material site into another",
		Long: `Merge an orphan material site into a surviving material site.

Required flags:
  --orphan    Orphan material site ID (required)
  --survivor  Surviving material site ID (required)`,
		Example: `  # Merge an orphan material site into a survivor
  xbe do material-site-mergers create --orphan 123 --survivor 456

  # Output as JSON
  xbe do material-site-mergers create --orphan 123 --survivor 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSiteMergersCreate,
	}
	initDoMaterialSiteMergersCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteMergersCmd.AddCommand(newDoMaterialSiteMergersCreateCmd())
}

func initDoMaterialSiteMergersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("orphan", "", "Orphan material site ID (required)")
	cmd.Flags().String("survivor", "", "Surviving material site ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteMergersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteMergersCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Orphan) == "" {
		err := fmt.Errorf("--orphan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Survivor) == "" {
		err := fmt.Errorf("--survivor is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"orphan": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.Orphan,
			},
		},
		"survivor": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.Survivor,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-site-mergers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-mergers", jsonBody)
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

	row := materialSiteMergerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site merger %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteMergersCreateOptions(cmd *cobra.Command) (doMaterialSiteMergersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	orphan, _ := cmd.Flags().GetString("orphan")
	survivor, _ := cmd.Flags().GetString("survivor")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteMergersCreateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Orphan:   orphan,
		Survivor: survivor,
	}, nil
}

func materialSiteMergerRowFromSingle(resp jsonAPISingleResponse) materialSiteMergerRow {
	row := materialSiteMergerRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["orphan"]; ok && rel.Data != nil {
		row.OrphanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["survivor"]; ok && rel.Data != nil {
		row.SurvivorID = rel.Data.ID
	}

	row.OrphanID = firstNonEmpty(
		row.OrphanID,
		stringAttr(resp.Data.Attributes, "orphan-id"),
		stringAttr(resp.Data.Attributes, "orphan_id"),
	)
	row.SurvivorID = firstNonEmpty(
		row.SurvivorID,
		stringAttr(resp.Data.Attributes, "survivor-id"),
		stringAttr(resp.Data.Attributes, "survivor_id"),
	)

	return row
}
