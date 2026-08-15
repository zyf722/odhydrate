//go:build windows

package main

import (
	"sync/atomic"
	"syscall"
	"time"
)

var version = "0.1.0"

const (
	fileReadAttributes              = 0x00000080
	fileAttributeDirectory          = 0x00000010
	fileAttributeReparse            = 0x00000400
	fileAttributeOffline            = 0x00001000
	fileAttributeRecallOnOpen       = 0x00040000
	fileAttributePinned             = 0x00080000
	fileAttributeUnpinned           = 0x00100000
	fileAttributeRecallOnDataAccess = 0x00400000
	ioReparseTagMountPoint          = 0xA0000003
	ioReparseTagSymlink             = 0xA000000C
	openExisting                    = 3
	fileShareRead                   = 0x00000001
	fileShareWrite                  = 0x00000002
	fileShareDelete                 = 0x00000004

	cfPlaceholderInfoStandard = 1
	cfPinStateUnspecified     = 0
	cfPinStatePinned          = 1
	cfPinStateUnpinned        = 2
	cfPinStateExcluded        = 3
	cfPinStateInherit         = 4
	cfInSyncStateInSync       = 1

	cfSyncRootInfoStandard  = 1
	cfHydrationAlwaysFull   = 3
	cfOpenFileFlagExclusive = 0x00000001
	cfOpenFileFlagWrite     = 0x00000002
	cfUpdateFlagVerifySync  = 0x00000001
	cfUpdateFlagDehydrate   = 0x00000004

	hresultNotACloudFile = 0x80070178
	th32csSnapProcess    = 0x00000002
)

var (
	cldapi                      = syscall.NewLazyDLL("CldApi.dll")
	procCfGetPlaceholderInfo    = cldapi.NewProc("CfGetPlaceholderInfo")
	procCfGetSyncRootInfoByPath = cldapi.NewProc("CfGetSyncRootInfoByPath")
	procCfOpenFileWithOplock    = cldapi.NewProc("CfOpenFileWithOplock")
	procCfUpdatePlaceholder     = cldapi.NewProc("CfUpdatePlaceholder")
	procCfCloseHandle           = cldapi.NewProc("CfCloseHandle")

	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleOutputCP       = kernel32.NewProc("SetConsoleOutputCP")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
)

type placeholderInfo struct {
	OnDiskDataSize    int64
	ValidatedDataSize int64
	ModifiedDataSize  int64
	PropertiesSize    int64
	PinState          uint32
	InSyncState       uint32
	FileID            int64
	SyncRootFileID    int64
	FileIdentityLen   uint32
}

type syncRootInfo struct {
	SyncRootFileID   int64
	HydrationPrimary uint16
	ProviderName     string
	ProviderVersion  string
}

type hresultError struct {
	Op string
	HR uint32
}

type fileTask struct {
	path       string
	logical    int64
	attributes uint32
	reparseTag uint32
}

type scanResult struct {
	Status     string
	Reason     string
	Path       string
	Logical    int64
	Info       placeholderInfo
	Attributes uint32
	ReparseTag uint32
}

type candidate struct {
	status  string
	path    string
	logical int64
	onDisk  int64
}

type counters struct {
	filesSeen          atomic.Uint64
	dirsSeen           atomic.Uint64
	placeholders       atomic.Uint64
	unpinnedEnum       atomic.Uint64
	queried            atomic.Uint64
	resident           atomic.Uint64
	safe               atomic.Uint64
	review             atomic.Uint64
	residentBytes      atomic.Int64
	safeBytes          atomic.Int64
	reviewBytes        atomic.Int64
	notCloud           atomic.Uint64
	queryErrors        atomic.Uint64
	enumErrors         atomic.Uint64
	skippedReparseDirs atomic.Uint64
}

type scanConfig struct {
	root     string
	csvPath  string
	workers  int
	progress time.Duration
	deep     bool
	noCSV    bool
	topN     int
	redact   bool
}

type repairConfig struct {
	root       string
	apply      bool
	workers    int
	progress   time.Duration
	reportPath string
	noReport   bool
	redact     bool
	limit      int
}

type repairCandidate struct {
	path    string
	logical int64
	onDisk  int64
}

type repairOutcome struct {
	index       int
	status      string
	reason      string
	path        string
	logical     int64
	before      int64
	after       int64
	freed       int64
	pinState    uint32
	inSyncState uint32
	modified    int64
}

type processEntry32 struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	CntThreads      uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [syscall.MAX_PATH]uint16
}
