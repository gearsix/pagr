package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(test *testing.T) {
	test.Parallel()

	tdir := test.TempDir()
	src := filepath.Join(tdir, "/src")
	srcData := []byte("data")
	dst := filepath.Join(tdir, "/dst")

	if err := os.WriteFile(src, srcData, 0666); err != nil {
		test.Error("setup failed, could not write", tdir+"/src")
	}

	if err := CopyFile(src, dst); err != nil {
		test.Fatal("CopyFile failed", err)
	}
	if _, err := os.Stat(dst); err != nil {
		test.Fatalf("could not stat '%s'", dst)
	}

	if buf, err := os.ReadFile(dst); err != nil {
		test.Errorf("could not read '%s'", dst)
	} else if len(buf) < len(srcData) {
		test.Fatalf("not all srcData (%s) copied to '%s' (%s)", srcData, dst, buf)
	} else if string(buf) != string(srcData) {
		test.Fatalf("copied srcData (%s) does not match source (%s)", buf, srcData)
	}
}
