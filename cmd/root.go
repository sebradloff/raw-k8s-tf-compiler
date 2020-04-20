package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	k8sFile    string
	outputFile string
)

func TestCmd() *cobra.Command {
	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "rawk8stfc",
	Short: "A tool to create tf resources for all k8s objects inputed",
	RunE: func(cmd *cobra.Command, args []string) error {
		var filesToTransform []string

		fi, err := os.Stat(k8sFile)
		if err != nil {
			return fmt.Errorf("could not get os.Stat(%s); err = %v", k8sFile, err)
		}
		if fi.Mode().IsDir() {
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

		type resource struct {
			K8sManifest string `hcl:"k8s_manifest,label"`
			Name        string `hcl:"name,label"`
			Content     string `hcl:"content"`
		}
		type tfStruct struct {
			Resources []resource `hcl:"resource,block"`
		}

		var tfRes tfStruct

		for _, f := range filesToTransform {
			fmt.Println("****" + f)

			data, err := ioutil.ReadFile(f)
			if err != nil {
				return fmt.Errorf("could not ioutil.ReadFile(%s); err = %v", f, err)
			}

			var objects []*unstructured.Unstructured
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)

			// allows us to capture yaml streams
			for {
				var object *unstructured.Unstructured

				err = decoder.Decode(&object)
				if err != nil && err != io.EOF {
					return fmt.Errorf("Failed to unmarshal manifest: %s", err)
				}
				if err == io.EOF {
					break
				}

				objects = append(objects, object)
			}

			for _, v := range objects {
				res := resource{
					K8sManifest: "k8s_manifest",
					Name:        v.GetName(),
					Content:     "cool2",
				}

				tfRes.Resources = append(tfRes.Resources, res)
			}
		}

		f := hclwrite.NewEmptyFile()
		gohcl.EncodeIntoBody(&tfRes, f.Body())
		defer f.Body().Clear()

		oF, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("could not os.Create(%s); err = %v", outputFile, err)
		}
		defer oF.Close()
		_, err = f.WriteTo(oF)
		if err != nil {
			return fmt.Errorf("could not write to file %s; err = %v", oF.Name(), err)
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
