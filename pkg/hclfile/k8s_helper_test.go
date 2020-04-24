package hclfile_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sebradloff/rawk8stfc/pkg/hclfile"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetK8sObjectsFromFile(t *testing.T) {
	t.Run("give one k8s object in file, return object in slice", func(t *testing.T) {
		want := []*unstructured.Unstructured{}
		td, teardown := createTempDir(t, "")
		defer teardown()

		o := testObject("test")
		want = append(want, &o)

		fileName := createFile(t, td, "test")
		addK8sObjectToFile(t, &o, fileName)

		got, err := hclfile.GetK8sObjectsFromFile(fileName)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if len(got) != len(want) {
			t.Errorf("got %d objects; wanted %d", len(got), len(want))
		}
	})

	t.Run("give multiple k8s objects in file, return all objects in slice", func(t *testing.T) {
		want := []*unstructured.Unstructured{}
		td, teardown := createTempDir(t, "")
		defer teardown()

		fileName := createFile(t, td, "test")

		for i := 0; i < 4; i++ {
			name := fmt.Sprintf("test-%d", i)
			o := testObject(name)
			want = append(want, &o)
			addK8sObjectToFile(t, &o, fileName)
		}

		got, err := hclfile.GetK8sObjectsFromFile(fileName)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if len(got) != len(want) {
			t.Errorf("got %d objects; wanted %d", len(got), len(want))
		}
	})

	t.Run("give non k8s object, return error", func(t *testing.T) {
		want := []*unstructured.Unstructured{}
		td, teardown := createTempDir(t, "")
		defer teardown()

		fileName := createFile(t, td, "test")

		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()

		testBytes := []byte("not a k8s object")

		_, err = f.Write(testBytes)
		if err != nil {
			t.Fatalf("failed to write object into file: %v", err)
		}

		got, err := hclfile.GetK8sObjectsFromFile(fileName)
		if err == nil {
			t.Errorf("want err; got nil")
		}

		if !strings.Contains(err.Error(), "failed to decode portion of file") {
			t.Errorf("want err to have substring 'failed to decode portion of file'; got = %v", err)
		}

		if len(got) != len(want) {
			t.Errorf("got %d objects; wanted %d", len(got), len(want))
		}
	})

	t.Run("give non existant file, return error", func(t *testing.T) {
		want := []*unstructured.Unstructured{}
		td, teardown := createTempDir(t, "")

		fileName := createFile(t, td, "test")
		// removing file
		teardown()

		got, err := hclfile.GetK8sObjectsFromFile(fileName)
		if err == nil {
			t.Errorf("want err; got nil")
		}

		if !strings.Contains(err.Error(), "ioutil.ReadFile") {
			t.Errorf("want err to have substring 'ioutil.ReadFile'; got = %v", err)
		}

		if len(got) != len(want) {
			t.Errorf("got %d objects; wanted %d", len(got), len(want))
		}
	})
}

func TestGetK8sFilesToTransform(t *testing.T) {
	t.Run("given one file, return file in slice", func(t *testing.T) {
		td, teardown := createTempDir(t, "")
		defer teardown()

		want := []string{}
		testFilePath := createFile(t, td, "test")
		want = append(want, testFilePath)

		got, err := hclfile.GetK8sFilesToTransform(testFilePath)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})

	t.Run("given a directory with one file, return file in slice", func(t *testing.T) {
		td, teardown := createTempDir(t, "")
		defer teardown()

		want := []string{}
		testFilePath := createFile(t, td, "test")
		want = append(want, testFilePath)

		got, err := hclfile.GetK8sFilesToTransform(td)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})

	t.Run("given a directory with multiple files, return files in slice", func(t *testing.T) {
		td, teardown := createTempDir(t, "")
		defer teardown()

		want := []string{}

		for i := 0; i < 4; i++ {
			testFilePath := createFile(t, td, strconv.Itoa(i))
			want = append(want, testFilePath)
		}

		got, err := hclfile.GetK8sFilesToTransform(td)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})

	t.Run("given a directory with multiple subfolder with multiple files, return files in slice", func(t *testing.T) {
		tdTop, teardown := createTempDir(t, "")
		defer teardown()

		tdA, _ := createTempDir(t, tdTop)
		tdB, _ := createTempDir(t, tdTop)

		want := []string{}

		for i := 0; i < 2; i++ {
			testFilePath := createFile(t, tdA, strconv.Itoa(i))
			want = append(want, testFilePath)

			testFilePath = createFile(t, tdB, strconv.Itoa(i))
			want = append(want, testFilePath)
		}

		got, err := hclfile.GetK8sFilesToTransform(tdTop)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})

	t.Run("given a directory with no files, return empty slice", func(t *testing.T) {
		td, teardown := createTempDir(t, "")
		defer teardown()

		want := []string{}

		got, err := hclfile.GetK8sFilesToTransform(td)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})

	t.Run("given a non existant directory, return empty slice and error", func(t *testing.T) {
		td, teardown := createTempDir(t, "")
		// teardown immediately so directory doesn't exist
		teardown()

		want := []string{}

		got, err := hclfile.GetK8sFilesToTransform(td)
		if err == nil {
			t.Fatal("got nil; want err")
		}

		if !strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("error should have contained no such file or directory; err = %v", err)
		}

		if !Equal(got, want) {
			t.Errorf("got = %s; want = %s", got, want)
		}
	})
}

func createFile(t *testing.T, dirPath string, name string) string {
	testFilePath := filepath.Join(dirPath, fmt.Sprintf("%s.yaml", name))
	_, err := os.Create(testFilePath)
	if err != nil {
		t.Fatalf("could not os.Create(%s); err = %v", testFilePath, err)
	}
	return testFilePath
}

func addK8sObjectToFile(t *testing.T, o *unstructured.Unstructured, fileName string) {
	oJSON, err := o.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshall object into json: %v", err)
	}
	oYaml, err := yaml.JSONToYAML(oJSON)
	if err != nil {
		t.Fatalf("failed to transform object json into yaml: %v", err)
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	seperator := []byte("---\n")

	oYaml = append(oYaml, seperator...)
	_, err = f.Write(oYaml)
	if err != nil {
		t.Fatalf("failed to write object into file: %v", err)
	}
}

func createTempDir(t *testing.T, startDir string) (string, func()) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	name := strconv.Itoa(r.Intn(6))
	td, err := ioutil.TempDir(startDir, name)
	if err != nil {
		t.Fatalf("could not create test dir %s; err = %v", td, err)
	}
	return td, func() { os.RemoveAll(td) }
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
