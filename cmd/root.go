package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	k8sFile    string
	outputFile string
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rawk8stfc",
		Short: "A tool to create tf resources for all k8s objects inputed",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmd.PersistentFlags().StringVarP(&k8sFile, "k8sFile", "f", "", "k8s file or directory to read for tf resources")
	cmd.PersistentFlags().StringVarP(&outputFile, "outputFile", "o", "", "output file where generated tf will be written")
	err := cmd.MarkPersistentFlagRequired("k8sFile")
	if err != nil {
		panic(fmt.Errorf("error setting persistent flag k8sFile: %v", err))
	}
	err = cmd.MarkPersistentFlagRequired("outputFile")
	if err != nil {
		panic(fmt.Errorf("error setting persistent flag outputFile: %v", err))
	}

	cmd.AddCommand(inlineCmd)
	cmd.AddCommand(fileReferenceCmd)

	return cmd
}

// Execute root command
func Execute() error {
	return NewRootCmd().Execute()
}
