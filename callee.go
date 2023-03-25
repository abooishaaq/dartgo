package main

import "C"
import (
	"fmt"
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

func main() {
	fmt.Println("main")
}
