package kernels

var AtomicFHeader = `
#ifndef _ATOMICF_H_
#define _ATOMICF_H_

// Atomic max of abs value.
inline void atomicFmaxabs(volatile __global float* a, float b){
	b = fabs(b);
	atomic_max((__global int*)(a), *((int*)(&b)));
}

#endif

`
