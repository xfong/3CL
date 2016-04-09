package kernels

var CLRNG_MRG31k3p_Header = `
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

/* mrg31k3p.clh
 *  Specific device interface for the MRG31k3p generator
 *
 *  The functions defined in this file are not documented here.  Refer to the
 *  documentation of mrg31k3p.h.
 */

#ifndef CLRNG_SINGLE_PRECISION
#  define CLRNG_SINGLE_PRECISION
#endif

#ifndef CLRNG_ENABLE_SUBSTREAMS
#  define CLRNG_ENABLE_SUBSTREAMS
#endif

#ifndef MRG31K3P_CLH
#  define MRG31K3P_CLH

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
} clrngMrg31k3pStreamState;

struct clrngMrg31k3pStream_ {
    clrngMrg31k3pStreamState current;
    __global const clrngMrg31k3pStreamState* initial;
    clrngMrg31k3pStreamState substream;
};
typedef struct clrngMrg31k3pStream_ clrngMrg31k3pStream;

struct clrngMrg31k3pHostStream_ {
    clrngMrg31k3pStreamState current;
    clrngMrg31k3pStreamState initial;
    clrngMrg31k3pStreamState substream;
};
typedef struct clrngMrg31k3pHostStream_ clrngMrg31k3pHostStream;

clrngStatus clrngMrg31k3pCopyOverStreamsFromGlobal(size_t count, clrngMrg31k3pStream* destStreams, __global const clrngMrg31k3pHostStream* srcStreams)
{
    //Check params
    if (!destStreams)
	return CLRNG_NULL_DEST;
    if (!srcStreams)
	return CLRNG_NULL_SRC;

    for (size_t i = 0; i < count; i++) {
	destStreams[i].current   = srcStreams[i].current;
	destStreams[i].initial   = &srcStreams[i].initial;
	destStreams[i].substream = srcStreams[i].substream;
    }

    return CLRNG_SUCCESS;
}

clrngStatus clrngMrg31k3pCopyOverStreamsToGlobal(size_t count, __global clrngMrg31k3pHostStream* destStreams, const clrngMrg31k3pStream* srcStreams)
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
//  Also works if v = s.
//  Returns v = A*s % m
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
// Begin ./private/mrg31k3p.c.h
#  ifndef PRIVATE_MRG31K3P_CH
#    define PRIVATE_MRG31K3P_CH

#    define mrg31k3p_M1 2147483647             /* 2^31 - 1 */
#    define mrg31k3p_M2 2147462579             /* 2^31 - 21069 */

#    define mrg31k3p_MASK12 511                /* 2^9 - 1 */
#    define mrg31k3p_MASK13 16777215           /* 2^24 - 1 */
#    define mrg31k3p_MASK2 65535               /* 2^16 - 1 */
#    define mrg31k3p_MULT2 21069

#    define mrg31k3p_NORM_cl_float  4.6566126e-10



#    if defined(CLRNG_ENABLE_SUBSTREAMS) || !defined(__CLRNG_DEVICE_API)

// clrngMrg31k3p_A1p72 and clrngMrg31k3p_A2p72 jump 2^72 steps forward
#      if defined(__CLRNG_DEVICE_API)
__constant
#      else
static
#      endif
cl_uint clrngMrg31k3p_A1p72[3][3] = { 
    {1516919229,  758510237, 499121365},
    {1884998244, 1516919229, 335398200},
    {601897748,  1884998244, 358115744}
};

#      if defined(__CLRNG_DEVICE_API)
__constant
#      else
static
#      endif
cl_uint clrngMrg31k3p_A2p72[3][3] = { 
    {1228857673, 1496414766,  954677935},
    {1133297478, 1407477216, 1496414766},
    {2002613992, 1639496704, 1407477216}
};

#    endif


clrngStatus clrngMrg31k3pCopyOverStreams(size_t count, clrngMrg31k3pStream* destStreams, const clrngMrg31k3pStream* srcStreams)
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

/* Advance the rng one step and returns z such that 1 <= z <= mrg31k3p_M1
 */
static cl_uint clrngMrg31k3pNextState(clrngMrg31k3pStreamState* currentState)
{
	
	cl_uint* g1 = currentState->g1;
	cl_uint* g2 = currentState->g2;
	cl_uint y1, y2;

	// first component
	y1 = ((g1[1] & mrg31k3p_MASK12) << 22) + (g1[1] >> 9)
		+ ((g1[2] & mrg31k3p_MASK13) << 7) + (g1[2] >> 24);

	if (y1 >= mrg31k3p_M1)
		y1 -= mrg31k3p_M1;

	y1 += g1[2];
	if (y1 >= mrg31k3p_M1)
		y1 -= mrg31k3p_M1;

	g1[2] = g1[1];
	g1[1] = g1[0];
	g1[0] = y1;

	// second component
	y1 = ((g2[0] & mrg31k3p_MASK2) << 15) + (mrg31k3p_MULT2 * (g2[0] >> 16));
	if (y1 >= mrg31k3p_M2)
		y1 -= mrg31k3p_M2;
	y2 = ((g2[2] & mrg31k3p_MASK2) << 15) + (mrg31k3p_MULT2 * (g2[2] >> 16));
	if (y2 >= mrg31k3p_M2)
		y2 -= mrg31k3p_M2;
	y2 += g2[2];
	if (y2 >= mrg31k3p_M2)
		y2 -= mrg31k3p_M2;
	y2 += y1;
	if (y2 >= mrg31k3p_M2)
		y2 -= mrg31k3p_M2;

	g2[2] = g2[1];
	g2[1] = g2[0];
	g2[0] = y2;

	if (g1[0] <= g2[0])
		return (g1[0] - g2[0] + mrg31k3p_M1);
	else
		return (g1[0] - g2[0]);
}

cl_float clrngMrg31k3pRandomU01_cl_float(clrngMrg31k3pStream* stream) {
    return clrngMrg31k3pNextState(&stream->current) * mrg31k3p_NORM_cl_float;
}

cl_int clrngMrg31k3pRandomInteger_cl_float(clrngMrg31k3pStream* stream, cl_int i, cl_int j) {
    return i + (cl_int)((j - i + 1) * clrngMrg31k3pRandomU01_cl_float(stream));
}

clrngStatus clrngMrg31k3pRandomU01Array_cl_float(clrngMrg31k3pStream* stream, size_t count, cl_float* buffer) {
	if (!stream)
		return CLRNG_NULL_STREAM;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t i = 0; i < count; i++)
		buffer[i] = clrngMrg31k3pRandomU01_cl_float(stream);
	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg31k3pRandomIntegerArray_cl_float(clrngMrg31k3pStream* stream, cl_int i, cl_int j, size_t count, cl_int* buffer) {
	if (!stream)
		return CLRNG_NULL_STREAM;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t k = 0; k < count; k++)
		buffer[k] = clrngMrg31k3pRandomInteger_cl_float(stream, i, j);
	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg31k3pRewindStreams(size_t count, clrngMrg31k3pStream* streams)
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

#    if defined(CLRNG_ENABLE_SUBSTREAMS) || !defined(__CLRNG_DEVICE_API)
clrngStatus clrngMrg31k3pRewindSubstreams(size_t count, clrngMrg31k3pStream* streams)
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

clrngStatus clrngMrg31k3pForwardToNextSubstreams(size_t count, clrngMrg31k3pStream* streams)
{
	//Check params
	if (!streams)
		return CLRNG_NULL_STREAM;
	
	for (size_t k = 0; k < count; k++) {
		modMatVec (clrngMrg31k3p_A1p72, streams[k].substream.g1, streams[k].substream.g1, mrg31k3p_M1);
		modMatVec (clrngMrg31k3p_A2p72, streams[k].substream.g2, streams[k].substream.g2, mrg31k3p_M2);
		streams[k].current = streams[k].substream;
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngMrg31k3pMakeOverSubstreams(clrngMrg31k3pStream* stream, size_t count, clrngMrg31k3pStream* substreams)
{
	for (size_t i = 0; i < count; i++) {
		clrngStatus err;
		// snapshot current stream into substreams[i]
		err = clrngMrg31k3pCopyOverStreams(1, &substreams[i], stream);
		if (err != CLRNG_SUCCESS)
		    return err;
		// advance to next substream
		err = clrngMrg31k3pForwardToNextSubstreams(1, stream);
		if (err != CLRNG_SUCCESS)
		    return err;
	}
	return CLRNG_SUCCESS;
}
#    endif // substreams

#  endif // PRIVATE_MRG31K3P_CH
// End ./private/mrg31k3p.c.h

#endif

`
