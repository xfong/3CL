package cl

// This file provides CGO flags to find OpecnCL libraries and headers.

//#cgo darwin LDFLAGS: -framework OpenCL
//#cgo !darwin LDFLAGS: -lOpenCL
//
////default location:
//#cgo LDFLAGS:-L/usr/lib/x86_64-linux-gnu/
//#cgo LDFLAGS:-L/usr/local/cuda/lib64
//#cgo LDFLAGS:-L/usr/local/cuda/lib
//#cgo CFLAGS: -I/usr/include
//#cgo CFLAGS: -I/usr/local/clFFT/include
//#cgo CFLAGS: -I/usr/local/cuda/include
//
////Ubuntu 15.04::
//#cgo LDFLAGS:-L/usr/lib/x86_64-linux-gnu/
//#cgo CFLAGS: -I/usr/include
//#cgo CFLAGS: -I/usr/local/clFFT/include
//
////arch linux:
//#cgo LDFLAGS:-L/opt/lib64 -L/opt/lib
//#cgo CFLAGS: -I/opt/include
//
////WINDOWS:
//#cgo windows LDFLAGS:-LC:/Intel/opencl/lib/x64
//#cgo windows LDFLAGS:-LC:/Stuff/proj/clMathLibraries/clFFT/v2.12.2/bin
//#cgo windows LDFLAGS:-l clFFT
//#cgo windows CFLAGS: -IC:/Intel/opencl/include
//#cgo windows CFLAGS: -IC:/Stuff/proj/clMathLibraries/clFFT/v2.12.2/include
import "C"
