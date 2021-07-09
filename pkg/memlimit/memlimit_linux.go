// +build linux

package memlimit

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"syscall"
)

const (
	defaultMemoryLimit        = uint64(4 * 1024 * 1024 * 1024) // 4 GiB
	memoryLimitCgroupLocation = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
)

func init() {
	limit, err := readMemoryLimit(memoryLimitCgroupLocation)
	if err != nil {
		limit = defaultMemoryLimit
	}
	if err := set(limit); err != nil {
		log.Printf("memory limit is not set: %v", err)
	}
}

func readMemoryLimit(filename string) (uint64, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	limitStr := strings.TrimSuffix(string(d), "\n")
	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return limit, nil
}

func set(limit uint64) error {
	rlimit := syscall.Rlimit{
		Cur: limit,
		Max: limit,
	}
	return syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit)
}
