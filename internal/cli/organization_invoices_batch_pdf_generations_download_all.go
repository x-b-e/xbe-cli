package cli

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
	"github.com/xbe-inc/xbe-cli/internal/version"
)

type organizationInvoicesBatchPdfGenerationsDownloadAllOptions struct {
	BaseURL   string
	Token     string
	NoAuth    bool
	Output    string
	Overwrite bool
}

func newOrganizationInvoicesBatchPdfGenerationsDownloadAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download-all <id>",
		Short: "Download all completed PDFs for a batch PDF generation",
		Long: `Download a ZIP archive containing all completed PDFs for a generation.

If no PDFs are completed, the API returns a 404 error.

Arguments:
  <id>    The PDF generation ID (required)

Global flags (see xbe --help): --base-url, --token, --no-auth`,
		Example: `  # Download all PDFs for a generation
  xbe view organization-invoices-batch-pdf-generations download-all 123 --output ./batch_pdfs.zip

  # Use server filename in the current directory
  xbe view organization-invoices-batch-pdf-generations download-all 123`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchPdfGenerationsDownloadAll,
	}
	initOrganizationInvoicesBatchPdfGenerationsDownloadAllFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfGenerationsCmd.AddCommand(newOrganizationInvoicesBatchPdfGenerationsDownloadAllCmd())
}

func initOrganizationInvoicesBatchPdfGenerationsDownloadAllFlags(cmd *cobra.Command) {
	cmd.Flags().String("output", "", "Output file path (optional)")
	cmd.Flags().Bool("overwrite", false, "Overwrite output file if it exists")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfGenerationsDownloadAll(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationInvoicesBatchPdfGenerationsDownloadAllOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("organization invoices batch PDF generation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	base, err := url.Parse(client.BaseURL)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	path := "/v1/organization-invoices-batch-pdf-generations/" + id + "/download-all"
	base.Path = strings.TrimRight(base.Path, "/") + path

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, base.String(), nil)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if client.Token != "" {
		req.Header.Set("Authorization", "Bearer "+client.Token)
	}
	req.Header.Set("Accept", "application/zip")
	req.Header.Set("User-Agent", "xbe-cli/"+version.String())

	client.HTTPClient.Timeout = 60 * time.Second
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	filename := filenameFromDisposition(resp.Header.Get("Content-Disposition"))
	if filename == "" {
		filename = fmt.Sprintf("organization-invoices-batch-pdf-generation-%s.zip", id)
	}

	outputPath := strings.TrimSpace(opts.Output)
	if outputPath == "" {
		outputPath = filename
	} else if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, filename)
	}

	if !opts.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("output file already exists: %s", outputPath)
		}
	}

	if err := os.WriteFile(outputPath, body, 0600); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Saved PDF archive to %s\n", outputPath)
	return nil
}

func parseOrganizationInvoicesBatchPdfGenerationsDownloadAllOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfGenerationsDownloadAllOptions, error) {
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{}, err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{}, err
	}
	overwrite, err := cmd.Flags().GetBool("overwrite")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{}, err
	}

	return organizationInvoicesBatchPdfGenerationsDownloadAllOptions{
		BaseURL:   baseURL,
		Token:     token,
		NoAuth:    noAuth,
		Output:    output,
		Overwrite: overwrite,
	}, nil
}

func filenameFromDisposition(disposition string) string {
	if strings.TrimSpace(disposition) == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(params["filename"])
}
