package main

import (
	"os"
	"testing"
)

func TestCopyFile(test *testing.T) {
	test.Parallel()

	var err error
	tdir := test.TempDir()
	src := tdir + "/src"
	dst := tdir + "/dst"
	data := []byte("data")
	if err = os.WriteFile(src, data, 0666); err != nil {
		test.Error("setup failed, could not write", tdir+"/src")
	}

	if err = CopyFile(src, dst); err != nil {
		test.Fatal("CopyFile failed", err)
	}
	if _, err = os.Stat(dst); err != nil {
		test.Fatalf("could not stat '%s'", dst)
	}
	var buf []byte
	if buf, err = os.ReadFile(dst); err != nil {
		test.Errorf("could not read '%s'", dst)
	} else if len(buf) < len(data) {
		test.Fatalf("not all data (%s) copied to '%s' (%s)", data, dst, buf)
	} else if string(buf) != string(data) {
		test.Fatalf("copied data (%s) does not match source (%s)", buf, data)
	}
}
