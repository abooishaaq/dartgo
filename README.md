# dart-go interop

exploring how to call go from dart. 
a naive GOgen code in `gen.go` which outputs go wrapper code which can be called using `dart:ffi`.

## caller/callee

- `go build -buildmode=c-shared -o callee.so callee.go`
- `dart run caller.dart`

## go gen

- `go build gen.go`
- `./gen lib.go > generated.go`
- `go build -buildmode=c-shared -o lib.so lib.go generated.go`
- `dart run libcaller.dart`

## hurdles

*how to handle arrays?*

basic struct types can be allocated on the heap and passed to dart.
we can't access go heap from dart as that can be GCed.
as for arrays in go then they can be appended to without us knowing.

*platform independence?*

*open files?*

*how to handle errors?*

*generics?*
