/* Code generated by cmd/cgo; DO NOT EDIT. */

/* package command-line-arguments */


#line 1 "cgo-builtin-export-prolog"

#include <stddef.h>

#ifndef GO_CGO_EXPORT_PROLOGUE_H
#define GO_CGO_EXPORT_PROLOGUE_H

#ifndef GO_CGO_GOSTRING_TYPEDEF
typedef struct { const char *p; ptrdiff_t n; } _GoString_;
#endif

#endif

/* Start of preamble from import "C" comments.  */


#line 7 "bridge.go"

#include <stddef.h>
#include <stdint.h>
typedef struct GoDataChannelMessage
{
    // bool in cgo is really weird, so it's simpler to just use classic int as bool
    int is_string;
    void *data;
    size_t data_len;
} GoDataChannelMessage;

typedef uint16_t GoDataChannel;
typedef void (*GoOnDataChannelFunc)(GoDataChannel);
typedef void (*GoOnOpenFunc)(GoDataChannel);
typedef void (*GoOnMessageFunc)(GoDataChannel, GoDataChannelMessage);

// Calling C function pointers is currently not supported, however you can
// declare Go variables which hold C function pointers and pass them back
// and forth between Go and C. C code may call function pointers received from Go.
// Reference: https://golang.org/cmd/cgo/#hdr-Go_references_to_C
inline void bridge_on_data_channel(GoOnDataChannelFunc cb, GoDataChannel d)
{
    cb(d);
}

inline void bridge_on_open(GoOnOpenFunc cb, GoDataChannel d)
{
	cb(d);
}

inline void bridge_on_message(GoOnMessageFunc cb, GoDataChannel d, GoDataChannelMessage msg)
{
	cb(d, msg);
}

#line 1 "cgo-generated-wrapper"


/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef size_t GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
#ifdef _MSC_VER
#include <complex.h>
typedef _Fcomplex GoComplex64;
typedef _Dcomplex GoComplex128;
#else
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;
#endif

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

#ifndef GO_CGO_GOSTRING_TYPEDEF
typedef _GoString_ GoString;
#endif
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


// nolint
//
extern void GoRun(GoOnDataChannelFunc f);

// nolint
//
extern void GoOnOpen(GoDataChannel d, GoOnOpenFunc f);

// nolint
//
extern void GoOnMessage(GoDataChannel d, GoOnMessageFunc f);

// nolint
//
extern void GoSendText(GoDataChannel d, char* t);

// nolint
//
extern char* GoLabel(GoDataChannel d);

#ifdef __cplusplus
}
#endif
