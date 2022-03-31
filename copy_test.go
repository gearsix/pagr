package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(test *testing.T) {
	test.Parallel()

	tdir := filepath.Join(os.TempDir(), "pagr_test", "TestCopyFile")
	if err := os.MkdirAll(tdir, 0775); err != nil {
		test.Errorf("failed to create temporary test dir: %s", tdir)
	}
	src := filepath.Join(tdir, "/src")
	srcData := []byte("data")
	dst := filepath.Join(tdir, "/dst")

	if err := ioutil.WriteFile(src, srcData, 0666); err != nil {
		test.Error("setup failed, could not write", tdir+"/src")
	}

	if err := CopyFile(src, dst); err != nil {
		test.Fatal("CopyFile failed", err)
	}
	if _, err := os.Stat(dst); err != nil {
		test.Fatalf("could not stat '%s'", dst)
	}

	if buf, err := ioutil.ReadFile(dst); err != nil {
		test.Errorf("could not read '%s'", dst)
	} else if len(buf) < len(srcData) {
		test.Fatalf("not all srcData (%s) copied to '%s' (%s)", srcData, dst, buf)
	} else if string(buf) != string(srcData) {
		test.Fatalf("copied srcData (%s) does not match source (%s)", buf, srcData)
	}

	if err := os.RemoveAll(tdir); err != nil {
		test.Error(err)
	}
}
