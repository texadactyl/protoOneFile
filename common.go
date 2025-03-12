package main

import (
	"fmt"
	"os"
)

// Data and its associated index file.
const pathData = "./saucisse.data"

// Some maximums.
const maxValueChanges = 1000 //
const maxDisplaySamples = 20

// Record anatomy:
// * Prefix - common across record types
// * Payload - specific to record type

// Record types.
const rtypeBeginFrame = "FRMBEG"
const rtypeEndFrame = "FRMEND"
const rtypeI64Change = "CHGI64"
const rtypeF64Change = "CHGF64"
const rtypeStatics = "STATICS"
const rtypeOpCode = "OPCODE"
const rtypePush = "PUSH"
const rtypePop = "POP"
const rtypeGfuncCall = "GCALL"
const rtypeGfuncReturn = "GRETURN"

// Field types.
const ftypeLocal = 1    // Non-static variable within a function
const ftypeInstance = 2 // Non-static variable, global to the instance of an object
const ftypeStatic = 3   // Static variables

// Data record definitions.

type RecordPrefix struct {
	Rtype       [8]byte // rtype*
	Counter     int32
	PayloadSize int32
}

type RecordBeginFrame struct {
	FQNsize  int16
	FQNbytes [100]byte
}

type RecordEndFrame struct {
	FQNsize  int16
	FQNbytes [100]byte
}

type RecordI64Change struct {
	FieldType int16 // ftype*
	ValueOld  int64
	ValueNew  int64
}

type RecordF64Change struct {
	FieldType int16 // ftype*
	ValueOld  float64
	ValueNew  float64
}

type RecordStatics struct {
	TblSize  int32
	TblBytes [0]byte // Serialized statics table
}

type RecordOpCode struct {
	PC     int32
	OpCode int32
}

type RecordPush struct {
	PushSize  int32   // size of PushBytes
	PushBytes [0]byte // String-ified push operand, converted to [size]byte
}

type RecordPop struct {
	PopSize  int32  // size of PopBytes
	PopBytes []byte // String-ified push operand, converted to [size]byte
}

type RecordGfuncCall struct {
	// TODO
}

type RecordGfuncReturn struct {
	// TODO
}

// fileSize gets the current size of the file using its open file handle.
func fileSize(file *os.File) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return -1, err
	}
	return info.Size(), nil
}

// Convert a string to a fixed-length byte array with space filled on the right.
func stringToFixedBytes(s string, size int) []byte {
	padded := fmt.Sprintf("%-*s", size, s) // Left-align and pad with spaces.
	return []byte(padded)[:size]           // Ensure it is exactly 'size' bytes.
}
