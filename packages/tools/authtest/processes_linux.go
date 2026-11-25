package authtest

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func liveGroup(group int) (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, errors.New("auth fixture: process inspection unavailable")
	}
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		stat, err := os.ReadFile("/proc/" + entry.Name() + "/stat")
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return false, errors.New("auth fixture: process inspection failed")
		}
		end := strings.LastIndexByte(string(stat), ')')
		if end < 0 {
			return false, errors.New("auth fixture: invalid process state")
		}
		fields := strings.Fields(string(stat)[end+1:])
		if len(fields) < 3 {
			return false, errors.New("auth fixture: incomplete process state")
		}
		pgid, err := strconv.Atoi(fields[2])
		if err != nil {
			return false, errors.New("auth fixture: invalid process group")
		}
		if pgid == group && fields[0] != "Z" && fields[0] != "X" {
			return true, nil
		}
	}
	return false, nil
}

func exitedUnreaped(pid int) (bool, error) {
	stat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false, errors.New("auth fixture: owned leader state unavailable")
	}
	end := strings.LastIndexByte(string(stat), ')')
	if end < 0 {
		return false, errors.New("auth fixture: invalid leader state")
	}
	fields := strings.Fields(string(stat)[end+1:])
	if len(fields) < 1 {
		return false, errors.New("auth fixture: incomplete leader state")
	}
	return fields[0] == "Z" || fields[0] == "X", nil
}
