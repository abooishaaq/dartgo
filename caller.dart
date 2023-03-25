import 'dart:ffi' as ffi;
import 'dart:io' show Platform;

// Define the Go struct
class GoPoint extends ffi.Struct {
  @ffi.Double()
  external double x;

  @ffi.Double()
  external double y;
}

// Define the FFI functions
typedef new_point_func = ffi.Pointer<GoPoint> Function(
    ffi.Double x, ffi.Double y);
typedef NewPoint = ffi.Pointer<GoPoint> Function(double x, double y);

typedef add_point_func = ffi.Pointer<GoPoint> Function(
    ffi.Pointer<GoPoint> p1, ffi.Pointer<GoPoint> p2);
typedef AddPoint = ffi.Pointer<GoPoint> Function(
    ffi.Pointer<GoPoint> p1, ffi.Pointer<GoPoint> p2);

typedef mutate_point_func = ffi.Void Function(ffi.Pointer<GoPoint> p);
typedef MutatePoint = void Function(ffi.Pointer<GoPoint> p);

void main() {
  // Load the shared library
  var libraryPath = './callee.so';
  if (Platform.isMacOS) {
    libraryPath = './callee.dylib';
  } else if (Platform.isWindows) {
    libraryPath = 'callee.dll';
  }
  final dylib = ffi.DynamicLibrary.open(libraryPath);

  // Look up the "NewPoint" function
  final NewPoint newPoint = dylib
      .lookup<ffi.NativeFunction<new_point_func>>('WrapNewPoint')
      .asFunction();
  final AddPoint addPoint = dylib
      .lookup<ffi.NativeFunction<add_point_func>>('WrapAddPoint')
      .asFunction();
  final MutatePoint mutatePoint = dylib
      .lookup<ffi.NativeFunction<mutate_point_func>>('WrapMutatePoint')
      .asFunction();

  // Call the "NewPoint" function and get the result as a Dart struct
  final goPoint = newPoint(2.0, 3.0);
  final goPoint2 = newPoint(4.0, 5.0);

  // Print the result
  print("point 1: ${goPoint.ref.x}, ${goPoint.ref.y}");
  print("point 2: ${goPoint2.ref.x}, ${goPoint2.ref.y}");

  var added = addPoint(goPoint, goPoint2);
  print("added  : ${added.ref.x}, ${added.ref.y}");

  mutatePoint(added);
  print("mutated: ${added.ref.x}, ${added.ref.y}");

  final NewPoint NewPointFail = dylib
      .lookup<ffi.NativeFunction<new_point_func>>('OldNewPoint')
      .asFunction();
  NewPointFail(2.0, 3.0);
}
