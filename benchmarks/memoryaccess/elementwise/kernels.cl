// elementwise
__kernel void elementwise(__global float* A, 
                          __global float* B, 
                          __global float* C) {
  // Get thread's idx
  int index = get_global_id(0);

  // Perform addition using global memory
  C[index] = A[index] * B[index];
}