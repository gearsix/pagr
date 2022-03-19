package main

/*
#include <stdio.h>
#include <stdlib.h>

// copyf copies data from file at `src` to the file at `dst`
// in 4kb chunks`.
// any existing file @ `dst` will be overwritten.
// returns EXIT_SUCCESS or EXIT_FAILURE.
int copyf(const char *src, const char *dst)
{
	int ret = EXIT_FAILURE;

	FILE *srcf, *dstf;
	if ((!(srcf = fopen(src, "rb"))) ||
		(!(dstf = fopen(dst, "wb"))))
		goto ABORT;

	fseek(srcf, 0, SEEK_END);
	size_t siz = ftell(srcf);
	rewind(srcf);

	char buf[4096]; // 4kb chunks
	size_t r, w, total = 0;
	do {
		r = fread(buf, sizeof(char), sizeof(buf), srcf);
		if (ferror(srcf)) goto ABORT;
		else {
			w = fwrite(buf, sizeof(char), r, dstf);
			if (ferror(dstf)) goto ABORT;
			total += w;
		}
	} while (!feof(srcf));
	
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
	"os"
	"path/filepath"
	"unsafe"
)

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
	if dstfi == nil || srcfi.Name() != dstfi.Name() ||
		srcfi.Size() != dstfi.Size() || srcfi.ModTime() != dstfi.ModTime() {
		cSrc := C.CString(src)
		cDst := C.CString(dst)
		if uint32(C.copyf(cSrc, cDst)) != 0 {
			err = fmt.Errorf("copyf failed ('%s' -> '%s')", src, dst)
		}
		C.free(unsafe.Pointer(cSrc))
		C.free(unsafe.Pointer(cDst))
	}

	return
}
