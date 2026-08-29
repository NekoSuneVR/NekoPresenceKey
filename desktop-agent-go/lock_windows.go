//go:build windows

package main

import (
	"fmt"
	"log"
	"syscall"
)

func lockSession(dry bool) error {
	if dry {
		log.Printf("DRY RUN: lock session")
		return nil
	}
	dll := syscall.NewLazyDLL("user32.dll")
	proc := dll.NewProc("LockWorkStation")
	r, _, e := proc.Call()
	if r == 0 {
		return e
	}
	return nil
}

func unlockSession(dry bool) error {
	if dry {
		log.Printf("DRY RUN: presence verified; Windows Credential Provider unlock permitted")
		return nil
	}
	return fmt.Errorf("presence verified; Windows requires the included Credential Provider adapter for secure unlock")
}
