
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

//export Wrap__F_new_point
func Wrap__F_new_point(x , y float64) (unsafe.Pointer) {
ret1 := F_new_point(x, y)
ret1_ptr := C.malloc(goint_to_csize_t(16))
*(*P)(ret1_ptr) = ret1
return ret1_ptr
}

//export Wrap__F_add_point
func Wrap__F_add_point(p1 , p2 unsafe.Pointer) (unsafe.Pointer) {
ret1 := F_add_point(*(*P)(p1), *(*P)(p2))
ret1_ptr := C.malloc(goint_to_csize_t(16))
*(*P)(ret1_ptr) = ret1
return ret1_ptr
}

//export Wrap__F_mutate_point
func Wrap__F_mutate_point(p unsafe.Pointer)  {
F_mutate_point((*P)(p))
}

//export Wrap__main
func Wrap__main()  {
main()
}


