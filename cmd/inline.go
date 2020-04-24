package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sebradloff/rawk8stfc/pkg/hclfile"
	"github.com/spf13/cobra"
)

var inlineCmd = &cobra.Command{
	Use:   "inline",
	Short: "Create tf resource block with content inline for each kubernetes resource specified by filename or directory",
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

			for _, o := range k8sObjects {
				err := hF.AddK8sObjectToResourceBlockContentInline(o)
				if err != nil {
					return fmt.Errorf("AddK8sObjectToResourceBlockContentInline called with object name (%s): %v", o.GetName(), err)
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
