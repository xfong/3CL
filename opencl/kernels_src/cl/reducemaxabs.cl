__kernel void
reducemaxabs(__global float* __restrict src, __global float* __restrict dst, float initVal, int n, __local float* scratch) {
	// Calculate indices
	int local_idx = get_local_id(0); // Work-item index within workgroup
	int grp_sz = get_local_size(0); // Total number of work-items in each workgroup
	int grp_id = get_group_id(0); // Index of workgroup
	int global_idx = grp_id * grp_sz + local_idx; // Calculate global index of work-item
	int grp_offset = get_num_groups(0) * grp_sz; // Offset for memory access

	// Initialize memory
	float currVal = initVal;
	float other;
	float mine;

	// Loop over input elements in chunks and store max in each chunk into local memory
	while (global_idx < n) {
		float element = fabs(src[global_idx]);
		currVal = fmax(currVal, element);
		global_idx += grp_offset;
	}

	// At this point, max values on chunks are in local memory. Perform parallel reduction
	scratch[local_idx] = currVal;
	// Add barrier to sync all threads
	barrier(CLK_LOCAL_MEM_FENCE);
	for (int offset = grp_sz / 2; offset > 32; offset >>= 1) {
		if (local_idx < offset) {
			other = scratch[local_idx + offset];
			mine = scratch[local_idx];
			scratch[local_idx] = fmax(mine, other);
		}
		// barrier for syncing work group
		barrier(CLK_LOCAL_MEM_FENCE);
	}

	if (local_idx < 32) {
		other = scratch[local_idx + 32];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
		other = scratch[local_idx + 16];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
		other = scratch[local_idx + 8];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
		other = scratch[local_idx + 4];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
		other = scratch[local_idx + 2];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
		other = scratch[local_idx + 1];
		mine = scratch[local_idx];
		scratch[local_idx] = fmax(mine, other);
	}
	
	if (local_idx == 0) {
		dst[get_group_id(0)] = scratch[0];
	}
}
