package main

import "C"
import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

type Point struct {
	X, Y float64
}

func NewPoint(x, y float64) Point {
	return Point{x, y}
}

//export WrapNewPoint
func WrapNewPoint(x, y float64) unsafe.Pointer {
	mem := C.malloc(C.sizeof_double * 2)
	p := (*Point)(mem)
	*p = NewPoint(x, y)
	return mem
}

// calling this from dart will cause a runtime error
//
//export OldNewPoint
func OldNewPoint(x, y float64) unsafe.Pointer {
	p := NewPoint(x, y)
	return unsafe.Pointer(&p)
}

func AddPoint(p1, p2 Point) Point {
	return Point{p1.X + p2.X, p1.Y + p2.Y}
}

//export WrapAddPoint
func WrapAddPoint(p1, p2 unsafe.Pointer) unsafe.Pointer {
	mem := C.malloc(C.sizeof_double * 2)
	p := (*Point)(mem)
	*p = AddPoint(*(*Point)(p1), *(*Point)(p2))
	return mem
}

// calling this from dart will cause a runtime error
//
//export OldAddPoint
func OldAddPoint(p1, p2 unsafe.Pointer) unsafe.Pointer {
	p := AddPoint(*(*Point)(p1), *(*Point)(p2))
	return unsafe.Pointer(&p)
}

func MutatePoint(p *Point) {
	p.X = p.X * 2
	p.Y = p.Y * 2
}

//export WrapMutatePoint
func WrapMutatePoint(p unsafe.Pointer) {
	MutatePoint((*Point)(p))
}

func PointArray() []Point {
	arr := make([]Point, 3)
	return arr
}

//export WrapPointArray
func WrapPointArray() unsafe.Pointer {
	arr := PointArray()
	r := reflect.ValueOf(arr)
	s := binary.Size(r)
	mem := C.malloc(goint_to_csize_t(s))
	p := (*[]Point)(mem)
	*p = arr
	return mem
}

func Appendd(arr []Point, p Point) []Point {
	return append(arr, p)
}

//export WrapAppendd
func WrapAppendd(arr unsafe.Pointer, p unsafe.Pointer) unsafe.Pointer {
	arr_itn := (*[]Point)(arr)
	p_itn := (*Point)(p)
	res := Appendd(*arr_itn, *p_itn)
	return unsafe.Pointer(&res)
}

func PrintPointArray(arr []Point) {
	for _, p := range arr {
		println(p.X, p.Y)
	}
}

//export WrapPrintPointArray
func WrapPrintPointArray(arr unsafe.Pointer) {
	arr_itn := (*[]Point)(arr)
	PrintPointArray(*arr_itn)
}

func goint_to_csize_t(i int) C.size_t {
	s := make([]byte, 8)
	binary.LittleEndian.PutUint32(s, uint32(i))
	return C.size_t(binary.LittleEndian.Uint64(s))
}

func main() {
	fmt.Println("main")
}
