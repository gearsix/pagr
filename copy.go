package main

/*
#include <stdio.h>
#include <stdlib.h>

// copyf copies data from file at `src` to the file at `dst`
// in 4kb chunks`.
// The file @ `src` *must* be **less than 2GB**. This is a
// limitation of libc (staying portable).
// Any existing file @ `dst` will be overwritten.
// returns EXIT_SUCCESS or EXIT_FAILURE.
int copyf(const char *src, const char *dst)
{
	int ret = EXIT_FAILURE;

	FILE *srcf = fopen(src, "rb"), *dstf = fopen(dst, "wb");
	if (!src || !dst) goto ABORT;

	fseek(srcf, 0, SEEK_END);
	size_t siz = ftell(srcf); // 2GB limit, returns long int
	rewind(srcf);

	char buf[4096]; // 4kb chunks
	size_t r, w, total = 0;
	while ((r = fread(buf, sizeof(char), sizeof(buf), srcf)) > 0) {
		if (ferror(srcf)) goto ABORT;
		w = fwrite(buf, sizeof(char), r, dstf);
		if (ferror(dstf)) goto ABORT;
		total += w;
	}

	if (total == siz) ret = EXIT_SUCCESS;

ABORT:
	if (srcf) fclose(srcf);
	if (dstf) fclose(dstf);
	return ret;
}
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"
)

func copyFile(src, dst string) (err error) {
	var srcf, dstf *os.File
	if srcf, err = os.Open(src); err != nil {
		return err
	}
	defer srcf.Close()
	if dstf, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return err
	}
	defer dstf.Close()

	if _, err = io.Copy(dstf, srcf); err != nil {
		return err
	}
	return dstf.Sync()
}

func CopyFile(src, dst string) (err error) {
	var srcfi, dstfi os.FileInfo

	if srcfi, err = os.Stat(src); err != nil {
		return err
	} else if !srcfi.Mode().IsRegular() {
		return fmt.Errorf("cannot copy from non-regular source file %s (%q)",
			srcfi.Name(), srcfi.Mode().String())
	}

	if dstfi, err = os.Stat(dst); err != nil && !os.IsNotExist(err) {
		return err
	} else if dstfi != nil && !dstfi.Mode().IsRegular() {
		return fmt.Errorf("cannot copy to non-regular destination file %s (%q)",
			dstfi.Name(), dstfi.Mode().String())
	} else if os.SameFile(srcfi, dstfi) {
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	// only copy if dst doesnt exist or has different name/size/modtime
	// and has a size less than 2GB (libc limit)
	if dstfi == nil || srcfi.Name() != dstfi.Name() ||
		srcfi.Size() != dstfi.Size() || srcfi.ModTime() != dstfi.ModTime() {
		if srcfi.Size() > 2000000000 {
			copyFile(src, dst)
		} else {
			cSrc := C.CString(src)
			cDst := C.CString(dst)
			if uint32(C.copyf(cSrc, cDst)) != 0 {
				err = fmt.Errorf("copyf failed ('%s' -> '%s')", src, dst)
			}
			C.free(unsafe.Pointer(cSrc))
			C.free(unsafe.Pointer(cDst))
		}
	}

	return
}
