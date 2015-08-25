package main

import (
	"fmt"
	"os"
)

// TODO extract into package or such
//
// because this is "frikkin rad" it's basically upload
// with the args moved up one.

type RunCmd struct {
	UploadCmd
}

func (r *RunCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron run [-zip my.zip] -name NAME [OPTIONS] some/image[:tag] [command...]`)
	r.flags.PrintDefaults()
}

func (r *RunCmd) Args() error {
	r.UploadCmd.codes.Host = "true"
	return r.UploadCmd.Args()
}
