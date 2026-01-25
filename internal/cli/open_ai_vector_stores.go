package cli

import "github.com/spf13/cobra"

var openAiVectorStoresCmd = &cobra.Command{
	Use:   "open-ai-vector-stores",
	Short: "Browse OpenAI vector stores",
	Long: `Browse OpenAI vector stores used for retrieval and embeddings.

Vector stores are tied to scopes such as user post feeds and brokers and are
used for features like content search and recap generation.

Commands:
  list    List vector stores with filtering
  show    Show vector store details`,
	Example: `  # List vector stores
  xbe view open-ai-vector-stores list

  # Filter by purpose
  xbe view open-ai-vector-stores list --purpose user_post_feed

  # Show details
  xbe view open-ai-vector-stores show 123`,
}

func init() {
	viewCmd.AddCommand(openAiVectorStoresCmd)
}
