package opencl

import (
	"unsafe"

	"github.com/mumax/3cl/opencl/cl"
	"github.com/mumax/3cl/data"
)

// Add exchange field to Beff.
// 	m: normalized magnetization
// 	B: effective field in Tesla
// 	Aex_red: Aex / (Msat * 1e18 m2)
// see exchange.cu
func AddExchange(B, m *data.Slice, Aex_red SymmLUT, regions *Bytes, mesh *data.Mesh) {
	c := mesh.CellSize()
	wx := float32(2 * 1e-18 / (c[X] * c[X]))
	wy := float32(2 * 1e-18 / (c[Y] * c[Y]))
	wz := float32(2 * 1e-18 / (c[Z] * c[Z]))
	N := mesh.Size()
	pbc := mesh.PBC_code()
	cfg := make3DConf(N)
	event := k_addexchange_async(B.DevPtr(X), B.DevPtr(Y), B.DevPtr(Z),
				     m.DevPtr(X), m.DevPtr(Y), m.DevPtr(Z),
				     unsafe.Pointer(Aex_red), regions.Ptr,
				     wx, wy, wz, N[X], N[Y], N[Z], pbc, cfg,
				     [](*cl.Event){B.GetEvent(X), B.GetEvent(Y),
				     B.GetEvent(Z), m.GetEvent(X), m.GetEvent(Y), m.GetEvent(Z)})
	B.SetEvent(X, event)
	B.SetEvent(Y, event)
	B.SetEvent(Z, event)
	m.SetEvent(X, event)
	m.SetEvent(Y, event)
	m.SetEvent(Z, event)
}

// Finds the average exchange strength around each cell, for debugging.
func ExchangeDecode(dst *data.Slice, Aex_red SymmLUT, regions *Bytes, mesh *data.Mesh) {
	c := mesh.CellSize()
	wx := float32(2 * 1e-18 / (c[X] * c[X]))
	wy := float32(2 * 1e-18 / (c[Y] * c[Y]))
	wz := float32(2 * 1e-18 / (c[Z] * c[Z]))
	N := mesh.Size()
	pbc := mesh.PBC_code()
	cfg := make3DConf(N)
	event := k_exchangedecode_async(dst.DevPtr(0), unsafe.Pointer(Aex_red), regions.Ptr,
					wx, wy, wz, N[X], N[Y], N[Z], pbc, cfg, [](*cl.Event){dst.GetEvent(0)})
	dst.SetEvent(0, event)
}