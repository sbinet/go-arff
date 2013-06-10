package arff

import (
	"bufio"
	"fmt"
	//"os"
	"bytes"
	"io"
	//"math"
	"reflect"
	//"strconv"
	"strings"
)

type Encoder struct {
	Header Header

	w       io.Writer
	state   state

}

func NewEncoder(w io.Writer) (*Encoder, error) {
	enc := &Encoder{
		w:w,
		state: st_comment,
	}
	return enc, nil
}

func (enc *Encoder) Encode(v interface{}) error {

	if enc.state != st_in_data {
		err := enc.init()
		if err != nil {
			return err
		}
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		type map_type map[string]interface{}
		var mm map_type
		ok := rv.Type().ConvertibleTo(reflect.TypeOf(mm))
		if !ok {
			return fmt.Errorf(
				"arff.Encode: invalid map type (expected map[string]interface{}, got %T)",
				v,
			)
		}
		rvv := rv.Convert(reflect.TypeOf(mm))
		mm = rvv.Interface().(map_type)
		data := make([]interface{}, len(enc.Header.Attrs))
		if len(data) != len(mm) {
			return fmt.Errorf(
				"arff.Encode: number of attributes don't match (got %d, expected %d)",
				len(mm), 
				len(data),
			)
		}
		for i, attr := range enc.Header.Attrs {
			vv,ok := mm[attr.Name]
			if !ok {
				return fmt.Errorf(
					"arff.Encode: missing key [%s] in input map",
					attr.Name,
				)
			}
			data[i] = vv
		}
		return enc.write(data)
		
	case reflect.Ptr:
		// ok.

	default:
		return fmt.Errorf(
			"arff.Encode: invalid interface type (expected *%T, got %T)",
			v, v,
		)
	}
	
	rv = reflect.Indirect(rv)
	switch rv.Kind() {
	case reflect.Struct:
		attrs := enc.Header.Attrs
		data := make([]interface{}, len(attrs))
		rt := rv.Type()
		for i, attr := range attrs {
			st, ok := rt.FieldByName(attr.Name)
			if !ok {
				nfields := rt.NumField()
				for ifield := 0; ifield < nfields; ifield++ {
					st = rt.Field(ifield)
					if st.Name == strings.ToTitle(attr.Name) {
						ok = true
						break
					}
					if st.Tag.Get("arff") == attr.Name {
						ok = true
						break
					}
				}
			}
			if !ok {
				return fmt.Errorf("arff.Encode: could not find a matching field for attribute [%s] in struct %T",
					attr.Name,
					rv.Interface(),
				)
			}
			field := rv.FieldByIndex(st.Index)
			data[i] = field.Interface()
		}
		return enc.write(data)
	}
	
	panic("unreachable")
}

func (enc *Encoder) init() error {
	hdr := &enc.Header
	var err error

	scanner := bufio.NewScanner(bytes.NewBufferString(hdr.Comment))
	for scanner.Scan() {
		_, err = fmt.Fprintf(
			enc.w,
			"%% %s\n",
			scanner.Text(),
		)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(enc.w, "\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(
		enc.w,
		"@relation %s\n\n",
		hdr.Relation,
	)
	if err != nil {
		return err
	}

	for _,attr := range hdr.Attrs {
		err = nil
		switch attr.Type {
		case Numeric, Real, Integer:
			_, err = fmt.Fprintf(
				enc.w,
				"@attribute %s %v\n",
				attr.Name,
				attr.Type,
			)
		case Nominal:
			names := strings.Join(attr.Data, ",")
			_, err = fmt.Fprintf(
				enc.w,
				"@attribute %s {%s}\n",
				attr.Name,
				names,
			)
		default:
			err = fmt.Errorf(
				"arff.Encode: invalid attribute type (%v)", attr.Type,
			)
		}
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(enc.w, "\n@data\n")
	if err != nil {
		return err
	}

	enc.state = st_in_data
	return err
}

func (enc *Encoder) write(data []interface{}) error {
	attrs := enc.Header.Attrs
	imax := len(attrs)
	for i,attr := range attrs {
		str := "%v,"
		if i == imax -1 {
			str = "%v"
		}
		_, err := fmt.Fprintf(
			enc.w,
			str,
			data[i],
		)
		if err != nil {
			return fmt.Errorf(
				"arff.Encode: error writing attribute [%s] (value=%v): %v",
				attr.Name,
				data[i],
				err,
			)
		}
	}
	_, err := fmt.Fprintf(enc.w, "\n")
	return err
}

// EOF
