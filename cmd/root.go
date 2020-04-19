package cmd

import (
	"github.com/spf13/cobra"
)

var (
	k8sFile   string
	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "rawk8stfc",
	Short: "A tool to create tf resources for all k8s objects inputed",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&k8sFile, "k8sFile", "f", "", "k8s file or directory to read for tf resources")
	rootCmd.Flags().StringVarP(&outputDir, "outputDir", "o", "", "output directory where generated tf will be written")
	rootCmd.MarkFlagRequired("k8sFile")
	rootCmd.MarkFlagRequired("outputDir")
}
