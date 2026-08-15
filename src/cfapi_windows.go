//go:build windows

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

func (e *hresultError) Error() string {
	return fmt.Sprintf("%s HRESULT=0x%08X", e.Op, e.HR)
}

func ensureCFAPI() error {
	if err := cldapi.Load(); err != nil {
		return fmt.Errorf("cannot load CldApi.dll (Windows 10 version 1709 or later is required): %w", err)
	}
	for name, proc := range map[string]*syscall.LazyProc{
		"CfGetPlaceholderInfo":    procCfGetPlaceholderInfo,
		"CfGetSyncRootInfoByPath": procCfGetSyncRootInfoByPath,
		"CfOpenFileWithOplock":    procCfOpenFileWithOplock,
		"CfUpdatePlaceholder":     procCfUpdatePlaceholder,
		"CfCloseHandle":           procCfCloseHandle,
	} {
		if err := proc.Find(); err != nil {
			return fmt.Errorf("cannot resolve %s: %w", name, err)
		}
	}
	return nil
}

func getPlaceholderInfo(path string) (placeholderInfo, error) {
	var out placeholderInfo
	p, err := syscall.UTF16PtrFromString(win32Path(path))
	if err != nil {
		return out, err
	}
	h, err := syscall.CreateFile(
		p,
		fileReadAttributes,
		fileShareRead|fileShareWrite|fileShareDelete,
		nil,
		openExisting,
		0,
		0,
	)
	if err != nil {
		return out, fmt.Errorf("CreateFile(FILE_READ_ATTRIBUTES): %w", err)
	}
	defer syscall.CloseHandle(h)
	return getPlaceholderInfoByHandle(h)
}

func getPlaceholderInfoByHandle(h syscall.Handle) (placeholderInfo, error) {
	var out placeholderInfo
	buf := make([]byte, 4096)
	var returned uint32
	r1, _, _ := procCfGetPlaceholderInfo.Call(
		uintptr(h),
		uintptr(cfPlaceholderInfoStandard),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&returned)),
	)
	hr := uint32(r1)
	if hr != 0 {
		return out, &hresultError{Op: "CfGetPlaceholderInfo", HR: hr}
	}
	if returned < 60 {
		return out, fmt.Errorf("CfGetPlaceholderInfo returned only %d bytes", returned)
	}
	out.OnDiskDataSize = int64(le64(buf[0:8]))
	out.ValidatedDataSize = int64(le64(buf[8:16]))
	out.ModifiedDataSize = int64(le64(buf[16:24]))
	out.PropertiesSize = int64(le64(buf[24:32]))
	out.PinState = le32(buf[32:36])
	out.InSyncState = le32(buf[36:40])
	out.FileID = int64(le64(buf[40:48]))
	out.SyncRootFileID = int64(le64(buf[48:56]))
	out.FileIdentityLen = le32(buf[56:60])
	return out, nil
}

func getSyncRootInfo(path string) (syncRootInfo, error) {
	var out syncRootInfo
	p, err := syscall.UTF16PtrFromString(win32Path(path))
	if err != nil {
		return out, err
	}
	buf := make([]byte, 8192)
	var returned uint32
	r1, _, _ := procCfGetSyncRootInfoByPath.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(cfSyncRootInfoStandard),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&returned)),
	)
	hr := uint32(r1)
	if hr != 0 {
		return out, &hresultError{Op: "CfGetSyncRootInfoByPath", HR: hr}
	}

	// CF_SYNC_ROOT_STANDARD_INFO has a 1056-byte fixed prefix on the Windows ABI:
	// 8 + 4 + 4 + 4 + 4 + 4 + WCHAR[256] + WCHAR[256] + 4.
	if returned < 1056 {
		return out, fmt.Errorf("CfGetSyncRootInfoByPath returned only %d bytes", returned)
	}
	out.SyncRootFileID = int64(le64(buf[0:8]))
	out.HydrationPrimary = le16(buf[8:10])
	out.ProviderName = utf16BytesToString(buf[28:540])
	out.ProviderVersion = utf16BytesToString(buf[540:1052])
	return out, nil
}

func openProtectedExclusive(path string) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(win32Path(path))
	if err != nil {
		return 0, err
	}
	var h syscall.Handle
	r1, _, _ := procCfOpenFileWithOplock.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(cfOpenFileFlagExclusive|cfOpenFileFlagWrite),
		uintptr(unsafe.Pointer(&h)),
	)
	hr := uint32(r1)
	if hr != 0 {
		return 0, &hresultError{Op: "CfOpenFileWithOplock(EXCLUSIVE|WRITE_ACCESS)", HR: hr}
	}
	if h == 0 || h == syscall.InvalidHandle {
		return 0, errors.New("CfOpenFileWithOplock returned an invalid protected handle")
	}
	return h, nil
}

func closeProtectedHandle(h syscall.Handle) {
	if h != 0 && h != syscall.InvalidHandle {
		procCfCloseHandle.Call(uintptr(h))
	}
}

func updatePlaceholderDehydrate(h syscall.Handle) error {
	flags := cfUpdateFlagVerifySync | cfUpdateFlagDehydrate
	// HRESULT CfUpdatePlaceholder(HANDLE, FsMetadata, FileIdentity, FileIdentityLength,
	//   DehydrateRangeArray, DehydrateRangeCount, UpdateFlags, UpdateUsn, Overlapped)
	// All optional mutation inputs are NULL. The only requested change is dehydration,
	// guarded by VERIFY_IN_SYNC.
	r1, _, _ := procCfUpdatePlaceholder.Call(
		uintptr(h),
		0, // FsMetadata = NULL
		0, // FileIdentity = NULL
		0, // FileIdentityLength = 0
		0, // DehydrateRangeArray = NULL
		0, // DehydrateRangeCount = 0
		uintptr(flags),
		0, // UpdateUsn = NULL
		0, // Overlapped = NULL -> synchronous
	)
	hr := uint32(r1)
	if hr != 0 {
		return &hresultError{Op: "CfUpdatePlaceholder(VERIFY_IN_SYNC|DEHYDRATE)", HR: hr}
	}
	return nil
}

func isSafeCandidate(info placeholderInfo) (bool, string) {
	if info.PinState != cfPinStateUnpinned {
		return false, fmt.Sprintf("PinState=%d, want UNPINNED(2)", info.PinState)
	}
	if info.InSyncState != cfInSyncStateInSync {
		return false, fmt.Sprintf("InSyncState=%d, want IN_SYNC(1)", info.InSyncState)
	}
	if info.ModifiedDataSize != 0 {
		return false, fmt.Sprintf("ModifiedDataSize=%d, want 0", info.ModifiedDataSize)
	}
	if info.OnDiskDataSize <= 0 {
		return false, "OnDiskDataSize=0"
	}
	return true, ""
}

func isOneDriveProvider(name string) bool {
	return strings.Contains(strings.ToLower(name), "onedrive")
}

func enumLabelHydrationPrimary(v uint16) string {
	switch v {
	case 0:
		return " (PARTIAL)"
	case 1:
		return " (PROGRESSIVE)"
	case 2:
		return " (FULL)"
	case 3:
		return " (ALWAYS_FULL)"
	default:
		return ""
	}
}

func enumLabelPin(v uint32) string {
	switch v {
	case cfPinStateUnspecified:
		return " (UNSPECIFIED)"
	case cfPinStatePinned:
		return " (PINNED)"
	case cfPinStateUnpinned:
		return " (UNPINNED)"
	case cfPinStateExcluded:
		return " (EXCLUDED)"
	case cfPinStateInherit:
		return " (INHERIT)"
	default:
		return ""
	}
}

func enumLabelSync(v uint32) string {
	if v == cfInSyncStateInSync {
		return " (IN_SYNC)"
	}
	return ""
}

func utf16BytesToString(b []byte) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		v := le16(b[i : i+2])
		if v == 0 {
			break
		}
		u = append(u, v)
	}
	return syscall.UTF16ToString(u)
}

func le16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func le32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func le64(b []byte) uint64 {
	return uint64(le32(b[0:4])) | uint64(le32(b[4:8]))<<32
}

func win32Path(path string) string {
	if strings.HasPrefix(path, `\\?\`) {
		return path
	}
	if strings.HasPrefix(path, `\\`) {
		return `\\?\UNC\` + strings.TrimPrefix(path, `\\`)
	}
	return `\\?\` + path
}
