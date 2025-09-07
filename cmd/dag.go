package cmd

import (
	"davidterranova/jurigen/internal/dag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	dagFile string
)

var dagCmd = &cobra.Command{
	Use:   "dag",
	Short: "Manage the DAG",
	Run: func(cmd *cobra.Command, args []string) {
		// Validate that the file path is provided
		if dagFile == "" {
			log.Fatalf("DAG file path is required. Use -d or --dag flag to specify the file path")
		}

		// Check if file exists
		if _, err := os.Stat(dagFile); os.IsNotExist(err) {
			log.Fatalf("DAG file '%s' does not exist", dagFile)
		}

		data, err := os.ReadFile(dagFile)
		if err != nil {
			log.Fatalf("error reading file '%s': %v", dagFile, err)
		}

		var dag = dag.NewDAG("Sample DAG")
		err = dag.UnmarshalJSON(data)
		if err != nil {
			log.Fatalf("error unmarshalling file '%s': %v", dagFile, err)
		}

		fmt.Println(dag)
	},
}

func init() {
	dagCmd.Flags().StringVarP(&dagFile, "dag", "d", "", "Path to the DAG JSON file (required)")
	err := dagCmd.MarkFlagRequired("dag")
	if err != nil {
		log.Fatalf("error marking flag as required: %v", err)
	}

	rootCmd.AddCommand(dagCmd)
}
