package cmd

import (
	"davidterranova/jurigen/internal/dag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	interactiveDagFile string
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive session to walk through the DAG",
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

		var d = dag.NewDAG()
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

		// Custom walk that tracks questions and answers
		type QuestionAnswer struct {
			Question string
			Answer   string
		}

		var context []QuestionAnswer
		currentNodeId := rootNode.Id

		// Interactive walk with question tracking
		for {
			// Get the current node
			currentNode, err := d.GetNode(currentNodeId)
			if err != nil {
				log.Fatalf("error getting node %s: %v", currentNodeId, err)
			}

			// If this is a leaf node (no answers), we're done
			if len(currentNode.Answers) == 0 {
				break
			}

			// Get the answer choice from CLI
			selectedAnswer, err := dag.CLIFnAnswer(currentNode)
			if err != nil {
				log.Fatalf("error getting answer for node %s: %v", currentNodeId, err)
			}

			// Add to context
			context = append(context, QuestionAnswer{
				Question: currentNode.Question,
				Answer:   selectedAnswer.Statement,
			})

			// If this answer has no next node, we've reached a leaf
			if selectedAnswer.NextNode == nil {
				break
			}

			// Move to the next node
			currentNodeId = *selectedAnswer.NextNode
		}

		// Display the final context
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("CASE CONTEXT SUMMARY")
		fmt.Println(strings.Repeat("=", 60))

		for i, qa := range context {
			fmt.Printf("%d. Q: %s\n", i+1, qa.Question)
			fmt.Printf("   A: %s\n\n", qa.Answer)
		}

		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Context built successfully with %d question-answer pairs.\n", len(context))
	},
}

func init() {
	interactiveCmd.Flags().StringVarP(&interactiveDagFile, "dag", "d", "", "Path to the DAG JSON file (required)")
	interactiveCmd.MarkFlagRequired("dag")
	rootCmd.AddCommand(interactiveCmd)
}
