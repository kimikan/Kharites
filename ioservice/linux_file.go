package ioservice

import (
	"fmt"
	"os"
)

const (
	fileError     = -1
	fileEOF       = -0x10
	filePathError = -0x11
	fileReadError = -0x12
)

/*ReadBytesFromFile size buf from selected file*/
func ReadBytesFromFile(file string, data []byte, offset int64) int {
	_fio := new(fio)
	if _fio.Open(file) {
		nums, err := _fio.ReadBytes(data, offset)
		defer _fio.Close()
		if err != nil {
			fmt.Println("file read error:", file, offset, err)
			return fileReadError
		}

		if nums <= 0 {
			fmt.Println("no data read:", file, offset, len(data), nums)
			return fileEOF
		}
		return nums
	}

	fmt.Println("file path error:", file, offset)
	return filePathError
}

type fio struct {
	r *os.File
}

func (_fio *fio) Open(str string) bool {
	var err error
	_fio.r, err = os.Open(str)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (_fio *fio) ReadBytes(data []byte, offset int64) (int, error) {
	return _fio.r.ReadAt(data, offset)
}

func (_fio *fio) Close() {
	if _fio.r != nil {
		_fio.r.Close()
	}
}
