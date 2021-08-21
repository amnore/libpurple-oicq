package main

/*
   #cgo pkg-config: purple
   #define PURPLE_PLUGINS
   #include <debug.h>
   #include <stdio.h>

   static void oicq_debug(PurpleDebugLevel level, _GoString_ str) {
     size_t len = _GoStringLen(str);
     const char *chars = _GoStringPtr(str);

     char format[10];
     sprintf(format, "%%.%zus", len);
     purple_debug(level, "oicq", format, chars);
   }
*/
import "C"

import "fmt"

func debug(level C.PurpleDebugLevel, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	C.oicq_debug(level, str)
}

func debugMisc(format string, args ...interface{}) {
	debug(C.PURPLE_DEBUG_MISC, format, args...)
}

func debugInfo(format string, args ...interface{}) {
	debug(C.PURPLE_DEBUG_INFO, format, args...)
}

func debugWarning(format string, args ...interface{}) {
	debug(C.PURPLE_DEBUG_WARNING, format, args...)
}

func debugError(format string, args ...interface{}) {
	debug(C.PURPLE_DEBUG_ERROR, format, args...)
}

func debugFatal(format string, args ...interface{}) {
	debug(C.PURPLE_DEBUG_FATAL, format, args...)
}
