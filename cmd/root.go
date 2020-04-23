package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/sebradloff/rawk8stfc/pkg/hclfile"
	"github.com/spf13/cobra"
)

var (
	k8sFile       string
	outputFile    string
	contentInline bool
)

// TestCmd returns the unexported rootCmd variable
func TestCmd() *cobra.Command {
	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "rawk8stfc",
	Short: "A tool to create tf resources for all k8s objects inputed",
	RunE: func(cmd *cobra.Command, args []string) error {
		var filesToTransform []string

		// determine if k8sFile is a file or directory
		fi, err := os.Stat(k8sFile)
		if err != nil {
			return fmt.Errorf("os.Stat(%s): %v", k8sFile, err)
		}
		if fi.Mode().IsDir() {
			// get all files in directory
			files, err := ioutil.ReadDir(k8sFile)
			if err != nil {
				return fmt.Errorf("ioutil.ReadDir(%s): %v", k8sFile, err)
			}
			for _, f := range files {
				fn := filepath.Join(k8sFile, f.Name())
				filesToTransform = append(filesToTransform, fn)
			}
		} else {
			// passed in object was file
			filesToTransform = append(filesToTransform, k8sFile)
		}

		// write hcl to tf file
		hF := hclfile.NewHCLFile()

		for _, f := range filesToTransform {
			k8sObjects, err := hclfile.GetK8sObjectsFromFile(f)
			if err != nil {
				return fmt.Errorf("GetK8sObjectsFromFile(%s): %v", f, err)
			}

			// you currently can not send multiple kubernetes objects in a
			var inlineOverideNeeded bool
			if len(k8sObjects) > 1 && contentInline == false {
				log.Printf("will inline resources in file %s, since terraform-provider-k8s can not handle mutiple objects per resource block", f)
				inlineOverideNeeded = true
			}

			for _, o := range k8sObjects {
				if inlineOverideNeeded || contentInline {
					err := hF.AddK8sObjectToResourceBlockContentInline(o)
					if err != nil {
						return fmt.Errorf("AddK8sObjectToResourceBlockContentInline called with object name (%s): %v", o.GetName(), err)
					}
				} else {
					err := hF.AddK8sObjectToResourceBlockContentFile(o, f)
					if err != nil {
						return fmt.Errorf("AddK8sObjectToResourceBlockContentFile called with object name (%s) and k8s file to reference (%s): %v", o.GetName(), f, err)
					}
				}
			}
		}

		err = hF.WriteToFile(outputFile)
		if err != nil {
			return fmt.Errorf("error writing hcl to file %s: %v", outputFile, err)
		}
		return nil
	},
}

// Execute root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&k8sFile, "k8sFile", "f", "", "k8s file or directory to read for tf resources")
	rootCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "", "output file where generated tf will be written")
	rootCmd.MarkFlagRequired("k8sFile")
	rootCmd.MarkFlagRequired("outputFile")
	rootCmd.Flags().BoolVarP(&contentInline, "contentInline", "i", true, "the content attribute in the resource block will have k8s yaml as a heredoc by default. If false, it will refernce the k8s file.")
}
