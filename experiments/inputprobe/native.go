//go:build linux && cgo

package main

/*
#cgo pkg-config: libevdev
#include <errno.h>
#include <stdlib.h>
#include <time.h>
#include "native.h"
static long long probe_monotonic_ns(void) {
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) return -1;
    return (long long)ts.tv_sec * 1000000000LL + ts.tv_nsec;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

type event struct {
	kind  uint16
	code  uint16
	value int32
}

type axisInfo struct {
	Code       int `json:"code"`
	Minimum    int `json:"minimum"`
	Maximum    int `json:"maximum"`
	Resolution int `json:"resolution"`
}

type deviceInfo struct {
	Name          string     `json:"name"`
	Slots         int        `json:"slots"`
	FourthContact bool       `json:"fourth_contact_indicator"`
	Axes          []axisInfo `json:"axes"`
}

type physical struct {
	ptr       *C.struct_probe_device
	eventTime time.Duration
}
type virtual struct{ ptr *C.struct_probe_virtual }

func nativeError(operation string, rc C.int) error {
	if rc >= 0 {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, syscall.Errno(-rc))
}

func openPhysical(path string) (*physical, error) {
	s := C.CString(path)
	defer C.free(unsafe.Pointer(s))
	p := &physical{}
	if err := nativeError("open selected touchpad", C.probe_open(s, &p.ptr)); err != nil {
		return nil, err
	}
	if C.probe_is_touchpad(p.ptr) == 0 {
		p.close()
		return nil, errors.New("inspect selected touchpad: not a supported physical multitouch device")
	}
	return p, nil
}

func (p *physical) close() {
	if p.ptr != nil {
		C.probe_close(p.ptr)
		p.ptr = nil
	}
}
func (p *physical) grab() error { return nativeError("grab selected touchpad", C.probe_grab(p.ptr)) }
func (p *physical) summary() (contacts, palms, buttons int) {
	var s C.struct_probe_summary
	C.probe_summary(p.ptr, &s)
	return int(s.contacts), int(s.palms), int(s.buttons)
}

func (p *physical) currentSlot() int {
	return int(C.libevdev_get_current_slot(p.ptr.dev))
}

// snapshot restores full virtual state after suppressed incremental reports.
func (p *physical) snapshot() []event {
	var frame []event
	for slot := 0; slot < int(C.libevdev_get_num_slots(p.ptr.dev)); slot++ {
		frame = append(frame, event{3, 47, int32(slot)})
		for code := C.uint(C.ABS_MT_TOUCH_MAJOR); code <= C.ABS_MT_TOOL_Y; code++ {
			if C.libevdev_has_event_code(p.ptr.dev, C.EV_ABS, code) != 0 {
				frame = append(frame, event{3, uint16(code), int32(C.libevdev_get_slot_value(p.ptr.dev, C.uint(slot), code))})
			}
		}
	}
	for code := C.uint(0); code < C.ABS_MT_SLOT; code++ {
		if C.libevdev_has_event_code(p.ptr.dev, C.EV_ABS, code) != 0 {
			frame = append(frame, event{3, uint16(code), int32(C.libevdev_get_event_value(p.ptr.dev, C.EV_ABS, code))})
		}
	}
	for code := C.uint(0); code <= C.KEY_MAX; code++ {
		if C.libevdev_has_event_code(p.ptr.dev, C.EV_KEY, code) != 0 {
			frame = append(frame, event{1, uint16(code), int32(C.libevdev_get_event_value(p.ptr.dev, C.EV_KEY, code))})
		}
	}
	return append(frame, event{3, 47, int32(p.currentSlot())}, event{0, 0, 0})
}

func (p *physical) countKeys() []uint16 {
	var keys []uint16
	for _, key := range []C.uint{C.BTN_TOUCH, C.BTN_TOOL_FINGER, C.BTN_TOOL_DOUBLETAP,
		C.BTN_TOOL_TRIPLETAP, C.BTN_TOOL_QUADTAP, C.BTN_TOOL_QUINTTAP} {
		if C.libevdev_has_event_code(p.ptr.dev, C.EV_KEY, key) != 0 {
			keys = append(keys, uint16(key))
		}
	}
	return keys
}

func (p *physical) info() deviceInfo {
	i := deviceInfo{Name: C.GoString(C.libevdev_get_name(p.ptr.dev)), Slots: int(C.libevdev_get_num_slots(p.ptr.dev))}
	i.FourthContact = C.libevdev_has_event_code(p.ptr.dev, C.EV_KEY, C.BTN_TOOL_QUADTAP) != 0
	for _, code := range []C.uint{C.ABS_MT_POSITION_X, C.ABS_MT_POSITION_Y, C.ABS_MT_TRACKING_ID, C.ABS_MT_TOOL_TYPE} {
		a := C.libevdev_get_abs_info(p.ptr.dev, code)
		if a != nil {
			i.Axes = append(i.Axes, axisInfo{int(code), int(a.minimum), int(a.maximum), int(a.resolution)})
		}
	}
	return i
}

func (p *physical) next() (event, bool, error) {
	var ev C.struct_input_event
	rc := C.probe_next(p.ptr, &ev)
	if rc == -C.EAGAIN {
		return event{}, false, nil
	}
	if rc == C.LIBEVDEV_READ_STATUS_SYNC {
		return event{}, false, errors.New("read selected touchpad: input synchronization lost; probe stopped")
	}
	if err := nativeError("read selected touchpad", rc); err != nil {
		return event{}, false, err
	}
	p.eventTime = time.Duration(ev.time.tv_sec)*time.Second + time.Duration(ev.time.tv_usec)*time.Microsecond
	return event{uint16(ev._type), uint16(ev.code), int32(ev.value)}, true, nil
}

func monotonicNow() (time.Duration, error) {
	n := C.probe_monotonic_ns()
	if n < 0 {
		return 0, errors.New("read monotonic clock: clock_gettime failed")
	}
	return time.Duration(n), nil
}

func (p *physical) contacts() ([]contact, error) {
	rx := C.libevdev_get_abs_resolution(p.ptr.dev, C.ABS_MT_POSITION_X)
	ry := C.libevdev_get_abs_resolution(p.ptr.dev, C.ABS_MT_POSITION_Y)
	if rx <= 0 || ry <= 0 {
		return nil, errors.New("measure contacts: positive axis resolution is required")
	}
	var contacts []contact
	for slot := 0; slot < int(C.libevdev_get_num_slots(p.ptr.dev)); slot++ {
		id := C.libevdev_get_slot_value(p.ptr.dev, C.uint(slot), C.ABS_MT_TRACKING_ID)
		if id < 0 {
			continue
		}
		palm := C.libevdev_has_event_code(p.ptr.dev, C.EV_ABS, C.ABS_MT_TOOL_TYPE) != 0 &&
			C.libevdev_get_slot_value(p.ptr.dev, C.uint(slot), C.ABS_MT_TOOL_TYPE) == C.MT_TOOL_PALM
		x := C.libevdev_get_slot_value(p.ptr.dev, C.uint(slot), C.ABS_MT_POSITION_X)
		y := C.libevdev_get_slot_value(p.ptr.dev, C.uint(slot), C.ABS_MT_POSITION_Y)
		contacts = append(contacts, contact{id: int(id), x: float64(x) / float64(rx), y: float64(y) / float64(ry), palm: palm})
	}
	return contacts, nil
}

func createVirtual(source *physical, touchpad bool, name string) (*virtual, error) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	var src *C.struct_probe_device
	if source != nil {
		src = source.ptr
	}
	var kind C.int
	if touchpad {
		kind = 1
	}
	v := &virtual{}
	if err := nativeError("create virtual device (requires /dev/uinput access)", C.probe_create(src, kind, n, &v.ptr)); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *virtual) close() {
	if v.ptr != nil {
		C.probe_destroy(v.ptr)
		v.ptr = nil
	}
}
func (v *virtual) node() string { return C.GoString(C.probe_node(v.ptr)) }
func (v *virtual) write(e event) error {
	return nativeError("submit virtual input event", C.probe_write(v.ptr, C.uint(e.kind), C.uint(e.code), C.int(e.value)))
}
