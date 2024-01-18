package protogenic

import (
	"runtime"
	"sync"
)

var (
	_once               sync.Once
	_protogenicFileName string
	_path               string
)

func GetPathAndExecutable() (string, string) {
	_once.Do(func() {
		os := runtime.GOOS
		switch os {
		case "windows":
			{
				_protogenicFileName = "protogenic.exe"
				_path = ""
			}
		default:
			{
				_protogenicFileName = "protogenic"
				_path = "/"
			}
		}
	})
	return _path, _protogenicFileName
}
