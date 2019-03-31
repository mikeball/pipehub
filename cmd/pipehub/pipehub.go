package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
)

func fatal(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func wait() {
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}

func asyncErrHandler(err error) {
	fmt.Println(errors.Wrap(err, "async error occurred").Error())
	done <- syscall.SIGTERM
}

func loadConfig(path string) ([]byte, error) {
	payload, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "load file error")
	}
	return payload, nil
}
