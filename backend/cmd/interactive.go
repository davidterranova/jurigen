package cmd

import (
	"davidterranova/jurigen/backend/internal/dag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	interactiveDagFile string
	collectContext     bool
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive session to walk through the DAG with optional context collection",
	Run: func(cmd *cobra.Command, args []string) {
		// Validate that the file path is provided
		if interactiveDagFile == "" {
			log.Fatalf("DAG file path is required. Use -d or --dag flag to specify the file path")
		}

		// Check if file exists
		if _, err := os.Stat(interactiveDagFile); os.IsNotExist(err) {
			log.Fatalf("DAG file '%s' does not exist", interactiveDagFile)
		}

		// Load DAG from file
		data, err := os.ReadFile(interactiveDagFile)
		if err != nil {
			log.Fatalf("error reading file '%s': %v", interactiveDagFile, err)
		}

		var d = dag.NewDAG("Interactive DAG")
		err = d.UnmarshalJSON(data)
		if err != nil {
			log.Fatalf("error unmarshalling file '%s': %v", interactiveDagFile, err)
		}

		// Find the root node
		rootNode, err := d.GetRootNode()
		if err != nil {
			log.Fatalf("error finding root node: %v", err)
		}

		fmt.Println("=== Interactive Legal Case Context Builder ===")
		fmt.Println("Answer the following questions to build your case context.")
		fmt.Println("Enter the number corresponding to your choice.")
		fmt.Println()

		// Choose the appropriate answer provider based on context flag
		var answerProvider func(dag.Node) (dag.Answer, error)
		if collectContext {
			answerProvider = dag.CLIFnAnswerWithContext
			fmt.Println("📝 Context collection enabled - you'll be prompted for additional details.")
			fmt.Println()
		} else {
			answerProvider = dag.CLIFnAnswer
		}

		// Use the DAG's Walk function with the selected answer provider
		path, err := d.Walk(rootNode.Id, answerProvider)
		if err != nil {
			log.Fatalf("error walking through DAG: %v", err)
		}

		// Display the final context
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("CASE CONTEXT SUMMARY")
		fmt.Println(strings.Repeat("=", 60))

		for i, answer := range path {
			fmt.Printf("%d. Q: %s\n", i+1, answer.ParentNode.Question)
			fmt.Printf("   A: %s\n", answer.Statement)

			// Display additional context if available
			if answer.UserContext != "" {
				fmt.Printf("   📝 Notes: %s\n", answer.UserContext)
			}

			if len(answer.Metadata) > 0 {
				if conf, ok := answer.Metadata["confidence"].(float64); ok {
					fmt.Printf("   📊 Confidence: %.1f/1.0\n", conf)
				}
				if tagsRaw, ok := answer.Metadata["tags"]; ok {
					var tagStrs []string
					switch tags := tagsRaw.(type) {
					case []string:
						tagStrs = tags
					case []interface{}:
						tagStrs = make([]string, len(tags))
						for i, tag := range tags {
							if tagStr, ok := tag.(string); ok {
								tagStrs[i] = tagStr
							}
						}
					}
					if len(tagStrs) > 0 {
						fmt.Printf("   🏷️  Tags: %s\n", strings.Join(tagStrs, ", "))
					}
				}
			}

			fmt.Println()
		}

		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Context built successfully with %d question-answer pairs.\n", len(path))
	},
}

func init() {
	interactiveCmd.Flags().StringVarP(&interactiveDagFile, "dag", "d", "", "Path to the DAG JSON file (required)")
	interactiveCmd.Flags().BoolVarP(&collectContext, "context", "c", false, "Collect additional context and metadata for each answer")
	err := interactiveCmd.MarkFlagRequired("dag")
	if err != nil {
		log.Fatalf("error marking flag as required: %v", err)
	}

	rootCmd.AddCommand(interactiveCmd)
}
