package util

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

//DiskID2Path ...
func DiskID2Path2(diskID uint32) (string, bool) {
	path := ""
	switch diskID {
	case 66:
		path = "/dev/sdc1"
	case 80:
		path = "/dev/sde1"
	case 89:
		path = "/dev/sdc3"
	case 90:
		path = "/dev/sdf1"
	default:
		return path, false
	}

	return path, true
}

//DiskID2Path ...For 179.3
func DiskID2Path(diskID uint32) (string, bool) {
	disks := GetConfig().Disks

	for _, disk := range disks {
		if disk.Key == diskID {
			return disk.Value, true
		}
	}

	return "", false
}

//GetSectorCount ...
func GetSectorCount(diskid uint32) uint64 {
	path, ok := DiskID2Path(diskid)
	if !ok {
		return 0
	}
	fd, err := os.Open(path)
	if err != nil {
		fmt.Println("Open file: ", err)
		return 0
	}
	defer fd.Close()

	var result uint64
	_, _, err2 := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), uintptr(2148012658), uintptr(unsafe.Pointer(&result)))
	if err2 != 0 {
		fmt.Println(err2)
	}
	return result / 512
}
