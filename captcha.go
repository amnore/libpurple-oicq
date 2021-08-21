package main

/*
   #cgo pkg-config: purple glib-2.0 gtk+-3.0 webkit2gtk-4.0
   #include <glib.h>
   #include <glib/gi18n.h>
   #include <purple.h>
   #include <stdbool.h>
   #include <gtk/gtk.h>
   #include <webkit2/webkit2.h>

   void get_slider_ticket(PurpleAccount *account, const char *url);
   void wait_for_qrcode(PurpleAccount *account, const char *url);
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

var myChan chan interface{} = make(chan interface{})

func getSliderTicket(account *C.PurpleAccount, url string) (string, error) {
	glib.IdleAdd(func() {
		curl := C.CString(url)
		defer C.free(unsafe.Pointer(curl))

		C.get_slider_ticket(account, curl)
	})

	ticket := (<-myChan).(*string)
	if ticket == nil {
		return "", fmt.Errorf("Captcha cancelled")
	}
	return *ticket, nil
}

func waitForQRCode(account *C.PurpleAccount, url string) error {
	glib.IdleAdd(func() {
		curl := C.CString(strings.Replace(url, "safe/verify", "safe/qrcode", 1))
		defer C.free(unsafe.Pointer(curl))

		C.wait_for_qrcode(account, curl)
	})

	ok := (<-myChan).(bool)
	if ok {
		return nil
	} else {
		return fmt.Errorf("Captcha cancelled")
	}
}

//export sendString
func sendString(ticket *C.char) {
	if ticket == nil {
		myChan <- nil
	} else {
		str := C.GoString(ticket)
		myChan <- &str
	}
}

//export sendBool
func sendBool(value bool) {
	myChan <- value
}
