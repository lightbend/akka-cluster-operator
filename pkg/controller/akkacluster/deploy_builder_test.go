package akkacluster

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	appv1alpha1 "github.com/lightbend/akka-cluster-operator/pkg/apis/app/v1alpha1"
)

// go test -update  # to update the expected output files, aka golden files
var update = flag.Bool("update", false, "update golden files")

// firstVersioner
// Encode() will look at the resource type and look up a candidate table of kinds. For our purposes,
// we can just take the first one in the list as the desired version, as our resources have one entry.
// This implicit GroupVersioner is handy for testing, but for production one would be explicit.
type firstVersioner struct{}

func (firstVersioner) KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (target schema.GroupVersionKind, ok bool) {
	return kinds[0], true
}

// yamlizers returns decoder, encoder suitable for testing
func yamlizers() (func([]byte) runtime.Object, func(runtime.Object, io.Writer) error) {
	// add our custom resources like AkkaCluster et al to runtime Scheme
	appv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	// decode like the kubectl client (read anything)
	decode := scheme.Codecs.UniversalDeserializer().Decode

	// encode to yaml
	yaml, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		panic("no encoder for yaml")
	}

	// ...using first matching version registered in codec, based on object type lookup
	encoder := scheme.Codecs.EncoderForVersion(yaml.Serializer, firstVersioner{})

	// wrap decode encode
	return func(b []byte) runtime.Object {
		obj, kind, err := decode(b, nil, nil)
		if err != nil {
			panic(fmt.Sprintf("error decoding %v: %v", kind, err))
		}
		return obj
	}, encoder.Encode
}

// gold file tests: read input AkkaCluster, test for expected generated resources
func TestGenerateResources(t *testing.T) {
	decoder, encoder := yamlizers()

	inputSuffix := "_in.yaml"
	akkaClusterFiles, err := filepath.Glob("testdata/*" + inputSuffix)
	if err != nil || len(akkaClusterFiles) == 0 {
		t.Fatalf("failed to find testdata/*%s: %v", inputSuffix, err)
	}

	for _, inFile := range akkaClusterFiles {
		akkaClusterYaml, err := ioutil.ReadFile(inFile)
		if err != nil {
			t.Fatal(err.Error())
		}

		testName := inFile[:len(inFile)-len(inputSuffix)] // take off the suffix
		expectedOutputs, _ := filepath.Glob(testName + "_out_*")
		seenOutputs := map[string]bool{}

		obj := decoder(akkaClusterYaml)
		base, _ := obj.(*appv1alpha1.AkkaCluster)
		res := generateResources(base)

		res = append(res, base)
		for _, r := range res {
			var got bytes.Buffer
			err = encoder(r, &got)
			if err != nil {
				t.Fatalf("encoder error %s", err)
			}

			resourceType := reflect.ValueOf(r).Elem().Type().String()
			outFile := testName + "_out_" + resourceType + ".yaml"
			want, err := ioutil.ReadFile(outFile)
			if err != nil && !*update {
				t.Fatal(err.Error())
			}
			seenOutputs[outFile] = true
			if !bytes.Equal(want, got.Bytes()) {
				if *update {
					err := ioutil.WriteFile(outFile, got.Bytes(), 0666)
					if err != nil {
						t.Fatalf(err.Error())
					}
				} else {
					t.Errorf("%s mismatch:\n%s\n", outFile, cmp.Diff(want, got.Bytes()))
				}
			}
		}
		if !*update {
			// we've already tested that all generated outputs match the expected file,
			// now check the other way: make sure expected files were all generated
			for _, f := range expectedOutputs {
				if !seenOutputs[f] {
					t.Errorf("expected %s but none generated", f)
				}
			}
		}
	}
}
