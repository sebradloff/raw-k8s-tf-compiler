package hclfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func GetK8sObjectsFromFile(filepath string) ([]*unstructured.Unstructured, error) {
	k8sObjects := []*unstructured.Unstructured{}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return k8sObjects, fmt.Errorf("ioutil.ReadFile(%s): %v", filepath, err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 4096)

	for {
		var o *unstructured.Unstructured
		// decode one yaml strem into a k8s object
		err = decoder.Decode(&o)
		if err != nil && err != io.EOF {
			return k8sObjects, fmt.Errorf("failed to decode portion of file (%s) into k8s object: %v", filepath, err)
		}
		if err == io.EOF {
			break
		}

		k8sObjects = append(k8sObjects, o)
	}

	return k8sObjects, nil
}

func GetK8sFilesToTransform(filePath string) ([]string, error) {
	var filesToTransform []string

	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing a path %q: %v", path, err)
		}
		if info.IsDir() {
			return nil
		}

		filesToTransform = append(filesToTransform, path)
		return nil
	})

	if err != nil {
		return filesToTransform, err
	}

	return filesToTransform, nil
}
