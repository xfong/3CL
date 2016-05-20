// Package opencl provides GPU interaction
package opencl

import (
	"fmt"
//	"log"
	"runtime"

	"github.com/mumax/3cl/opencl/cl"
	"github.com/mumax/3cl/opencl/kernels"
//	"github.com/mumax/3cl/util"
)

var (
	Version     	string			// OpenCL version
	DevName     	string     		// GPU name
	TotalMem    	int64      		// total GPU memory
	PlatformInfo	string     		// Human-readable OpenCL platform description
	GPUInfo     	string     		// Human-readable GPU description
	Synchronous 	bool       		// for debug: synchronize stream0 at every kernel launch
	ClPlatforms     []*cl.Platform          // list of platforms available
	ClPlatform      *cl.Platform            // platform the global OpenCL context is attached to
	ClDevices       []*cl.Device            // list of devices global OpenCL context may be associated with
	ClDevice        *cl.Device              // device associated with global OpenCL context
	ClCtx       	*cl.Context 		// global OpenCL context
	ClCmdQueue      *cl.CommandQueue        // command queue attached to global OpenCL context
	ClProgram   	*cl.Program		// handle to program in the global OpenCL context
	KernList    	map[string]*cl.Kernel	// Store pointers to all compiled kernels
	initialized     = false                 // Initial state defaults to false
)

// Locks to an OS thread and initializes CUDA for that thread.
func Init(gpu, platformId int) {
	if (initialized) {
		fmt.Printf("Already initialized \n")
		return // needed for tests
	}

	runtime.LockOSThread()
	platforms, err := cl.GetPlatforms()
	if err != nil {
		fmt.Printf("Failed to get platforms: %+v \n", err)
	}

	fmt.Printf("// Platform %d: \n", platformId)
	platform := platforms[platformId]

	PlatformName := platform.Name()
	PlatformVendor := platform.Vendor()
	PlatformProfile := platform.Profile()
	PlatformVersion := platform.Version()
	PlatformInfo = fmt.Sprint("//   Name: ", PlatformName, "\n//   Vendor: ", PlatformVendor, "\n//   Profile: ", PlatformProfile, "\n//   Version: ", PlatformVersion,"\n")
	ClPlatforms = platforms
	ClPlatform = platform
	
	devices, err := platform.GetDevices(cl.DeviceTypeGPU)
	if err != nil {
		fmt.Printf("Failed to get devices: %+v \n", err)
		return
	}
	if len(devices) == 0 {
		fmt.Printf("GetDevices returned no devices \n")
		return
	}
	deviceIndex := -1

	if gpu < len(devices) {
	        deviceIndex = gpu
	} else {
	        fmt.Println("GPU choice not selectable... falling back to first GPU found!")
		deviceIndex = 0
	}

	if deviceIndex < 0 {
	   	deviceIndex = 0
	}

	DevName = devices[deviceIndex].Name()
	TotalMem = devices[deviceIndex].GlobalMemSize()
	Version = devices[deviceIndex].OpenCLCVersion()
	GPUInfo = fmt.Sprint("OpenCL C Version ", Version, "\n// GPU: ", DevName, "(", (TotalMem)/(1024*1024), "MB) \n")
	device := devices[deviceIndex]
	ClDevices = devices
	ClDevice = device
	context, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		fmt.Printf("CreateContext failed: %+v \n", err)
	}
	queue, err := context.CreateCommandQueue(device, 0)
	if err != nil {
		fmt.Printf("CreateCommandQueue failed: %+v \n", err)
	}
	program, err := context.CreateProgramWithSource([]string{kernels.GenMergedKernelSource()})
	if err != nil {
		fmt.Printf("CreateProgramWithSource failed: %+v \n", err)
	}
	if err := program.BuildProgram(nil, "-cl-std=CL1.2 -cl-kernel-arg-info"); err != nil {
		fmt.Printf("BuildProgram failed: %+v \n", err)
	}

	KernList = map[string]*cl.Kernel{}
	for i0 := range kernels.KernelsList {
		kernName := kernels.KernelsList[i0]
		KernList[kernName], err = program.CreateKernel(kernName)
		if err != nil {
		       fmt.Printf("CreateKernel failed: %+v \n", err)
		}
	}
	ClCtx = context
	ClCmdQueue = queue
	ClProgram = program

}

func ReleaseAndClean() {
	ClCmdQueue.Release()
	ClProgram.Release()
	ClCtx.Release()
}

// Global stream used for everything
//const stream0 = cu.Stream(0)

// Synchronize the global stream
// This is called before and after all memcopy operations between host and device.
//func Sync() {
//	stream0.Synchronize()
//}
