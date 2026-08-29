//go:build linux

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func lockSession(dry bool) error {
	if dry {
		log.Printf("DRY RUN: lock session")
		return nil
	}
	cmds := [][]string{{"loginctl", "lock-session"}, {"xdg-screensaver", "lock"}, {"dm-tool", "lock"}}
	for _, c := range cmds {
		if _, err := exec.LookPath(c[0]); err == nil {
			if exec.Command(c[0], c[1:]...).Run() == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no supported session locker found")
}

func unlockSession(dry bool) error {
	if dry {
		log.Printf("DRY RUN: presence verified; unlock permitted")
		return nil
	}
	if _, err := exec.LookPath("loginctl"); err == nil {
		if exec.Command("loginctl", "unlock-session").Run() == nil {
			return nil
		}
	}
	return fmt.Errorf("presence verified, but this Linux lock screen does not accept loginctl unlock; PAM adapter is required")
}
