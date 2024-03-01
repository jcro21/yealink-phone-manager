package main

import (
	"fmt"
	"net/http"
	// "os"
	// "syscall"
)

func (a *appContext) handleAPIHealth(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("content-type", "application/json")

	err := a.health()
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(fmt.Sprintf("{\"error\": \"%+v\"}", err)))
	}

	rw.Write([]byte(fmt.Sprintf("{\"version\": \"%s\"}", canary)))
}

func (a *appContext) health() error {
	// var stat syscall.Statfs_t

	// wd, err := os.Getwd()
	// if err != nil {
	// 	return err
	// }

	// err = syscall.Statfs(wd, &stat)
	// if err != nil {
	// 	return err
	// }

	// Available blocks * size per block = available space in bytes
	// freeBytes := int(stat.Bavail * uint64(stat.Bsize))

	// if freeBytes < diskCriticalBytes {
	// 	return fmt.Errorf("disk space critically low")
	// }

	return nil
}
