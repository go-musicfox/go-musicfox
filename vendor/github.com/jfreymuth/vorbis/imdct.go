package vorbis

import (
	"math"
	"math/bits"
)

type imdctLookup struct {
	A, B, C []float32
}

func generateIMDCTLookup(n int, l *imdctLookup) {
	l.A = make([]float32, n/2)
	l.B = make([]float32, n/2)
	l.C = make([]float32, n/4)
	fn := float64(n)
	for k := 0; k < n/4; k++ {
		fk := float64(k)
		l.A[2*k] = float32(math.Cos(4 * fk * math.Pi / fn))
		l.A[2*k+1] = float32(-math.Sin(4 * fk * math.Pi / fn))
		l.B[2*k] = float32(math.Cos((2*fk + 1) * math.Pi / fn / 2))
		l.B[2*k+1] = float32(math.Sin((2*fk + 1) * math.Pi / fn / 2))
	}
	for k := 0; k < n/8; k++ {
		fk := float64(k)
		l.C[2*k] = float32(math.Cos(2 * (2*fk + 1) * math.Pi / fn))
		l.C[2*k+1] = float32(-math.Sin(2 * (2*fk + 1) * math.Pi / fn))
	}
}

// "inverse modified discrete cosine transform"
func imdct(t *imdctLookup, in, out []float32) {
	n := len(in) * 2

	n2, n4, n8 := n/2, n/4, n/8
	n3_4 := n - n4

	// more of these steps could be done in place, but we need two arrays anyway
	for j := 0; j < n8; j++ {
		a0 := t.A[n2-2*j-1]
		a1 := t.A[n2-2*j-2]
		a2 := t.A[n4-2*j-1]
		a3 := t.A[n4-2*j-2]
		a4 := t.A[n2-4-4*j]
		a5 := t.A[n2-3-4*j]
		v0 := (-in[4*j+3])*a0 + (-in[4*j+1])*a1
		v1 := (-in[4*j+3])*a1 - (-in[4*j+1])*a0
		v2 := (in[n2-4*j-4])*a2 + (in[n2-4*j-2])*a3
		v3 := (in[n2-4*j-4])*a3 - (in[n2-4*j-2])*a2
		out[n4+2*j+1] = v3 + v1
		out[n4+2*j] = v2 + v0
		out[2*j+1] = (v3-v1)*a4 - (v2-v0)*a5
		out[2*j] = (v2-v0)*a4 + (v3-v1)*a5
	}
	ld := int(ilog(n) - 1)
	for l := 0; l < ld-3; l++ {
		k0 := n >> uint(l+3)
		k1 := 1 << uint(l+3)
		rlim := n >> uint(l+4)
		s2lim := 1 << uint(l+2)
		for r := 0; r < rlim; r++ {
			a0 := t.A[r*k1]
			a1 := t.A[r*k1+1]
			i0 := n2 - 1 - 2*r
			i1 := n2 - 2 - 2*r
			i2 := n2 - 1 - k0 - 2*r
			i3 := n2 - 2 - k0 - 2*r
			for s2 := 0; s2 < s2lim; s2 += 2 {
				v0, v1 := out[i0], out[i1]
				v2, v3 := out[i2], out[i3]
				out[i0] = v0 + v2
				out[i1] = v1 + v3
				out[i2] = (v0-v2)*a0 - (v1-v3)*a1
				out[i3] = (v1-v3)*a0 + (v0-v2)*a1
				i0 -= 2 * k0
				i1 -= 2 * k0
				i2 -= 2 * k0
				i3 -= 2 * k0
			}
		}
	}
	for i := 0; i < n8; i++ {
		j := int(bits.Reverse32(uint32(i)) >> uint(32-ld+3))
		if i < j {
			out[4*j], out[4*i] = out[4*i], out[4*j]
			out[4*j+1], out[4*i+1] = out[4*i+1], out[4*j+1]
			out[4*j+2], out[4*i+2] = out[4*i+2], out[4*j+2]
			out[4*j+3], out[4*i+3] = out[4*i+3], out[4*j+3]
		}
	}
	for k := 0; k < n8; k++ {
		in[n2-1-2*k] = out[4*k]
		in[n2-2-2*k] = out[4*k+1]
		in[n4-1-2*k] = out[4*k+2]
		in[n4-2-2*k] = out[4*k+3]
	}
	i0 := 0
	i1 := 1
	i2 := n2 - 2
	i3 := n2 - 1
	for k := 0; k < n8; k++ {
		v0, v1 := in[i0], in[i1]
		v2, v3 := in[i2], in[i3]
		c0 := t.C[i0]
		c1 := t.C[i1]
		out[i0] = (v0 + v2 + c1*(v0-v2) + c0*(v1+v3)) / 2
		out[i2] = (v0 + v2 - c1*(v0-v2) - c0*(v1+v3)) / 2
		out[i1] = (v1 - v3 + c1*(v1+v3) - c0*(v0-v2)) / 2
		out[i3] = (-v1 + v3 + c1*(v1+v3) - c0*(v0-v2)) / 2
		i0 += 2
		i1 += 2
		i2 -= 2
		i3 -= 2
	}
	for k := 0; k < n4; k++ {
		b0 := t.B[2*k]
		b1 := t.B[2*k+1]
		v0 := out[2*k]
		v1 := out[2*k+1]
		in[k] = v0*b0 + v1*b1
		in[n2-1-k] = v0*b1 - v1*b0
	}
	for i := 0; i < n4; i++ {
		out[i] = in[i+n4]
		out[n-i-1] = -in[n-i-n3_4-1]
	}
	for i := n4; i < n3_4; i++ {
		out[i] = -in[n3_4-i-1]
	}
}

// original c code from stb_vorbis

/*
// this is the original version of the above code, if you want to optimize it from scratch
void inverse_mdct_naive(float *buffer, int n)
{
   float s;
   float A[1 << 12], B[1 << 12], C[1 << 11];
   int i,k,k2,k4, n2 = n >> 1, n4 = n >> 2, n8 = n >> 3, l;
   int n3_4 = n - n4, ld;
   // how can they claim this only uses N words?!
   // oh, because they're only used sparsely, whoops
   float u[1 << 13], X[1 << 13], v[1 << 13], w[1 << 13];
   // set up twiddle factors

   for (k=k2=0; k < n4; ++k,k2+=2) {
      A[k2  ] = (float)  cos(4*k*M_PI/n);
      A[k2+1] = (float) -sin(4*k*M_PI/n);
      B[k2  ] = (float)  cos((k2+1)*M_PI/n/2);
      B[k2+1] = (float)  sin((k2+1)*M_PI/n/2);
   }
   for (k=k2=0; k < n8; ++k,k2+=2) {
      C[k2  ] = (float)  cos(2*(k2+1)*M_PI/n);
      C[k2+1] = (float) -sin(2*(k2+1)*M_PI/n);
   }

   // IMDCT algorithm from "The use of multirate filter banks for coding of high quality digital audio"
   // Note there are bugs in that pseudocode, presumably due to them attempting
   // to rename the arrays nicely rather than representing the way their actual
   // implementation bounces buffers back and forth. As a result, even in the
   // "some formulars corrected" version, a direct implementation fails. These
   // are noted below as "paper bug".

   // copy and reflect spectral data
   for (k=0; k < n2; ++k) u[k] = buffer[k];
   for (   ; k < n ; ++k) u[k] = -buffer[n - k - 1];
   // kernel from paper
   // step 1
   for (k=k2=k4=0; k < n4; k+=1, k2+=2, k4+=4) {
      v[n-k4-1] = (u[k4] - u[n-k4-1]) * A[k2]   - (u[k4+2] - u[n-k4-3])*A[k2+1];
      v[n-k4-3] = (u[k4] - u[n-k4-1]) * A[k2+1] + (u[k4+2] - u[n-k4-3])*A[k2];
   }
   // step 2
   for (k=k4=0; k < n8; k+=1, k4+=4) {
      w[n2+3+k4] = v[n2+3+k4] + v[k4+3];
      w[n2+1+k4] = v[n2+1+k4] + v[k4+1];
      w[k4+3]    = (v[n2+3+k4] - v[k4+3])*A[n2-4-k4] - (v[n2+1+k4]-v[k4+1])*A[n2-3-k4];
      w[k4+1]    = (v[n2+1+k4] - v[k4+1])*A[n2-4-k4] + (v[n2+3+k4]-v[k4+3])*A[n2-3-k4];
   }
   // step 3
   ld = ilog(n) - 1; // ilog is off-by-one from normal definitions
   for (l=0; l < ld-3; ++l) {
      int k0 = n >> (l+2), k1 = 1 << (l+3);
      int rlim = n >> (l+4), r4, r;
      int s2lim = 1 << (l+2), s2;
      for (r=r4=0; r < rlim; r4+=4,++r) {
         for (s2=0; s2 < s2lim; s2+=2) {
            u[n-1-k0*s2-r4] = w[n-1-k0*s2-r4] + w[n-1-k0*(s2+1)-r4];
            u[n-3-k0*s2-r4] = w[n-3-k0*s2-r4] + w[n-3-k0*(s2+1)-r4];
            u[n-1-k0*(s2+1)-r4] = (w[n-1-k0*s2-r4] - w[n-1-k0*(s2+1)-r4]) * A[r*k1]
                                - (w[n-3-k0*s2-r4] - w[n-3-k0*(s2+1)-r4]) * A[r*k1+1];
            u[n-3-k0*(s2+1)-r4] = (w[n-3-k0*s2-r4] - w[n-3-k0*(s2+1)-r4]) * A[r*k1]
                                + (w[n-1-k0*s2-r4] - w[n-1-k0*(s2+1)-r4]) * A[r*k1+1];
         }
      }
      if (l+1 < ld-3) {
         // paper bug: ping-ponging of u&w here is omitted
         memcpy(w, u, sizeof(u));
      }
   }

   // step 4
   for (i=0; i < n8; ++i) {
      int j = bit_reverse(i) >> (32-ld+3);
      assert(j < n8);
      if (i == j) {
         // paper bug: original code probably swapped in place; if copying,
         //            need to directly copy in this case
         int i8 = i << 3;
         v[i8+1] = u[i8+1];
         v[i8+3] = u[i8+3];
         v[i8+5] = u[i8+5];
         v[i8+7] = u[i8+7];
      } else if (i < j) {
         int i8 = i << 3, j8 = j << 3;
         v[j8+1] = u[i8+1], v[i8+1] = u[j8 + 1];
         v[j8+3] = u[i8+3], v[i8+3] = u[j8 + 3];
         v[j8+5] = u[i8+5], v[i8+5] = u[j8 + 5];
         v[j8+7] = u[i8+7], v[i8+7] = u[j8 + 7];
      }
   }
   // step 5
   for (k=0; k < n2; ++k) {
      w[k] = v[k*2+1];
   }
   // step 6
   for (k=k2=k4=0; k < n8; ++k, k2 += 2, k4 += 4) {
      u[n-1-k2] = w[k4];
      u[n-2-k2] = w[k4+1];
      u[n3_4 - 1 - k2] = w[k4+2];
      u[n3_4 - 2 - k2] = w[k4+3];
   }
   // step 7
   for (k=k2=0; k < n8; ++k, k2 += 2) {
      v[n2 + k2 ] = ( u[n2 + k2] + u[n-2-k2] + C[k2+1]*(u[n2+k2]-u[n-2-k2]) + C[k2]*(u[n2+k2+1]+u[n-2-k2+1]))/2;
      v[n-2 - k2] = ( u[n2 + k2] + u[n-2-k2] - C[k2+1]*(u[n2+k2]-u[n-2-k2]) - C[k2]*(u[n2+k2+1]+u[n-2-k2+1]))/2;
      v[n2+1+ k2] = ( u[n2+1+k2] - u[n-1-k2] + C[k2+1]*(u[n2+1+k2]+u[n-1-k2]) - C[k2]*(u[n2+k2]-u[n-2-k2]))/2;
      v[n-1 - k2] = (-u[n2+1+k2] + u[n-1-k2] + C[k2+1]*(u[n2+1+k2]+u[n-1-k2]) - C[k2]*(u[n2+k2]-u[n-2-k2]))/2;
   }
   // step 8
   for (k=k2=0; k < n4; ++k,k2 += 2) {
      X[k]      = v[k2+n2]*B[k2  ] + v[k2+1+n2]*B[k2+1];
      X[n2-1-k] = v[k2+n2]*B[k2+1] - v[k2+1+n2]*B[k2  ];
   }

   // decode kernel to output
   // determined the following value experimentally
   // (by first figuring out what made inverse_mdct_slow work); then matching that here
   // (probably vorbis encoder premultiplies by n or n/2, to save it on the decoder?)
   s = 0.5; // theoretically would be n4

   // [[[ note! the s value of 0.5 is compensated for by the B[] in the current code,
   //     so it needs to use the "old" B values to behave correctly, or else
   //     set s to 1.0 ]]]
   for (i=0; i < n4  ; ++i) buffer[i] = s * X[i+n4];
   for (   ; i < n3_4; ++i) buffer[i] = -s * X[n3_4 - i - 1];
   for (   ; i < n   ; ++i) buffer[i] = -s * X[i - n3_4];
}
*/
