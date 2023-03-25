# dart-go interop

exploring how to call go from dart

## run

- `go build -buildmode=c-shared -o callee.so callee.go`
- `dart run caller.dart`
