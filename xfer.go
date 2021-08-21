package main

/*
   #cgo pkg-config: purple glib-2.0
   #include <purple.h>
   #include <stddef.h>

   PurpleXfer *new_xfer_receive(PurpleAccount *account, const char *who, const char *url, const char *filename, size_t size);
*/
import "C"

func newXferReceive(
	account *C.PurpleAccount,
	who *C.char,
	url string,
	filename string,
	size int64,
) *C.PurpleXfer {
}
