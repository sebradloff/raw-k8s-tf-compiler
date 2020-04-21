package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml2 "github.com/ghodss/yaml"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
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

		tfFile := hclwrite.NewEmptyFile()
		rootBody := tfFile.Body()
		// write hcl to tf file
		defer tfFile.Body().Clear()

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

				objectJSON, err := o.MarshalJSON()
				if err != nil {
					return fmt.Errorf("Failed to marshall one object into json: %v", err)
				}
				objectYaml, err := yaml2.JSONToYAML(objectJSON)
				if err != nil {
					return fmt.Errorf("Failed to marshall one object json into yaml: %v", err)
				}

				ns := o.GetNamespace()
				if ns == "" {
					ns = "default"
				}
				groupVersion := strings.Replace(o.GroupVersionKind().GroupVersion().String(), "/", "_", -1)
				resourceName := strings.Join([]string{ns, groupVersion, o.GetKind(), o.GetName()}, "-")

				contentBytes := []byte("<<EOT\n")
				contentBytes = append(contentBytes, objectYaml...)
				contentBytes = append(contentBytes, []byte("EOT\n")...)

				// create tf resource block
				resourceBlock := rootBody.AppendNewBlock("resource", []string{"k8s_manifest", resourceName})

				tokens := hclwrite.Tokens{

					{
						Type: hclsyntax.TokenTabs,
					},
					{
						Type:  hclsyntax.TokenCQuote,
						Bytes: []byte("content = "),
					},
					{
						Type:  hclsyntax.TokenOHeredoc,
						Bytes: contentBytes,
					},
					{
						Type: hclsyntax.TokenNewline,
					},
				}
				resourceBlock.Body().BuildTokens(tokens)
				resourceBlock.Body().AppendUnstructuredTokens(tokens)
				rootBody.AppendNewline()
			}
		}

		oF, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("could not os.Create(%s); err = %v", outputFile, err)
		}
		defer oF.Close()
		_, err = tfFile.WriteTo(oF)
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
