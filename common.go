package main

/***
	Record anatomy: [RecordPrefix][PayloadX]
	where:

	RecordPrefix - common to all record types
	PayloadX - specific to record type X
 ***/

// Data and its associated index file.
const pathData = "./saucisse.data"

// Maximum array lengths.
const maxFqnLength = 100
const maxNameLength = 100
const maxStaticsLength = 65535
const maxPushPopArg = 100

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
const ftypeStatic = 3   // Static variable, global to all functions of every object

//  -------------------------- Begin data record definitions

type RecordPrefix struct {
	Rtype       [8]byte // rtype*
	Counter     int32
	PayloadSize int32
}

type PayloadBeginFrame struct {
	FQNsize  int16
	FQNbytes [maxFqnLength]byte
}

type PayloadEndFrame struct {
	FQNsize  int16
	FQNbytes [maxFqnLength]byte
}

type PayloadI64Change struct {
	FieldType int16 // ftype*
	ValueOld  int64
	ValueNew  int64
	Index     int64
	NameSize  int16
	NameBytes [maxNameLength]byte
}

type PayloadF64Change struct {
	FieldType int16 // ftype*
	ValueOld  float64
	ValueNew  float64
	Index     int64
	NameSize  int16
	NameBytes [maxNameLength]byte
}

type PayloadStatics struct {
	TblSize  int32
	TblBytes [maxStaticsLength]byte // Serialized statics table
}

type PayloadOpCode struct {
	PC     int32
	OpCode int32
}

type PayloadPush struct {
	PushSize  int16               // size of PushBytes
	PushBytes [maxPushPopArg]byte // String-ified push operand, converted to [size]byte
}

type PayloadPop struct {
	PopSize  int16               // size of PopBytes
	PopBytes [maxPushPopArg]byte // String-ified push operand, converted to [size]byte
}

type PayloadGfuncCall struct {
	// TODO
}

type PayloadGfuncReturn struct {
	// TODO
}

// -------------------------- End of data record definitions
