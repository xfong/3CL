package opencl

import (
	"fmt"

	"github.com/mumax/3cl/data"
	"github.com/mumax/3cl/opencl/cl"
	"github.com/mumax/3cl/util"
)

func SetPhi(s *data.Slice, m *data.Slice) {
	N := s.Size()
	util.Argument(m.Size() == N)
	cfg := make3DConf(N)
	eventList := make([]*cl.Event, 1)
	eventList[0] = k_setPhi_async(s.DevPtr(X), m.DevPtr(X), m.DevPtr(Y), N[X], N[Y], N[Z], cfg,
		[](*cl.Event){s.GetEvent(0), m.GetEvent(X), m.GetEvent(Y)})
	err := cl.WaitForEvents(eventList)
	if err != nil {
		fmt.Printf("WaitForEvents failed in mul: %+v \n", err)
	}
	return
}

func SetTheta(s *data.Slice, m *data.Slice) {
	N := s.Size()
	util.Argument(m.Size() == N)
	cfg := make3DConf(N)
	eventList := make([]*cl.Event, 1)
	eventList[0] = k_setTheta_async(s.DevPtr(X), m.DevPtr(Z), N[X], N[Y], N[Z], cfg,
		[](*cl.Event){s.GetEvent(0), m.GetEvent(Z)})
	err := cl.WaitForEvents(eventList)
	if err != nil {
		fmt.Printf("WaitForEvents failed in mul: %+v \n", err)
	}
	return
}
