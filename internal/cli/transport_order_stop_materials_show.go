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

type transportOrderStopMaterialsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newTransportOrderStopMaterialsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show transport order stop material details",
		Long: `Show the full details of a transport order stop material.

Output Fields:
  ID
  Quantity Explicit
  Transport Order Material ID
  Transport Order Stop ID

Arguments:
  <id>  Transport order stop material ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a transport order stop material
  xbe view transport-order-stop-materials show 123

  # Output as JSON
  xbe view transport-order-stop-materials show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTransportOrderStopMaterialsShow,
	}
	initTransportOrderStopMaterialsShowFlags(cmd)
	return cmd
}

func init() {
	transportOrderStopMaterialsCmd.AddCommand(newTransportOrderStopMaterialsShowCmd())
}

func initTransportOrderStopMaterialsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTransportOrderStopMaterialsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTransportOrderStopMaterialsShowOptions(cmd)
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
		return fmt.Errorf("transport order stop material id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[transport-order-stop-materials]", "quantity-explicit,transport-order-material,transport-order-stop")

	body, _, err := client.Get(cmd.Context(), "/v1/transport-order-stop-materials/"+id, query)
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

	details := buildTransportOrderStopMaterialRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTransportOrderStopMaterialDetails(cmd, details)
}

func parseTransportOrderStopMaterialsShowOptions(cmd *cobra.Command) (transportOrderStopMaterialsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return transportOrderStopMaterialsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderTransportOrderStopMaterialDetails(cmd *cobra.Command, details transportOrderStopMaterialRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.QuantityExplicit != "" {
		fmt.Fprintf(out, "Quantity Explicit: %s\n", details.QuantityExplicit)
	}
	if details.TransportOrderMaterial != "" {
		fmt.Fprintf(out, "Transport Order Material ID: %s\n", details.TransportOrderMaterial)
	}
	if details.TransportOrderStop != "" {
		fmt.Fprintf(out, "Transport Order Stop ID: %s\n", details.TransportOrderStop)
	}

	return nil
}
