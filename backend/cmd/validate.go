package cmd

import (
	"davidterranova/jurigen/backend/internal/dag"
	"davidterranova/jurigen/backend/internal/usecase"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate DAG structure and integrity",
	Long: `Validate a DAG file to ensure it has:
- A single root node
- No cycles (acyclic structure)
- Valid node and answer relationships
- Proper UUID formats

This command will check the DAG structure and provide detailed validation results.`,
}

var validateFileCmd = &cobra.Command{
	Use:   "file [path]",
	Short: "Validate a DAG file",
	Long: `Validate a DAG file to ensure it meets all structural requirements.
	
Examples:
  jurigen validate file data/my-dag.json
  jurigen validate file data/my-dag.json --detailed
  jurigen validate file data/my-dag.json --stats-only`,
	Args: cobra.ExactArgs(1),
	RunE: validateDAGFile,
}

var (
	detailedOutput bool
	statsOnly      bool
	outputFormat   string
)

func init() {
	validateFileCmd.Flags().BoolVar(&detailedOutput, "detailed", false, "Show detailed validation errors and warnings")
	validateFileCmd.Flags().BoolVar(&statsOnly, "stats-only", false, "Show only DAG statistics")
	validateFileCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json")

	validateCmd.AddCommand(validateFileCmd)
	rootCmd.AddCommand(validateCmd)
}

func validateDAGFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read the DAG file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse JSON
	var dagData dag.DAG
	if err := json.Unmarshal(data, &dagData); err != nil {
		return fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
	}

	// Validate DAG
	validator := usecase.NewDAGValidator()
	result := validator.ValidateDAG(&dagData)

	// Output results based on format and options
	switch outputFormat {
	case "json":
		return outputJSONResults(result)
	default:
		return outputTextResults(filePath, result)
	}
}

func outputJSONResults(result usecase.ValidationResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal validation results: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func outputTextResults(filePath string, result usecase.ValidationResult) error {
	fmt.Printf("🔍 DAG Validation Results for: %s\n", filePath)
	fmt.Println(strings.Repeat("=", 50))

	// Overall status
	if result.IsValid {
		fmt.Println("✅ DAG is VALID")
	} else {
		fmt.Println("❌ DAG is INVALID")
	}

	fmt.Println()

	// Statistics (always show unless it's a fatal error)
	if result.Statistics.TotalNodes > 0 || len(result.Errors) == 0 {
		fmt.Println("📊 DAG Statistics:")
		fmt.Printf("   Total Nodes: %d\n", result.Statistics.TotalNodes)
		fmt.Printf("   Root Nodes: %d\n", result.Statistics.RootNodes)
		fmt.Printf("   Leaf Nodes: %d\n", result.Statistics.LeafNodes)
		fmt.Printf("   Total Answers: %d\n", result.Statistics.TotalAnswers)
		fmt.Printf("   Max Depth: %d\n", result.Statistics.MaxDepth)
		fmt.Printf("   Has Cycles: %v\n", result.Statistics.HasCycles)

		if len(result.Statistics.RootNodeIDs) > 0 {
			fmt.Printf("   Root Node IDs: %v\n", result.Statistics.RootNodeIDs)
		}

		if len(result.Statistics.CyclePaths) > 0 {
			fmt.Printf("   Cycle Paths: %v\n", result.Statistics.CyclePaths)
		}
		fmt.Println()
	}

	if statsOnly {
		return nil
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Printf("❌ Validation Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("   %d. [%s] %s", i+1, err.Code, err.Message)
			if err.NodeID != "" {
				fmt.Printf(" (Node: %s)", err.NodeID)
			}
			if err.AnswerID != "" {
				fmt.Printf(" (Answer: %s)", err.AnswerID)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  Validation Warnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Printf("   %d. [%s] %s", i+1, warning.Code, warning.Message)
			if warning.NodeID != "" {
				fmt.Printf(" (Node: %s)", warning.NodeID)
			}
			if warning.AnswerID != "" {
				fmt.Printf(" (Answer: %s)", warning.AnswerID)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Detailed output
	if detailedOutput && result.IsValid {
		fmt.Println("🔍 Detailed Analysis:")
		fmt.Printf("   ✓ Single root node validation passed\n")
		fmt.Printf("   ✓ Acyclic structure validation passed\n")
		fmt.Printf("   ✓ All node references are valid\n")
		fmt.Printf("   ✓ All UUIDs are properly formatted\n")
		fmt.Println()
	}

	// Exit code based on validation result
	if !result.IsValid {
		os.Exit(1)
	}

	return nil
}
