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

	// Create or open the data file (where the records will be stored)
	dataFile, err := os.Create(pathData)
	if err != nil {
		fmt.Println("capture: Error creating data file:", err)
		return err
	}
	defer dataFile.Close()

	// Begin frame.
	var rbfr RecordBeginFrame
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

	var ri64chg RecordI64Change
	recordPrefix.Counter = recordCounter
	recordPrefix.PayloadSize = int32(unsafe.Sizeof(ri64chg))
	ri64chg.ValueOld = int64(0)

	var rf64chg RecordF64Change
	recordPrefix.Counter = recordCounter
	recordPrefix.PayloadSize = int32(unsafe.Sizeof(rf64chg))
	rf64chg.ValueOld = float64(0)

	for recordCounter = 0; recordCounter < maxValueChanges; {

		// Write int64 change record to the data file.
		recordCounter++
		recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeI64Change, len(recordPrefix.Rtype)))
		recordPrefix.Counter = recordCounter
		ri64chg.FieldType = ftypeLocal
		ri64chg.ValueNew = int64(recordCounter)
		err := writeRecordToFile(dataFile, recordPrefix, ri64chg)
		if err != nil {
			return err
		}

		// Write float64 change record to the data file.
		recordCounter++
		recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeF64Change, len(recordPrefix.Rtype)))
		recordPrefix.Counter = recordCounter
		rf64chg.FieldType = ftypeLocal
		rf64chg.ValueNew = float64(recordCounter)
		err = writeRecordToFile(dataFile, recordPrefix, rf64chg)
		if err != nil {
			return err
		}

		// Update old values.
		ri64chg.ValueOld = ri64chg.ValueNew
		rf64chg.ValueOld = rf64chg.ValueNew
	}

	// Write end frame record.
	recordCounter++
	var refr RecordEndFrame
	recordPrefix.Rtype = [len(recordPrefix.Rtype)]byte(stringToFixedBytes(rtypeEndFrame, len(recordPrefix.Rtype)))
	recordPrefix.Counter = recordCounter
	refr.FQNsize = rbfr.FQNsize
	refr.FQNbytes = rbfr.FQNbytes
	recordPrefix.PayloadSize = int32(refr.FQNsize + 2)
	err = writeRecordToFile(dataFile, recordPrefix, refr)
	if err != nil {
		return err
	}

	fmt.Printf("capture: Finished writing record number %d\n", recordCounter)

	return nil
}

// Write a record to the data file.
func writeRecordToFile(dataFile *os.File, recordPrefix RecordPrefix, recordPayload any) error {

	// Write record prefix to data file
	err := binary.Write(dataFile, binary.LittleEndian, recordPrefix)
	if err != nil {
		fmt.Printf("writeRecordToFile *** ERROR: binary.Write(recordPrefix) failed, Rtype=%s, Counter=%d, PayloadSize=%d, err: %v\n",
			string(recordPrefix.Rtype[:]), recordPrefix.Counter, recordPrefix.PayloadSize, err)
		return err
	}

	// Write record payload to data file
	err = binary.Write(dataFile, binary.LittleEndian, recordPayload)
	if err != nil {
		fmt.Printf("writeRecordToFile *** ERROR: binary.Write(recordPayload) failed, Rtype=%s, Counter=%d, PayloadSize=%d, err: %v\n",
			string(recordPrefix.Rtype[:]), recordPrefix.Counter, recordPrefix.PayloadSize, err)
		return err
	}

	// Return success to caller.
	return nil
}
