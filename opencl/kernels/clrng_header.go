package kernels

var CLRNG_Header = `
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

/* lfsr113.clh, mrg31k3p.clh, mrg32k3a.clh
*   Specific device interface for the lfsr113, mrg31k3p and mrg32k3a generators
*
*  The functions defined in this file are not documented here.  Refer to the
*  documentation of lfsr113.h, mrg31k3p.h, and mrg32k3a.h.
*/

#ifndef CLRNG_SINGLE_PRECISION
#  define CLRNG_SINGLE_PRECISION
#endif

#ifndef CLRNG_ENABLE_SUBSTREAMS
#  define CLRNG_ENABLE_SUBSTREAMS
#endif

#ifndef CLRNG_CLH
#  define CLRNG_CLH

#  ifndef __OPENCL_C_VERSION__
#    error "clRNG.clh can be included in device code only"
#  endif

#  define __CLRNG_DEVICE_API

#  define _CLRNG_FPTYPE cl_float

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

#endif

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

#ifndef PRIVATE_MODULAR_CH

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

#endif

#ifndef LFSR113_CLH
#  define LFSR113_CLH

/********************************************************************************
 * Functions and types declarations                                             *
 ********************************************************************************/

typedef struct {
	/* Seed for the lfsr113 generator
	*/
	cl_uint g[4];

} clrngLfsr113StreamState;

struct clrngLfsr113Stream_ {
	clrngLfsr113StreamState current;
	__global const clrngLfsr113StreamState* initial;
	clrngLfsr113StreamState substream;
};
typedef struct clrngLfsr113Stream_ clrngLfsr113Stream;

struct clrngLfsr113HostStream_ {
	clrngLfsr113StreamState current;
	clrngLfsr113StreamState initial;
	clrngLfsr113StreamState substream;
};
typedef struct clrngLfsr113HostStream_ clrngLfsr113HostStream;

/********************************************************************************
 * Implementation                                                               *
 ********************************************************************************/

clrngStatus clrngLfsr113CopyOverStreamsFromGlobal(size_t count, clrngLfsr113Stream* destStreams, __global const clrngLfsr113HostStream* srcStreams)
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

clrngStatus clrngLfsr113CopyOverStreamsToGlobal(size_t count, __global clrngLfsr113HostStream* destStreams, const clrngLfsr113Stream* srcStreams)
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

// code that is common to host and device
#  ifndef PRIVATE_LFSR113_CH
#    define PRIVATE_LFSR113_CH

#    define Lfsr113_NORM_cl_float  2.3283063e-10


clrngStatus clrngLfsr113CopyOverStreams(size_t count, clrngLfsr113Stream* destStreams, const clrngLfsr113Stream* srcStreams)
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

/* Advance the rng one step and returns z such that 1 <= z <= lfsr113_M1
*/
static cl_uint clrngLfsr113NextState(clrngLfsr113StreamState *currentState) {

	cl_uint b;

	b = (((currentState->g[0] << 6) ^ currentState->g[0]) >> 13);
	currentState->g[0] = (((currentState->g[0] & 4294967294U) << 18) ^ b);

	b = (((currentState->g[1] << 2) ^ currentState->g[1]) >> 27);
	currentState->g[1] = (((currentState->g[1] & 4294967288U) << 2) ^ b);

	b = (((currentState->g[2] << 13) ^ currentState->g[2]) >> 21);
	currentState->g[2] = (((currentState->g[2] & 4294967280U) << 7) ^ b);

	b = (((currentState->g[3] << 3) ^ currentState->g[3]) >> 12);
	currentState->g[3] = (((currentState->g[3] & 4294967168U) << 13) ^ b);

	return (currentState->g[0] ^ currentState->g[1] ^ currentState->g[2] ^ currentState->g[3]);

}

cl_float clrngLfsr113RandomU01_cl_float(clrngLfsr113Stream* stream) {
    return clrngLfsr113NextState(&stream->current) * Lfsr113_NORM_cl_float;
}

cl_int clrngLfsr113RandomInteger_cl_float(clrngLfsr113Stream* stream, cl_int i, cl_int j) {
    return i + (cl_int)((j - i + 1) * clrngLfsr113RandomU01_cl_float(stream));
}

clrngStatus clrngLfsr113RandomU01Array_cl_float(clrngLfsr113Stream* stream, size_t count, cl_float* buffer) {
	if (!stream)
		return CLRNG_NULL_STREAM;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t i = 0; i < count; i++)
		buffer[i] = clrngLfsr113RandomU01_cl_float(stream);
	return CLRNG_SUCCESS;
}

clrngStatus clrngLfsr113RandomIntegerArray_cl_float(clrngLfsr113Stream* stream, cl_int i, cl_int j, size_t count, cl_int* buffer) {
	if (!stream)
		return CLRNG_NULL_STREAM;
	if (!buffer)
		return CLRNG_NULL_BUFFER;
	for (size_t k = 0; k < count; k++)
		buffer[k] = clrngLfsr113RandomInteger_cl_float(stream, i, j);
	return CLRNG_SUCCESS;
}

clrngStatus clrngLfsr113RewindStreams(size_t count, clrngLfsr113Stream* streams)
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
clrngStatus clrngLfsr113RewindSubstreams(size_t count, clrngLfsr113Stream* streams)
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

void lfsr113ResetNextSubStream(clrngLfsr113Stream* stream){

	/* The following operations make the jump ahead with
	2 ^ 55 iterations for every component of the generator.
	The internal state after the jump, however, is slightly different
	from 2 ^ 55 iterations since it ignores the state in
	which are found the first bits of each components,
	since they are ignored in the recurrence.The state becomes
	identical to what one would with normal iterations
	after a call nextValue().*/

	int z, b;

	cl_uint* subStreamState = stream->substream.g;

	//Calculate the first component
	z = subStreamState[0] & (cl_uint)-2;
	b = (z << 6) ^ z;

	z = (z) ^ (z << 3) ^ (z << 4) ^ (z << 6) ^ (z << 7) ^
		(z << 8) ^ (z << 10) ^ (z << 11) ^ (z << 13) ^ (z << 14) ^
		(z << 16) ^ (z << 17) ^ (z << 18) ^ (z << 22) ^
		(z << 24) ^ (z << 25) ^ (z << 26) ^ (z << 28) ^ (z << 30);

	z ^= ((b >> 1) & 0x7FFFFFFF) ^
		((b >> 3) & 0x1FFFFFFF) ^
		((b >> 5) & 0x07FFFFFF) ^
		((b >> 6) & 0x03FFFFFF) ^
		((b >> 7) & 0x01FFFFFF) ^
		((b >> 9) & 0x007FFFFF) ^
		((b >> 13) & 0x0007FFFF) ^
		((b >> 14) & 0x0003FFFF) ^
		((b >> 15) & 0x0001FFFF) ^
		((b >> 17) & 0x00007FFF) ^
		((b >> 18) & 0x00003FFF) ^
		((b >> 20) & 0x00000FFF) ^
		((b >> 21) & 0x000007FF) ^
		((b >> 23) & 0x000001FF) ^
		((b >> 24) & 0x000000FF) ^
		((b >> 25) & 0x0000007F) ^
		((b >> 26) & 0x0000003F) ^
		((b >> 27) & 0x0000001F) ^
		((b >> 30) & 0x00000003);
	subStreamState[0] = z;

	//Calculate the second component
	z = subStreamState[1] & (cl_uint)-8;
	b = z ^ (z << 1);
	b ^= (b << 2);
	b ^= (b << 4);
	b ^= (b << 8);

	b <<= 8;
	b ^= (z << 22) ^ (z << 25) ^ (z << 27);
	if ((z & 0x80000000) != 0) b ^= 0xABFFF000;
	if ((z & 0x40000000) != 0) b ^= 0x55FFF800;

	z = b ^ ((z >> 7) & 0x01FFFFFF) ^
		((z >> 20) & 0x00000FFF) ^
		((z >> 21) & 0x000007FF);

	subStreamState[1] = z;

	//Calculate the third component
	z = subStreamState[2] & (cl_uint)-16;
	b = (z << 13) ^ z;
	z = ((b >> 3) & 0x1FFFFFFF) ^
		((b >> 17) & 0x00007FFF) ^
		(z << 10) ^ (z << 11) ^ (z << 25);
	subStreamState[2] = z;

	//Calculate the forth component
	z = subStreamState[3] & (cl_uint)-128;
	b = (z << 3) ^ z;
	z = (z << 14) ^ (z << 16) ^ (z << 20) ^
		((b >> 5) & 0x07FFFFFF) ^
		((b >> 9) & 0x007FFFFF) ^
		((b >> 11) & 0x001FFFFF);
	subStreamState[3] = z;

	clrngLfsr113RewindSubstreams(1, stream);
}
clrngStatus clrngLfsr113ForwardToNextSubstreams(size_t count, clrngLfsr113Stream* streams)
{
	//Check params
	if (!streams)
		return CLRNG_NULL_STREAM;

	for (size_t k = 0; k < count; k++) {

		lfsr113ResetNextSubStream(&streams[k]);
	}

	return CLRNG_SUCCESS;
}

clrngStatus clrngLfsr113MakeOverSubstreams(clrngLfsr113Stream* stream, size_t count, clrngLfsr113Stream* substreams)
{
	for (size_t i = 0; i < count; i++) {
		clrngStatus err;
		// snapshot current stream into substreams[i]
		err = clrngLfsr113CopyOverStreams(1, &substreams[i], stream);
		if (err != CLRNG_SUCCESS)
		    return err;
		// advance to next substream
		err = clrngLfsr113ForwardToNextSubstreams(1, stream);
		if (err != CLRNG_SUCCESS)
		    return err;
	}
	return CLRNG_SUCCESS;
}

#    endif

#  endif // PRIVATE_Lfsr113_CH

#endif

#ifndef MRG31K3P_CLH
#  define MRG31K3P_CLH

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

#ifndef MRG32K3A_CLH
#  define MRG32K3A_CLH

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
