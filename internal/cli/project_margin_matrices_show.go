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

type projectMarginMatricesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectMarginMatrixDetails struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
	Scenarios any    `json:"scenarios,omitempty"`
}

func newProjectMarginMatricesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project margin matrix details",
		Long: `Show the full details of a project margin matrix.

Output Fields:
  ID          Matrix identifier
  Project ID  Project identifier
  Scenarios   Scenario array for margin analysis

Arguments:
  <id>    The project margin matrix ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project margin matrix
  xbe view project-margin-matrices show 123

  # Get JSON output
  xbe view project-margin-matrices show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectMarginMatricesShow,
	}
	initProjectMarginMatricesShowFlags(cmd)
	return cmd
}

func init() {
	projectMarginMatricesCmd.AddCommand(newProjectMarginMatricesShowCmd())
}

func initProjectMarginMatricesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectMarginMatricesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectMarginMatricesShowOptions(cmd)
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
		return fmt.Errorf("project margin matrix id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-margin-matrices/"+id, nil)
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

	details := buildProjectMarginMatrixDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectMarginMatrixDetails(cmd, details)
}

func parseProjectMarginMatricesShowOptions(cmd *cobra.Command) (projectMarginMatricesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectMarginMatricesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectMarginMatrixDetails(resp jsonAPISingleResponse) projectMarginMatrixDetails {
	resource := resp.Data
	details := projectMarginMatrixDetails{
		ID:        resource.ID,
		Scenarios: resource.Attributes["scenarios"],
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}

	return details
}

func renderProjectMarginMatrixDetails(cmd *cobra.Command, details projectMarginMatrixDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.Scenarios != nil {
		fmt.Fprintf(out, "Scenario Count: %d\n", projectMarginMatrixScenarioCount(details.Scenarios))
		fmt.Fprintln(out, "\nScenarios:")
		fmt.Fprintln(out, formatProjectMarginMatrixJSONBlock(details.Scenarios, "  "))
	}

	return nil
}
