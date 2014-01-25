package main

// #cgo LDFLAGS: -limobiledevice
// #include <libimobiledevice/libimobiledevice.h>
// #include <libimobiledevice/lockdown.h>
// #include <libimobiledevice/file_relay.h>
/*
#include <stdlib.h>
#include <stdio.h>
#include <strings.h>

typedef uint8_t byte;
typedef unsigned long uintgo;
typedef long _intgo;  // there is a name collision, but it doesn't seem to be defined until after my code.

struct String
{
        byte*   str;
        _intgo    len;
};

struct  Slice
{                               // must not move anything
        byte*   array;          // actual data
        uintgo  len;            // number of elements
        uintgo  cap;            // allocated number of elements
};


// Make a copy of a []string as a null terminated list of cstrings. Caller must C.free the result.
const char** StringArray(void* goStringArray) {
    struct Slice *values = goStringArray;
	int len = values->len;
	struct String* strings = (struct String*) values->array;

	int total = 0;
	for (int i=0;i<len;i++) {
		total += strings[i].len + 1;
	}
    
    // We allocate everything as one chunk, with internal pointers to make free()ing easier.
	void* rval = malloc(sizeof(void*)*(len+1)+total);
	char* payload = rval + sizeof(void*)*(len+1);
	void** ptrs = rval;
	
	for (int i=0;i<len;i++) {
		*(ptrs++) = payload;
		bcopy(strings[i].str, payload, strings[i].len);
		payload += strings[i].len;
		*(payload++) = 0;
	}
	*(ptrs++) = 0;
	
	return rval;
}
*/
import "C"

import (
	"unsafe"
	"os"
	"fmt"
	"log"
)

func must(err error) {
	log.Fatal(err)
}

func main() {
	
	if len(os.Args) < 3 {
		fmt.Println("Usage:", os.Args[0], "output.cpio.gz domains...")
		fmt.Println("Known domains: AppleSupport Network VPN WiFi UserDatabases CrashReporter tmp SystemConfiguration")
		return
	}
	
	output := os.Args[1]
	domains := os.Args[2:]
	
	// All of this stuff should be wrapped in a go library at some point.
	
	var dev C.idevice_t
	if result := C.idevice_new(&dev, nil); result != C.LOCKDOWN_E_SUCCESS {
		return
	}
	defer C.idevice_free(dev)
	var client C.lockdownd_client_t
	if r := C.lockdownd_client_new_with_handshake(dev, &client, nil); r != C.LOCKDOWN_E_SUCCESS {
		fmt.Println("LC fail")
		return
	}
	
 	var port C.lockdownd_service_descriptor_t
	if r := C.lockdownd_start_service(client, C.CString("com.apple.mobile.file_relay"), &port); r != C.LOCKDOWN_E_SUCCESS {
		fmt.Println("can't start file_relay")
		C.lockdownd_client_free(client)
		return
	}
	C.lockdownd_client_free(client)
	var frc C.file_relay_client_t
	if r := C.file_relay_client_new(dev, port, &frc); r != C.FILE_RELAY_E_SUCCESS {
		fmt.Println("could not start client")
	}
	defer C.file_relay_client_free(frc)

	sources := C.StringArray(unsafe.Pointer(&domains))
	defer C.free(unsafe.Pointer(sources))
	var dump C.idevice_connection_t
	if r := C.file_relay_request_sources(frc, sources, &dump); r != C.FILE_RELAY_E_SUCCESS {
		fmt.Println("could not get sources")
		return
	}
	
	w, err := os.Create(output)
	if err != nil {
		fmt.Println("Failed to create file", err)
		return
	}
	defer w.Close()
	
	buf := make([]byte, 4096)
	
	var total int
	var length C.uint32_t
	for {
		if r := C.idevice_connection_receive(dump, (*C.char)(unsafe.Pointer(&buf[0])), 4096, &length); r != C.IDEVICE_E_SUCCESS {
			break
		}
		w.Write(buf[:length])
		total += int(length)
	}
	fmt.Println("read", total, "bytes")
	fmt.Println(dev, sources)

}
	