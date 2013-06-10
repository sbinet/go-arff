package arff

import (
	"bufio"
	"fmt"
	//"os"
	"bytes"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
)

var (
	errNoData  = fmt.Errorf("arff.Decode: no data")
	errComment = fmt.Errorf("arff.Decode: comment line")
)

type Decoder struct {
	Header Header

	r       *bufio.Reader
	scanner *bufio.Scanner
	state   state
	line    int

	data []interface{}

	// scratch space
	buf []byte
}

func NewDecoder(r io.Reader) (*Decoder, error) {
	rr, ok := r.(*bufio.Reader)
	if !ok {
		rr = bufio.NewReader(r)
	}

	dec := &Decoder{
		r:       rr,
		scanner: bufio.NewScanner(rr),
		state:   st_comment,
		line:    0,
	}
	err := dec.parse_header()
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (dec *Decoder) do_scan() bool {
	ok := dec.scanner.Scan()
	dec.line += 1
	return ok
}

func (dec *Decoder) Decode(v interface{}) error {
	var err error

	for dec.do_scan() {
		err = dec.parse_line(dec.scanner.Bytes())
		if err == errComment || err == errNoData {
			continue
		}
		break
	}

	if err == errNoData || err == errComment {
		return io.EOF
	}

	if err != nil {
		return fmt.Errorf("arff.Decode: error at line:%d: %v\n", dec.line, err)
	}

	if len(dec.data) != len(dec.Header.Attrs) {
		return fmt.Errorf("arff.Decode: line:%d: invalid data (expected #%d, got #%d)",
			dec.line,
			len(dec.Header.Attrs),
			len(dec.data),
		)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		mm, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf(
				"arff.Decode: invalid map type (expected map[string]interface{}, got %T)",
				v,
			)
		}
		for i, attr := range dec.Header.Attrs {
			mm[attr.Name] = dec.data[i]
		}

	case reflect.Ptr:
		// ok.

	default:
		return fmt.Errorf("arff.Decode: invalid interface type (expected a pointer to T, got %T)",
			v,
		)
	}
	rv = reflect.Indirect(rv)
	switch rv.Kind() {
	case reflect.Struct:
		attrs := dec.Header.Attrs
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
				return fmt.Errorf("arff.Decode: could not find a matching field for attribute [%s] in struct %T",
					attr.Name,
					rv.Interface(),
				)
			}
			field := rv.FieldByIndex(st.Index)
			if field.IsValid() {
				vv := reflect.ValueOf(dec.data[i])
				field.Set(vv)
			}
		}
	}
	dec.data = nil
	return nil
}

func (dec *Decoder) parse_header() error {
	var err error

	for dec.do_scan() {
		err = dec.scanner.Err()
		if err != nil {
			return err
		}

		data := dec.scanner.Bytes()
		data = bytes.TrimSpace(data)

		err = dec.parse_line(data)
		if err == errNoData {
			continue
		}

		if err != nil {
			return fmt.Errorf("arf.Decoder: parse error at line [%d]: %v", dec.line, err)
		}

		if len(data) == 0 {
			continue
		}

		if dec.state == st_in_data {
			break
		}
	}
	return err
}

func (dec *Decoder) parse_line(data []byte) error {
	var err error

	if len(data) == 0 {
		return errNoData
	}

	switch dec.state {
	case st_comment:
		if data[0] == '%' {
			dec.buf = append(dec.buf, data...)
			dec.buf = append(dec.buf, '\n')
			return nil
		} else {
			if dec.state == st_comment {
				dec.state = st_in_header
				dec.Header.Comment = string(dec.buf)
				dec.buf = nil
				return dec.parse_line(data)
			}
		}

	case st_in_header:
		//fmt.Printf("line[%d]: %q\n", dec.line, string(data))
		ldata := bytes.ToLower(data)
		if bytes.HasPrefix(ldata, tok_relation) {
			idx := bytes.Index(ldata, tok_relation)
			ll := bytes.Trim(data[idx+len(tok_relation):], " ")
			dec.Header.Relation = string(ll)
			return nil
		}

		if bytes.HasPrefix(ldata, tok_attribute) {
			idx := bytes.Index(ldata, tok_attribute)
			data = bytes.Trim(data[idx+len(tok_attribute):], " ")
			data = bytes.Replace(data, []byte("\t"), []byte(" "), -1)
			tmp := strings.Split(string(data), " ")
			var tokens []string
			for _, tok := range tmp {
				tok := strings.TrimSpace(tok)
				if len(tok) > 0 {
					//fmt.Printf(" -- adding [%v]\n", tok)
					tokens = append(tokens, tok)
				}
			}
			//fmt.Printf("tokens: %v\n", tokens)
			attr_name := tokens[0]
			attr_type := newAttrTypeFromString(tokens[1])
			switch attr_type {
			case Numeric, Real, Integer:
				dec.Header.Attrs = append(
					dec.Header.Attrs,
					Attr{
						Name: attr_name,
						Type: attr_type,
					},
				)
			case Nominal:
				toks := string(tokens[1])
				tmp := strings.Split(string(toks[1:len(toks)-1]), ",")
				values := make([]string, 0, len(tmp))
				for _, val := range tmp {
					values = append(values, strings.Trim(val, " \t"))
				}
				dec.Header.Attrs = append(
					dec.Header.Attrs,
					Attr{
						Name: attr_name,
						Type: attr_type,
						Data: values,
					},
				)

			case Invalid:
				return fmt.Errorf("invalid AttrType: %v", attr_type)

			default:
				return fmt.Errorf("unknown AttrType: %v", attr_type)
			}
			return nil
		}

		if bytes.HasPrefix(ldata, tok_data) {
			dec.state = st_in_data
			return nil
		}

	case st_in_data:
		//fmt.Printf("line:%d:%v\n", dec.line, string(data))
		if len(data) == 0 {
			//fmt.Printf("line[%d]: empty line!\n", dec.line)
			return errNoData
		}
		if data[0] == '%' {
			//fmt.Printf("line[%d]: comment\n", dec.line)
			return errComment
		}
		tmp := bytes.Split(data, []byte(","))
		values := make([]string, 0)
		for _, val := range tmp {
			v := strings.Trim(string(val), " \t")
			if len(v) > 0 {
				values = append(values, v)
			}
		}
		if len(values) != len(dec.Header.Attrs) {
			return fmt.Errorf(
				"invalid number of values (got=%d, expected=%d)",
				len(values),
				len(dec.Header.Attrs),
			)
		}
		datum := make([]interface{}, 0, len(values))
		for i, v := range values {
			attr := dec.Header.Attrs[i]
			switch attr.Type {
			case Numeric, Real:
				if v == "?" {
					datum = append(datum, math.NaN())
					continue
				}

				var vv float64
				vv, err = strconv.ParseFloat(v, 64)
				if err != nil {
					return fmt.Errorf(
						"float parsing failed (str=[%v]): %v", v, err,
					)
				}
				datum = append(datum, vv)

			case Integer:
				if v == "?" {
					datum = append(datum, math.NaN())
					continue
				}

				var vv int64
				vv, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"int parsing failed (str=[%v]): %v", v, err,
					)
				}
				datum = append(datum, vv)

			case Nominal:
				if is_in_str_slice(attr.Data, v) {
					datum = append(datum, v)
				} else if v == "?" {
					datum = append(datum, nil)
				} else {
					return fmt.Errorf("incorrect value [%s] for nominal attribute %s",
						v,
						attr.Name,
					)
				}

			default:
				panic("impossible")
			}
		}
		dec.data = datum

	}

	return err
}

// EOF
