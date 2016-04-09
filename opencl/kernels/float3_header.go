package kernels

// This file implements common functions on float3 (vector).
// Author: Mykola Dvornik, Arne Vansteenkiste

var Float3Header = `
#ifndef _FLOAT3_H_
#define _FLOAT3_H_

// converting set of 3 floats into a 3-component vector
inline float3 make_float3(float a, float b, float c) {
	return (float3) {a, b, c};
}

// length of the 3-components vector
inline float len(float3 a) {
	return length(a);
}

// returns a normalized copy of the 3-components vector
inline float3 normalized(float3 a){
	return normalize(a);
}

// square
inline float pow2(float x){
	return x * x;
}


// pow(x, 3)
inline float pow3(float x){
	return x * x * x;
}


// pow(x, 4)
inline float pow4(float x){
	float s = x*x;
	return s*s;
}

#define is0(m) ( dot(m, m) == 0.0f )

#endif

`
