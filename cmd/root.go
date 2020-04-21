package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sebradloff/rawk8stfc/pkg/hcl"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	k8sFile    string
	outputFile string
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
			return fmt.Errorf("could not get os.Stat(%s); err = %v", k8sFile, err)
		}
		if fi.Mode().IsDir() {
			// get all files in directory
			files, err := ioutil.ReadDir(k8sFile)
			if err != nil {
				return fmt.Errorf("could not ioutil.ReadDir(%s); err = %v", k8sFile, err)
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
		hF := hcl.NewHCLFile()

		for _, f := range filesToTransform {

			data, err := ioutil.ReadFile(f)
			if err != nil {
				return fmt.Errorf("could not ioutil.ReadFile(%s); err = %v", f, err)
			}

			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)

			// allows us to capture yaml streams
			for {
				var o *unstructured.Unstructured
				// decode one yaml strem into a k8s object
				err = decoder.Decode(&o)
				if err != nil && err != io.EOF {
					return fmt.Errorf("Failed to unmarshal manifest: %v", err)
				}
				if err == io.EOF {
					break
				}

				err := hF.K8sObjectToResourceBlock(o)
				if err != nil {
					return fmt.Errorf("error adding k8s object to resource block: %v", err)
				}
			}
		}

		err = hF.WriteToFile(outputFile)
		if err != nil {
			return fmt.Errorf("error writing hcl to file %s; err = %v", outputFile, err)
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
}
