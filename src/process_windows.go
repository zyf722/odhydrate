//go:build windows

package main

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

func findProcessPIDs(name string) ([]uint32, error) {
	r1, _, e1 := procCreateToolhelp32Snapshot.Call(uintptr(th32csSnapProcess), 0)
	h := syscall.Handle(r1)
	if h == syscall.InvalidHandle || h == 0 {
		if e1 != nil && e1 != syscall.Errno(0) {
			return nil, e1
		}
		return nil, errors.New("CreateToolhelp32Snapshot failed")
	}
	defer syscall.CloseHandle(h)

	var pe processEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	r, _, e := procProcess32FirstW.Call(uintptr(h), uintptr(unsafe.Pointer(&pe)))
	if r == 0 {
		if e != nil && e != syscall.Errno(0) {
			return nil, e
		}
		return nil, nil
	}

	var pids []uint32
	for {
		exe := syscall.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(exe, name) {
			pids = append(pids, pe.ProcessID)
		}
		pe.Size = uint32(unsafe.Sizeof(pe))
		r, _, e = procProcess32NextW.Call(uintptr(h), uintptr(unsafe.Pointer(&pe)))
		if r == 0 {
			break
		}
	}
	if e != nil && e != syscall.Errno(0) && e != syscall.ERROR_NO_MORE_FILES {
		return pids, e
	}
	return pids, nil
}
