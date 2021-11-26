package main

/*
#include <stdio.h>
#include <stdlib.h>

int copyf(const char *src, const char *dst)
{
	int ret = EXIT_FAILURE;
	
	FILE *srcf, *dstf;
	if ((!(srcf = fopen(src, "r"))) ||
		(!(dstf = fopen(dst, "w"))))
		goto ABORT;

	char buf[BUFSIZ];
	int n;
	do {
		n = fread(buf, sizeof(char), BUFSIZ, srcf);
		if (ferror(srcf)) perror("fread failure");
		else {
			fwrite(buf, sizeof(char), n, dstf);
			if (ferror(dstf)) perror("fwrite failure");
		}
	} while (!feof(srcf) && !ferror(srcf) && !ferror(dstf));
	ret = EXIT_SUCCESS;
	
ABORT:
	if (srcf) fclose(srcf);
	if (dstf) fclose(dstf);
	return ret;
}
*/
import "C"
import (
	"fmt"
	"path/filepath"
	"os"
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

	cSrc := C.CString(src)
	cDst := C.CString(dst)
	if uint32(C.copyf(cSrc, cDst)) != 0 {
		err = fmt.Errorf("copyf failed ('%s' -> '%s')", src, dst)
	}
	C.free(unsafe.Pointer(cSrc))
	C.free(unsafe.Pointer(cDst))

	return
}
