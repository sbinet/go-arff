package arff_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/sbinet/go-arff"
)

func TestDecodeMap(t *testing.T) {
	f, err := os.Open("testdata/iris.arff")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	dec, err := arff.NewDecoder(f)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	o, err := os.Create("iris.map.txt")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}
	defer os.Remove(o.Name())

	//fmt.Printf("arff.Header.Comment:\n%v\n", dec.Header.Comment)
	fmt.Fprintf(o, "arff.Header.Relation: %v\n", dec.Header.Relation)
	fmt.Fprintf(o, "arff.Header.Attrs:    %v\n", len(dec.Header.Attrs))
	for _, attr := range dec.Header.Attrs {
		fmt.Fprintf(o, "\tname=%v, type=%v data=%v\n", attr.Name, attr.Type, attr.Data)
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
		fmt.Fprintf(
			o,
			"%v %v %v %v %v\n", 
			m["sepallength"], 
			m["sepalwidth"], 
			m["petallength"],
			m["petalwidth"],
			m["class"], 
		)
	}

	o.Sync()
	err = o.Close()
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	// test output file
	cmd := exec.Command(
		"diff", "-urN",
		"testdata/iris.ref.txt",
		o.Name(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecodeStruct(t *testing.T) {
	f, err := os.Open("testdata/iris.arff")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	dec, err := arff.NewDecoder(f)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	o, err := os.Create("iris.struct.txt")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}
	defer os.Remove(o.Name())
	
	//fmt.Printf("arff.Header.Comment:\n%v\n", dec.Header.Comment)
	fmt.Fprintf(o, "arff.Header.Relation: %v\n", dec.Header.Relation)
	fmt.Fprintf(o, "arff.Header.Attrs:    %v\n", len(dec.Header.Attrs))
	for _, attr := range dec.Header.Attrs {
		fmt.Fprintf(o, "\tname=%v, type=%v data=%v\n", attr.Name, attr.Type, attr.Data)
	}

	for {
		m := struct {
			Sepallength float64 `arff:"sepallength"`
			Sepalwidth  float64 `arff:"sepalwidth"`
			Petallength float64 `arff:"petallength"`
			Petalwidth  float64 `arff:"petalwidth"`
			Class       string  `arff:"class"`
		}{}
		err = dec.Decode(&m)
		if err != nil && err != io.EOF {
			t.Fatalf("error: %v\n", err)
		}
		if err == io.EOF {
			break
		}
		fmt.Fprintf(
			o, 
			"%v %v %v %v %v\n", 
			m.Sepallength, 
			m.Sepalwidth, 
			m.Petallength, 
			m.Petalwidth, 
			m.Class, 
		)
	}

	o.Sync()
	err = o.Close()
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	// test output file
	cmd := exec.Command(
		"diff", "-urN",
		"testdata/iris.ref.txt",
		o.Name(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	
}

func TestEncodeMap(t *testing.T) {
	o, err := os.Create("test.map.arff")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	//defer os.Remove(o.Name())

	enc, err := arff.NewEncoder(o)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	enc.Header.Relation = "simple"
	enc.Header.Comment = "a dummy comment\nanother one\n"
	enc.Header.AddAttr("a_int", arff.Integer, nil)
	enc.Header.AddAttr("a_float", arff.Real, nil)
	enc.Header.AddAttr("a_number", arff.Numeric, nil)

	type M map[string]interface{}
	for _,m := range []M{
		M{
			"a_int": 42,
			"a_float": 666.1,
			"a_number": 777.1,
		},
		M{
			"a_int": 43,
			"a_float": 666.2,
			"a_number": 777.2,
		},
	}{
		err = enc.Encode(m)
		if err != nil {
			t.Fatalf("error: %v\n", err)
		}
	}

	o.Sync()
	err = o.Close()
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	// test output file
	cmd := exec.Command(
		"diff", "-urN",
		"testdata/simple.ref.arff",
		o.Name(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	
}


func TestEncodeStruct(t *testing.T) {
	o, err := os.Create("test.struct.arff")
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	//defer os.Remove(o.Name())

	enc, err := arff.NewEncoder(o)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	enc.Header.Relation = "simple"
	enc.Header.Comment = "a dummy comment\nanother one\n"
	enc.Header.AddAttr("a_int", arff.Integer, nil)
	enc.Header.AddAttr("a_float", arff.Real, nil)
	enc.Header.AddAttr("a_number", arff.Numeric, nil)

	type S struct {
		I int `arff:"a_int"`
		F float64 `arff:"a_float"`
		N float64 `arff:"a_number"`
	}
	for _,m := range []S{
		{
			I: 42,
			F: 666.1,
			N: 777.1,
		},
		{
			I: 43,
			F: 666.2,
			N: 777.2,
		},
	}{
		err = enc.Encode(&m)
		if err != nil {
			t.Fatalf("error: %v\n", err)
		}
	}

	o.Sync()
	err = o.Close()
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	// test output file
	cmd := exec.Command(
		"diff", "-urN",
		"testdata/simple.ref.arff",
		o.Name(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

