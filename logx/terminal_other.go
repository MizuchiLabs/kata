//go:build !linux && !windows && !(darwin || freebsd || openbsd || netbsd || dragonfly)

package logx

func isTerminal(uintptr) bool { return false }
