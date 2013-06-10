package arff

import (
	"fmt"
	"reflect"
	"strings"
)

type AttrType int

const (
	Invalid AttrType = iota
	Numeric
	Real
	Integer
	String
	Nominal
)

func newAttrTypeFromString(s string) AttrType {
	switch strings.ToLower(s) {
	case "numeric":
		return Numeric
	case "real":
		return Real
	case "integer":
		return Integer
	case "string":
		return String
	case "nominal":
		return Nominal
	default:
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			return Nominal
		}
	}
	fmt.Printf("invalid AttrType string: [%v]\n", s)
	return Invalid
}

func (a AttrType) String() string {
	switch a {
	case Numeric:
		return "Numeric"
	case Real:
		return "Real"
	case Integer:
		return "Integer"
	case String:
		return "String"
	case Nominal:
		return "Norminal"
	default:
		err := fmt.Errorf("arff: invalid AttrType (value=%d)", int(a))
		panic(err)
	}
	panic("unreachable")
}

type state int

const (
	st_comment state = iota
	st_in_header
	st_in_data
)

var (
	tok_relation  = []byte("@relation")
	tok_attribute = []byte("@attribute")
	tok_data      = []byte("@data")
)

type Header struct {
	Comment  string
	Relation string
	Attrs    []Attr
}

type Attr struct {
	Name string
	Type AttrType
	Data []string
}

// NewHeaderFrom returns a Header with the Attrs automatically generated
// from the map or ptr-to-struct v
func NewHeaderFrom(v interface{}) (*Header, error) {
	hdr := Header{
		Comment:  "no comment",
		Relation: "mydata",
		Attrs:    nil,
	}
	err := hdr.init(v)
	if err != nil {
		return nil, err
	}
	return &hdr, nil
}

func (hdr *Header) AddAttr(name string, atype AttrType, data []string) error {
	// TODO: check for duplicates
	hdr.Attrs = append(
		hdr.Attrs,
		Attr{
			Name: name,
			Type: atype,
			Data: make([]string, len(data)),
		},
	)
	if len(data) > 0 {
		copy(hdr.Attrs[len(hdr.Attrs)-1].Data, data)
	}
	return nil
}

func (hdr *Header) init(v interface{}) error {
	var err error
	type attr_t struct {
		Name string
		Type reflect.Type
	}
	v_attrs := make([]attr_t, 0)

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		mm, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf(
				"arff.Header: invalid map type (expected map[string]interface{}, got %T)",
				v,
			)
		}
		for k := range mm {
			rkv := reflect.ValueOf(mm[k])
			v_attrs = append(v_attrs, attr_t{Name: k, Type: rkv.Type()})
		}

	case reflect.Ptr:
		// ok.
		rv = reflect.Indirect(rv)
		switch rv.Kind() {
		case reflect.Struct:
			rt := rv.Type()
			nfields := rt.NumField()
			for i := 0; i < nfields; i++ {
				f := rt.Field(i)
				v_attrs = append(v_attrs, attr_t{Name: f.Name, Type: f.Type})
			}
		default:
			return fmt.Errorf("arff.Header: invalid type (expected a pointer to struct, got %T)",
				v,
			)

		}

	default:
		return fmt.Errorf("arff.Header: invalid interface type (expected a pointer to T, got %T)",
			v,
		)
	}

	for _, vattr := range v_attrs {
		switch vattr.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			hdr.Attrs = append(
				hdr.Attrs,
				Attr{
					Name: vattr.Name,
					Type: Integer,
					Data: nil,
				},
			)
		case reflect.Float32, reflect.Float64:
			hdr.Attrs = append(
				hdr.Attrs,
				Attr{
					Name: vattr.Name,
					Type: Real,
					Data: nil,
				},
			)

		default:
			return fmt.Errorf(
				"arff.Header: unhandled type (%s) (valids are: floats, ints)",
				vattr.Type,
			)
		}
	}
	return err
}

// EOF
