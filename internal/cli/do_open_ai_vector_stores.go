package cli

import "github.com/spf13/cobra"

var doOpenAiVectorStoresCmd = &cobra.Command{
	Use:   "open-ai-vector-stores",
	Short: "Manage OpenAI vector stores",
	Long: `Manage OpenAI vector stores used for embeddings and retrieval.

Commands:
  create    Create a vector store
  update    Update a vector store
  delete    Delete a vector store`,
	Example: `  # Create a vector store
  xbe do open-ai-vector-stores create --purpose user_post_feed --scope-type UserPostFeed --scope-id 123

  # Update a vector store
  xbe do open-ai-vector-stores update 123 --purpose platform_content

  # Delete a vector store (requires --confirm)
  xbe do open-ai-vector-stores delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doOpenAiVectorStoresCmd)
}
