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

func (f *HCLFile) K8sObjectToResourceBlock(o *unstructured.Unstructured, pathToK8sFile string) error {
	ns := o.GetNamespace()
	if ns == "" {
		ns = "default"
	}
	groupVersion := strings.Replace(o.GetAPIVersion(), "/", "_", -1)
	resourceName := strings.Join([]string{ns, groupVersion, o.GetKind(), o.GetName()}, "-")

	resourceBlock := f.rootBody.AppendNewBlock("resource", []string{"k8s_manifest", resourceName})

	var tokens hclwrite.Tokens
	if pathToK8sFile != "" {
		tokens = hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(`file`),
			},
			{
				Type:  hclsyntax.TokenOParen,
				Bytes: []byte(`(`),
			},
			{
				Type:  hclsyntax.TokenOQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenTemplateInterp,
				Bytes: []byte(`${`),
			},
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(`path.module`),
			},
			{
				Type:  hclsyntax.TokenTemplateSeqEnd,
				Bytes: []byte(`}`),
			},
			{
				Type:  hclsyntax.TokenQuotedLit,
				Bytes: []byte(pathToK8sFile),
			},
			{
				Type:  hclsyntax.TokenCQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenCParen,
				Bytes: []byte(`)`),
			},
		}
	} else {
		oJSON, err := o.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshall one object into json: %v", err)
		}
		oYaml, err := yaml.JSONToYAML(oJSON)
		if err != nil {
			return fmt.Errorf("failed to marshall one object json into yaml: %v", err)
		}

		tokens = hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenOHeredoc,
				Bytes: []byte("<<EOT\n"),
			},
			{
				Type:  hclsyntax.TokenStringLit,
				Bytes: oYaml,
			},
			{
				Type:  hclsyntax.TokenCHeredoc,
				Bytes: []byte("EOT"),
			},
		}

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
