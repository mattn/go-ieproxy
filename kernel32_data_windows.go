package ieproxy

import "golang.org/x/sys/windows"

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")
var globalFree = kernel32.NewProc("GlobalFree")
