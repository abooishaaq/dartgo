// parse a library and generate a cgo file callable from dart
// just like
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
	"strings"
)

// func NewPoint(x, y float64) Point {
// 	return Point{x, y}
// }

// //export WrapNewPoint
// func WrapNewPoint(x, y float64) unsafe.Pointer {
// 	mem := C.malloc(C.sizeof_double * 2)
// 	p := (*Point)(mem)
// 	*p = NewPoint(x, y)
// 	return mem
// }

// func AddPoint(p1, p2 Point) Point {
// 	return Point{p1.X + p2.X, p1.Y + p2.Y}
// }

// //export WrapAddPoint
// func WrapAddPoint(p1, p2 unsafe.Pointer) unsafe.Pointer {
// 	mem := C.malloc(C.sizeof_double * 2)
// 	p := (*Point)(mem)
// 	*p = AddPoint(*(*Point)(p1), *(*Point)(p2))
// 	return mem
// }

const ptrSize = "unsafe.Sizeof(unsafe.Pointer(nil))"

const prelude = `
package main

import "C"
import (
	"encoding/binary"
	"unsafe"
)

func goint_to_csize_t(i int) C.size_t {
    s := make([]byte, 8)
    binary.LittleEndian.PutUint32(s, uint32(i))
    return C.size_t(binary.LittleEndian.Uint64(s))
}

`

func calcSizeOfStruct(s *ast.StructType) int {
	size := 0
	for _, field := range s.Fields.List {
		var fieldSize int
		switch field.Type.(type) {
		case *ast.Ident:
			switch field.Type.(*ast.Ident).Name {
			case "int":
				fieldSize = 8
			case "float64":
				fieldSize = 8
			}
		}
		fieldSize *= len(field.Names)
		size += fieldSize
	}
	return size
}

func intermediateName(boxed bool, name string) string {
	if boxed {
		return name + "_ptr"
	}
	return name
}

func zip(a, b []string) [][]string {
	if len(a) != len(b) {
		panic("zip: unequal lengths")
	}
	var res [][]string
	for i := range a {
		res = append(res, []string{a[i], b[i]})
	}
	return res
}

func mapFunc[T any, U any](f func(T) U, s []T) []U {
	var res []U
	for _, x := range s {
		res = append(res, f(x))
	}
	return res
}

func flatten[T any](s [][]T) []T {
	var res []T
	for _, x := range s {
		res = append(res, x...)
	}
	return res
}

func zipWith[T any, U any, V any](f func(T, U) V, a []T, b []U) []V {
	if len(a) != len(b) {
		panic("zipWith: unequal lengths")
	}
	var res []V
	for i := range a {
		res = append(res, f(a[i], b[i]))
	}
	return res
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("expected a single argument")
	}
	file, err := parser.ParseFile(token.NewFileSet(), args[0], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	boxedTypes := make(map[string]bool)
	boxedSizee := make(map[string]string)

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					switch typeSpec.Type.(type) {
					case *ast.StructType:
						s := typeSpec.Type.(*ast.StructType)
						boxedTypes[typeSpec.Name.Name] = true
						boxedSizee[typeSpec.Name.Name] = "goint_to_csize_t(" + strconv.Itoa(calcSizeOfStruct(s)) + ")"
					case *ast.ArrayType:
						boxedTypes[typeSpec.Name.Name] = true
						boxedSizee[typeSpec.Name.Name] = ptrSize
					default:
						boxedTypes[typeSpec.Name.Name] = false
					}
				}
			}
		}
	}

	var cgoFile string

	cgoFile += prelude

	isBoxedType := func(t ast.Expr) bool {
		switch t.(type) {
		case *ast.Ident:
			return boxedTypes[t.(*ast.Ident).Name]
		case *ast.ArrayType:
			return true
		case *ast.StarExpr:
			return true
		}
		return false
	}

	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			f := decl.(*ast.FuncDecl)
			// generate a wrapper function
			// that allocates memory for the boxed type
			// and returns a pointer to it
			// and a function that frees the memory
			var wrapperFn string
			// exact signature of the function expect the return type(s)
			fnName := "Wrap__" + f.Name.Name
			wrapperFn += "//export " + fnName + "\n"
			wrapperFn += "func " + fnName + "("
			for i, param := range f.Type.Params.List {
				if i > 0 {
					wrapperFn += ", "
				}

				for j, name := range param.Names {
					if j > 0 {
						wrapperFn += ", "
					}
					wrapperFn += name.Name + " "
				}

				if isBoxedType(param.Type) {
					wrapperFn += "unsafe.Pointer"
				} else {
					switch param.Type.(type) {
					case *ast.Ident:
						wrapperFn += param.Type.(*ast.Ident).Name
					}
				}
			}
			wrapperFn += ") "
			if f.Type.Results != nil {
				wrapperFn += "("
				for i, result := range f.Type.Results.List {
					if i > 0 {
						wrapperFn += ", "
					}
					if len(result.Names) > 0 {
						wrapperFn += result.Names[0].Name + "_arg "
					}

					if isBoxedType(result.Type) {
						wrapperFn += "unsafe.Pointer"
					} else {
						switch result.Type.(type) {
						case *ast.Ident:
							wrapperFn += result.Type.(*ast.Ident).Name
						}
					}
				}
				wrapperFn += ")"
			}
			wrapperFn += " {\n"

			// all return values
			var retNames [][]string
			var retboxed [][]bool
			unamedRetCount := 0
			if f.Type.Results != nil {
				for _, result := range f.Type.Results.List {
					count := len(result.Names)
					if count > 0 {
						rets := make([]string, count)
						retboxed := make([]bool, count)
						for _, name := range result.Names {
							if isBoxedType(result.Type) {
								rets = append(rets, name.Name)
								retboxed = append(retboxed, true)
							} else {
								rets = append(rets, name.Name)
								retboxed = append(retboxed, false)
							}
						}
					} else {
						unamedRetCount++
						retName := "ret" + strconv.Itoa(unamedRetCount)
						rets := []string{retName}
						retbox := []bool{false}
						if isBoxedType(result.Type) {
							retbox[0] = true
						}
						retNames = append(retNames, rets)
						retboxed = append(retboxed, retbox)
					}
				}
			}

			var callParams []string
			for _, param := range f.Type.Params.List {
				for _, name := range param.Names {
					if isBoxedType(param.Type) {
						switch param.Type.(type) {
						case *ast.Ident:
							tyname := param.Type.(*ast.Ident).Name
							callParams = append(callParams, "*(*"+tyname+")("+name.Name+")")
						case *ast.StarExpr:
							tyname := param.Type.(*ast.StarExpr).X.(*ast.Ident).Name
							callParams = append(callParams, "(*"+tyname+")("+name.Name+")")
						}
					} else {
						callParams = append(callParams, name.Name)
					}
				}
			}

			if f.Type.Results != nil {
				wrapperFn += strings.Join(flatten(retNames), ", ") + " := " + f.Name.Name + "(" + strings.Join(callParams, ", ") + ")\n"
				i := 0
				for _, result := range f.Type.Results.List {
					lim := len(result.Names)
					if lim == 0 {
						lim = 1
					}
					for j := 0; j < lim; j++ {
						// allocate memory for the boxed type
						switch result.Type.(type) {
						case *ast.Ident:
							tyname := result.Type.(*ast.Ident).Name
							if boxedTypes[tyname] {
								wrapperFn += retNames[i][j] + "_ptr := C.malloc(" + boxedSizee[tyname] + ")\n"
							}
						}
					}
					for j := 0; j < lim; j++ {
						if isBoxedType(result.Type) {
							switch result.Type.(type) {
							case *ast.Ident:
								tyname := result.Type.(*ast.Ident).Name
								if boxedTypes[tyname] {
									wrapperFn += "*(*" + tyname + ")(" + retNames[i][j] + "_ptr) = " + retNames[i][j] + "\n"
								}
							case *ast.StarExpr:
								switch result.Type.(*ast.StarExpr).X.(type) {
								case *ast.Ident:
									tyname := result.Type.(*ast.StarExpr).X.(*ast.Ident).Name
									wrapperFn += "*(*" + tyname + ")(" + retNames[i][j] + "_ptr) = *" + retNames[i][j] + "\n"
								}
							}
						}
					}
					wrapperFn += "return " + strings.Join(zipWith[bool, string, string](intermediateName, flatten[bool](retboxed), flatten[string](retNames)), ", ") + "\n"
				}
			} else {
				wrapperFn += f.Name.Name + "(" + strings.Join(callParams, ", ") + ")\n"
			}
			wrapperFn += "}\n\n"
			cgoFile += wrapperFn
		}
	}

	fmt.Println(cgoFile)
}
