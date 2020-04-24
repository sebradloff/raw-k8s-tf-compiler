package cmd

import (
	"fmt"
	"log"

	"github.com/sebradloff/rawk8stfc/pkg/hclfile"
	"github.com/spf13/cobra"
)

var fileReferenceCmd = &cobra.Command{
	Use:   "file-reference",
	Short: "Create tf resource block for each kubernetes resource specified by filename or directory with content referecing the file in which the resource resides",
	RunE: func(cmd *cobra.Command, args []string) error {
		filesToTransform, err := hclfile.GetK8sFilesToTransform(k8sFile)
		if err != nil {
			return fmt.Errorf("GetK8sFilesToTransform(%s): %v", k8sFile, err)
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
			if len(k8sObjects) > 1 {
				log.Printf("will inline resources in file %s, since terraform-provider-k8s can not handle mutiple objects per resource block", f)
				inlineOverideNeeded = true
			}

			for _, o := range k8sObjects {
				if inlineOverideNeeded {
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
