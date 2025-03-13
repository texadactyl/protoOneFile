package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

func capture(pathData string) error {
	var recordCounter = int32(0)
	var recordPrefix RecordPrefix

	// Create or re-create the data file.
	dataFile, err := os.Create(pathData)
	if err != nil {
		fmt.Printf("capture: os.Create(%s) failed, err: %v\n", pathData, err)
		return err
	}
	defer dataFile.Close()

	// Begin frame.
	var rbfr PayloadBeginFrame
	fqn := "java/lang/String.getBytes()[B"
	recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeBeginFrame, len(recordPrefix.Rtype)))
	recordPrefix.Counter = recordCounter
	rbfr.FQNsize = int16(len(fqn))
	recordPrefix.PayloadSize = int32(rbfr.FQNsize + 2)
	copy(rbfr.FQNbytes[:len(fqn)], fqn)
	rbfr.FQNbytes = [len(rbfr.FQNbytes)]byte(stringToFixedBytes(fqn, len(rbfr.FQNbytes)))
	err = writeRecordToFile(dataFile, recordPrefix, rbfr)
	if err != nil {
		return err
	}

	var ri64chg PayloadI64Change
	recordPrefix.Counter = recordCounter
	recordPrefix.PayloadSize = int32(unsafe.Sizeof(ri64chg))
	ri64chg.ValueOld = int64(0)

	var rf64chg PayloadF64Change
	recordPrefix.Counter = recordCounter
	recordPrefix.PayloadSize = int32(unsafe.Sizeof(rf64chg))
	rf64chg.ValueOld = float64(0)

	iota := "Iota"   // int64 field name
	gamma := "Gamma" // float64 field name

	for recordCounter = 0; recordCounter < maxValueChanges; {

		// Write int64 change record to the data file.
		recordCounter++
		recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeI64Change, len(recordPrefix.Rtype)))
		recordPrefix.Counter = recordCounter
		ri64chg.FieldType = ftypeLocal
		ri64chg.ValueNew = int64(recordCounter)
		ri64chg.Index = int64(recordCounter)
		ri64chg.NameSize = int16(len(iota))
		ri64chg.NameBytes = [len(ri64chg.NameBytes)]byte(stringToFixedBytes(iota, len(ri64chg.NameBytes)))
		err = writeRecordToFile(dataFile, recordPrefix, ri64chg)
		if err != nil {
			return err
		}

		// Write float64 change record to the data file.
		recordCounter++
		recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeF64Change, len(recordPrefix.Rtype)))
		recordPrefix.Counter = recordCounter
		rf64chg.FieldType = ftypeLocal
		rf64chg.ValueNew = float64(recordCounter)
		rf64chg.Index = int64(recordCounter)
		rf64chg.NameSize = int16(len(gamma))
		rf64chg.NameBytes = [len(rf64chg.NameBytes)]byte(stringToFixedBytes(gamma, len(rf64chg.NameBytes)))
		err = writeRecordToFile(dataFile, recordPrefix, rf64chg)
		if err != nil {
			return err
		}

		// New values become the old values.
		ri64chg.ValueOld = ri64chg.ValueNew
		rf64chg.ValueOld = rf64chg.ValueNew
	}

	// Write end frame record.
	recordCounter++
	var refr PayloadEndFrame
	recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeEndFrame, len(recordPrefix.Rtype)))
	recordPrefix.Counter = recordCounter
	refr.FQNsize = rbfr.FQNsize
	refr.FQNbytes = rbfr.FQNbytes
	recordPrefix.PayloadSize = int32(refr.FQNsize + 2)
	err = writeRecordToFile(dataFile, recordPrefix, refr)
	if err != nil {
		return err
	}

	fmt.Printf("capture: Stored %d records\n", recordCounter+1)

	return nil
}

// Write a record to the data file.
func writeRecordToFile(dataFile *os.File, recordPrefix RecordPrefix, recordPayload any) error {

	// Write the record prefix.
	err := binary.Write(dataFile, binary.LittleEndian, recordPrefix)
	if err != nil {
		fmt.Printf("writeRecordToFile *** ERROR: binary.Write(recordPrefix) failed, Rtype=%s, Counter=%d, PayloadSize=%d, err: %v\n",
			string(recordPrefix.Rtype[:]), recordPrefix.Counter, recordPrefix.PayloadSize, err)
		return err
	}

	// Write the record payload.
	err = binary.Write(dataFile, binary.LittleEndian, recordPayload)
	if err != nil {
		fmt.Printf("writeRecordToFile *** ERROR: binary.Write(recordPayload) failed, Rtype=%s, Counter=%d, PayloadSize=%d, err: %v\n",
			string(recordPrefix.Rtype[:]), recordPrefix.Counter, recordPrefix.PayloadSize, err)
		return err
	}

	// Return success to caller.
	return nil
}

// Convert a string to a fixed-length byte array with space filled on the right.
func stringToFixedBytes(s string, size int) []byte {
	padded := fmt.Sprintf("%-*s", size, s) // Left-align and pad with spaces.
	return []byte(padded)[:size]           // Ensure it is exactly 'size' bytes.
}
