go-arff
=======

`go-arff` is a simple pure-Go package to read and write Attribute-Relation File Format (`ARFF`) data files.

## Installation

```sh
$ go get github.com/sbinet/go-arff
```

## Documentation

http://godoc.org/github.com/sbinet/go-arff

## Example

### Writing `ARFF` files

```go
 f, err := os.Create("foo.arff")
 enc, err := arff.NewEncoder(f)
 enc.Header.Relation = "mydata"
 enc.Header.AddAttr("my-int", arff.Integer, nil)
 enc.Header.AddAttr("my-float", arff.Real, nil)
 enc.Header.AddAttr("my-string", arff.Nominal, []string{"a","b","c"})

 type S struct {
  I int     `arff:"my-int"`
  F float64 `arff:"my-float"`
  S string  `arff:"my-string"`
 }
 data := []S{.....}
 for _, v := range data {
   err = enc.Encode(&v)
 }
 f.Sync()
 f.Close()
```

### Reading `ARFF` files

```go
 f, err := os.Open("foo.arff")
 dec, err := arff.Decoder(f)

 type S struct {
  I int     `arff:"my-int"`
  F float64 `arff:"my-float"`
  S string  `arff:"my-string"`
 }

 for {
   var v S
   err = dec.Decode(&v)
   if err == io.EOF {
      break
   }
   fmt.Printf("I=%v F=%v S=%v\n", v.I, v.F, v.S)
 }
```


