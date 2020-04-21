package hcl

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type HCLFile struct {
	file     *hclwrite.File
	rootBody *hclwrite.Body
}

func NewHCLFile() *HCLFile {
	f := hclwrite.NewEmptyFile()
	return &HCLFile{
		file:     f,
		rootBody: f.Body(),
	}
}

func (f *HCLFile) GetFileBytes() []byte {
	return f.file.Bytes()
}

func (f *HCLFile) GetFileRootBody() *hclwrite.Body {
	return f.rootBody
}

func (f *HCLFile) K8sObjectToResourceBlock(o *unstructured.Unstructured) error {
	oJSON, err := o.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshall one object into json: %v", err)
	}
	oYaml, err := yaml.JSONToYAML(oJSON)
	if err != nil {
		return fmt.Errorf("Failed to marshall one object json into yaml: %v", err)
	}

	ns := o.GetNamespace()
	if ns == "" {
		ns = "default"
	}
	groupVersion := strings.Replace(o.GetAPIVersion(), "/", "_", -1)
	resourceName := strings.Join([]string{ns, groupVersion, o.GetKind(), o.GetName()}, "-")

	contentBytes := []byte("<<EOT\n")
	contentBytes = append(contentBytes, oYaml...)
	contentBytes = append(contentBytes, []byte("EOT\n")...)

	// create tf resource block
	resourceBlock := f.rootBody.AppendNewBlock("resource", []string{"k8s_manifest", resourceName})

	tokens := hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenOHeredoc,
			Bytes: contentBytes,
		},
	}

	bT := resourceBlock.Body().BuildTokens(tokens)
	resourceBlock.Body().SetAttributeRaw("content", bT)
	f.rootBody.AppendNewline()

	return nil
}

func (hF *HCLFile) WriteToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not os.Create(%s); err = %v", path, err)
	}
	defer f.Close()

	_, err = hF.file.WriteTo(f)
	if err != nil {
		return fmt.Errorf("could not write to file %s; err = %v", f.Name(), err)
	}
	return nil
}
