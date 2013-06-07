package arff_test

import (
	"fmt"
	"testing"
	"os"
	"io"

	"github.com/sbinet/go-arff"
)

func TestData(t *testing.T) {
	f, err := os.Open("testdata/iris.arff") 
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	dec, err := arff.NewDecoder(f)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	//fmt.Printf("arff.Header.Comment:\n%v\n", dec.Header.Comment)
	fmt.Printf("arff.Header.Relation: %v\n", dec.Header.Relation)
	fmt.Printf("arff.Header.Attrs:    %v\n", len(dec.Header.Attrs))
	for _, attr := range dec.Header.Attrs {
		fmt.Printf("\tname=%v, type=%v data=%v\n", attr.Name, attr.Type, attr.Data)
	}

	for {
		m := make(map[string]interface{})
		err = dec.Decode(m)
		if err != nil && err != io.EOF {
			t.Fatalf("error: %v\n", err)
		}
		if err == io.EOF {
			break
		}
		fmt.Printf("data= %v\n", m)
	}
}
