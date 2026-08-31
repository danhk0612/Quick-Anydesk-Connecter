//go:build windows

package main

import "syscall"

var (
	comctl32 = syscall.NewLazyDLL("comctl32.dll")

	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procSetTimer             = user32.NewProc("SetTimer")
	procKillTimer            = user32.NewProc("KillTimer")
)
