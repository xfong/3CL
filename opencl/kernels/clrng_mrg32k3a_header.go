package kernels

var CLRNG_MRG32k3a_Header = `
 /*
 ***********************************************************************
 Copyright (c) 2015 Advanced Micro Devices, Inc. 
 All rights reserved.
 
 Redistribution and use in source and binary forms, with or without 
 modification, are permitted provided that the following conditions 
 are met:
 
 1. Redistributions of source code must retain the above copyright 
 notice, this list of conditions and the following disclaimer.
 
 2. Redistributions in binary form must reproduce the above copyright 
 notice, this list of conditions and the following disclaimer in the 
 documentation and/or other materials provided with the distribution.
 
 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS 
 "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT 
 LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR 
 A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT 
 HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, 
 SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT 
 LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, 
 DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY 
 THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT 
 (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE 
 OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 
 ***********************************************************************
 */

/* Mrg32k3a.clh
*  Specific device interface for the Mrg32k3a generator
*
*  The functions defined in this file are not documented here.  Refer to the
*  documentation of Mrg32k3a.h.
*/

#ifndef CLRNG_SINGLE_PRECISION
#  define CLRNG_SINGLE_PRECISION
#endif

#ifndef CLRNG_ENABLE_SUBSTREAMS
#  define CLRNG_ENABLE_SUBSTREAMS
#endif

#ifndef MRG32K3A_CLH
#  define MRG32K3A_CLH

#  ifndef CLRNG_CLH
#    define CLRNG_CLH

#    ifndef __OPENCL_C_VERSION__
#      error "clRNG.clh can be included in device code only"
#    endif

#    define __CLRNG_DEVICE_API

#    define _CLRNG_FPTYPE cl_float

typedef float  cl_float;
typedef int    cl_int;
typedef uint   cl_uint;


typedef enum clrngStatus_ {
    CLRNG_SUCCESS              = 0,
    CLRNG_INVALID_VALUE        = -1,
    CLRNG_NULL_DEST            = -2,
    CLRNG_NULL_SRC             = -3,
    CLRNG_NULL_BUFFER          = -4,
    CLRNG_NULL_STREAM          = -5
} clrngStatus;

#  endif

/********************************************************************************
 * Functions and types declarations                                             *
 ********************************************************************************/

typedef struct {
	/* Seed for the first MRG component
	*/
	cl_uint g1[3];
	/* Seed for the second MRG component
	*/
	cl_uint g2[3];
} clrngMrg32k3aStreamState;

struct clrngMrg32k3aStream_ {
	clrngMrg32k3aStreamState current;
	__global const clrngMrg32k3aStreamState* initial;
	clrngMrg32k3aStreamState substream;
};
typedef struct clrngMrg32k3aStream_ clrngMrg32k3aStream;

struct clrngMrg32k3aHostStream_ {
	clrngMrg32k3aStreamState current;
	clrngMrg32k3aStreamState initial;
	clrngMrg32k3aStreamState substream;
};
typedef struct clrngMrg32k3aHostStream_ clrngMrg32k3aHostStream;

clrngStatus clrngMrg32k3aCopyOverStreamsFromGlobal(size_t count, clrngMrg32k3aStream* destStreams, __global const clrngMrg32k3aHostStream* srcStreams)
{
	//Check params
	if (!destStreams)
		return CLRNG_NULL_DEST;
	if (!srcStreams)
		return CLRNG_NULL_SRC;

	for (size_t i = 0; i < count; i++) {
		destStreams[i].current = srcStreams[i].current;
		destStreams[i].initial = &srcStreams[i].initial;
		destStreams[i].substream = srcStreams[i].substream;
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg32k3aCopyOverStreamsToGlobal(size_t count, __global clrngMrg32k3aHostStream* destStreams, const clrngMrg32k3aStream* srcStreams)
{
    //Check params
    if (!destStreams)
	return CLRNG_NULL_DEST;
    if (!srcStreams)
	return CLRNG_NULL_SRC;

    for (size_t i = 0; i < count; i++) {
	destStreams[i].current   = srcStreams[i].current;
	destStreams[i].initial   = *srcStreams[i].initial;
	destStreams[i].substream = srcStreams[i].substream;
    }

    return CLRNG_SUCCESS;
}

   // Begin ./private/modular.c.h
/* Modular arithmetic and linear algebra
 *
 *  This file provides the code common to the host and device.
 *
 *  The preprocessor symbol "MODULAR_NUMBER_TYPE" must be defined as the type
 *  of number (cl_uint, cl_ulong, etc.) on which the modular functions operate.
 *
 *  To use the fixed size variant, the preprocessor constant
 *  "MODULAR_FIXED_SIZE" must be set to the size (number of rows or of columns)
 *  of the matrix.
 *
 *  If the project is migrated to C++, this could be rewritten much more
 *  clearly using templates.
 */

#  ifndef PRIVATE_MODULAR_CH

//! Compute (a*s + c) % m
static cl_uint modMult(cl_uint a, cl_uint s, cl_uint c, cl_uint m)
{
    cl_uint v;
    v = (cl_uint) (((cl_uint) a * s + c) % m);
    return v;
}


// Matrix-vector modular multiplication
//   Also works if v = s.
//   Returns v = A*s % m
static void modMatVec (__constant cl_uint A[3][3], cl_uint s[3], cl_uint v[3], cl_uint m)
{
    cl_uint x[3];     // Necessary if v = s
    for (size_t i = 0; i < 3; ++i) {
        x[i] = 0;
        for (size_t j = 0; j < 3; j++)
            x[i] = modMult(A[i][j], s[j], x[i], m);
    }
    for (size_t i = 0; i < 3; ++i)
        v[i] = x[i];
}

#  endif
   // End ./private/modular.c.h

// code that is common to host and device
// Begin ./private/mrg32k3a.c.h
#  ifndef PRIVATE_MRG32K3A_CH
#    define PRIVATE_MRG32K3A_CH

#    define Mrg32k3a_M1 4294967087            
#    define Mrg32k3a_M2 4294944443             

#    define Mrg32k3a_NORM_cl_float  2.3283064e-10

#    if defined(CLRNG_ENABLE_SUBSTREAMS) || !defined(__CLRNG_DEVICE_API)

// clrngMrg32k3a_A1p76 and clrngMrg32k3a_A2p76 jump 2^76 steps forward
#      if defined(__CLRNG_DEVICE_API)
__constant
#      else
static
#      endif
cl_uint clrngMrg32k3a_A1p76[3][3] = {
	{ 82758667, 1871391091, 4127413238 },
	{ 3672831523, 69195019, 1871391091 },
	{ 3672091415, 3528743235, 69195019 }
};

#      if defined(__CLRNG_DEVICE_API)
__constant
#      else
static
#      endif
cl_uint clrngMrg32k3a_A2p76[3][3] = {
	{ 1511326704, 3759209742, 1610795712 },
	{ 4292754251, 1511326704, 3889917532 },
	{ 3859662829, 4292754251, 3708466080 }
};

#    endif


clrngStatus clrngMrg32k3aCopyOverStreams(size_t count, clrngMrg32k3aStream* destStreams, const clrngMrg32k3aStream* srcStreams)
{
	//Check params
	if (!destStreams)
		return CLRNG_NULL_DEST;
	if (!srcStreams)
		return CLRNG_NULL_SRC;

	for (size_t i = 0; i < count; i++)
		destStreams[i] = srcStreams[i];

	return CLRNG_SUCCESS;
}

/* Advance the rng one step and returns z such that 1 <= z <= Mrg32k3a_M1
*/
static cl_uint clrngMrg32k3aNextState(clrngMrg32k3aStreamState* currentState)
{

	cl_uint* g1 = currentState->g1;
	cl_uint* g2 = currentState->g2;

	cl_int p0, p1;

	/* component 1 */
	p0 = 1403580 * g1[1] - 810728 * g1[0];
	p0 %= Mrg32k3a_M1;
	if (p0 < 0)
		p0 += Mrg32k3a_M1;
	g1[0] = g1[1];
	g1[1] = g1[2];
	g1[2] = p0;

	/* component 2 */
	p1 = 527612 * g2[2] - 1370589 * g2[0];
	p1 %= Mrg32k3a_M2;
	if (p1 < 0)
		p1 += Mrg32k3a_M2;
	g2[0] = g2[1];
	g2[1] = g2[2];
	g2[2] = p1;

	/* combinations */
	if (p0 > p1)
		return (p0 - p1);
	else return (p0 - p1 + Mrg32k3a_M1);
}


cl_float clrngMrg32k3aRandomU01_cl_float(clrngMrg32k3aStream* stream) {
    return clrngMrg32k3aNextState(&stream->current) * Mrg32k3a_NORM_cl_float;
}

cl_int clrngMrg32k3aRandomInteger_cl_float(clrngMrg32k3aStream* stream, cl_int i, cl_int j) {
    return i + (cl_int)((j - i + 1) * clrngMrg32k3aRandomU01_cl_float(stream));
}

clrngStatus clrngMrg32k3aRandomU01Array_cl_float(clrngMrg32k3aStream* stream, size_t count, cl_float* buffer) {
	if (!stream)
		return CLRNG_NULL_SRC;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t i = 0; i < count; i++)
		buffer[i] = clrngMrg32k3aRandomU01_cl_float(stream);
	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg32k3aRandomIntegerArray_cl_float(clrngMrg32k3aStream* stream, cl_int i, cl_int j, size_t count, cl_int* buffer) {
	if (!stream)
		return CLRNG_NULL_SRC;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t k = 0; k < count; k++)
		buffer[k] = clrngMrg32k3aRandomInteger_cl_float(stream, i, j);
	return CLRNG_SUCCESS;
}


clrngStatus clrngMrg32k3aRewindStreams(size_t count, clrngMrg32k3aStream* streams)
{
	//Check params
	if (!streams)
		return CLRNG_NULL_STREAM;
	//Reset current state to the stream initial state
	for (size_t j = 0; j < count; j++) {
#    ifdef __CLRNG_DEVICE_API
		streams[j].current = streams[j].substream = *streams[j].initial;
#    else
		streams[j].current = streams[j].substream = streams[j].initial;
#    endif
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg32k3aRewindSubstreams(size_t count, clrngMrg32k3aStream* streams)
{
	//Check params
	if (!streams)
		return CLRNG_NULL_STREAM;
	//Reset current state to the subStream initial state
	for (size_t j = 0; j < count; j++) {
		streams[j].current = streams[j].substream;
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg32k3aForwardToNextSubstreams(size_t count, clrngMrg32k3aStream* streams)
{
	//Check params
	if (!streams)
		return CLRNG_NULL_STREAM;

	for (size_t k = 0; k < count; k++) {
		modMatVec(clrngMrg32k3a_A1p76, streams[k].substream.g1, streams[k].substream.g1, Mrg32k3a_M1);
		modMatVec(clrngMrg32k3a_A2p76, streams[k].substream.g2, streams[k].substream.g2, Mrg32k3a_M2);
		streams[k].current = streams[k].substream;
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg32k3aMakeOverSubstreams(clrngMrg32k3aStream* stream, size_t count, clrngMrg32k3aStream* substreams)
{
	for (size_t i = 0; i < count; i++) {
		clrngStatus err;
		// snapshot current stream into substreams[i]
		err = clrngMrg32k3aCopyOverStreams(1, &substreams[i], stream);
		if (err != CLRNG_SUCCESS)
		    return err;
		// advance to next substream
		err = clrngMrg32k3aForwardToNextSubstreams(1, stream);
		if (err != CLRNG_SUCCESS)
		    return err;
	}
	return CLRNG_SUCCESS;
}

#  endif // PRIVATE_Mrg32k3a_CH
// End ./private/mrg32k3a.c.h

#endif

`
