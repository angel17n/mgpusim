__kernel void explicit_read(__global const float* in) {

  float a;
  __global const float *p = in;
  
  __asm__ volatile("flat_load_dword %0, %1" : "=v"(a) : "v"(p));
}
