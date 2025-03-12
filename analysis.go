package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/btree"
	"io"
	"math/rand"
	"os"
	"strings"
)

func analysis(pathData string) error {

	tracing := false
	var recordNumber int32
	var recordPrefix RecordPrefix
	var rbfr RecordBeginFrame
	var ri64chg RecordI64Change
	var rf64chg RecordF64Change
	var refr RecordEndFrame

	// Open the data file.
	dataFile, err := os.Open(pathData)
	if err != nil {
		fmt.Println("analysis *** ERROR: os.Open(%s) failed, err: %v", pathData, err)
		return err
	}
	defer dataFile.Close()

	// Initialise a B-tree.
	bTree := btree.New(2)

	eofFlag := false
	for recordNumber = 0; ; recordNumber++ {

		// Read record prefix from data file.
		err = binary.Read(dataFile, binary.LittleEndian, &recordPrefix)
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			eofFlag = true // no more records left
		case errors.Is(err, io.ErrUnexpectedEOF):
			fmt.Printf("analysis *** ERROR: binary.Read(recordPrefix) encounted an unexpected EOF, recordNumber=%d, err: %v\n",
				recordNumber, err)
			return err
		default:
			fmt.Printf("analysis *** ERROR: binary.Read(recordPrefix) failed, recordNumber=%d, err: %v\n",
				recordNumber, err)
			return err
		}

		// If EOF break out of this loop.
		if eofFlag {
			break
		}

		// Got the record prefix. Now, read the payload, depending on the record type.
		rtype := strings.TrimSpace(string(recordPrefix.Rtype[:]))
		if tracing {
			fmt.Printf("analysis tracing: Read recordNumber=%d, prefix Rtype=%s, Counter=%d, PayloadSize=%d\n",
				recordNumber, recordPrefix.Rtype, recordPrefix.Counter, recordPrefix.PayloadSize)
		}
		switch rtype {
		case rtypeI64Change:
			err = binary.Read(dataFile, binary.LittleEndian, &ri64chg)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeI64Change) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&IndexRecord{Key: recordNumber, Prefix: recordPrefix, Payload: ri64chg})
		case rtypeF64Change:
			err = binary.Read(dataFile, binary.LittleEndian, &rf64chg)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeF64Change) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&IndexRecord{Key: recordNumber, Prefix: recordPrefix, Payload: rf64chg})
		case rtypeBeginFrame:
			err = binary.Read(dataFile, binary.LittleEndian, &rbfr)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeBeginFrame) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&IndexRecord{Key: recordNumber, Prefix: recordPrefix, Payload: rbfr})
		case rtypeEndFrame:
			err = binary.Read(dataFile, binary.LittleEndian, &refr)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeEndFrame) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&IndexRecord{Key: recordNumber, Prefix: recordPrefix, Payload: refr})
		default:
			fmt.Printf("analysis *** ERROR: binary.Read ==> unknown record type: %s, recordNumber=%d\n", rtype, recordNumber)
			return err
		}

	}

	fmt.Printf("analysis: Loaded %d records into the B-tree\n", recordNumber)

	// Use the loaded tree to retrieve data records.
	for ix := 0; ix < maxDisplaySamples; ix++ {
		err = reportData(randomPal(), bTree)
		if err != nil {
			return err
		}
	}

	reportData(int32(0), bTree)
	reportData(maxValueChanges+1, bTree)

	return nil
}

// Given a key, report the associated data record.
func reportData(recordNumber int32, loadedTree *btree.BTree) error {
	item := loadedTree.Get(&IndexRecord{Key: recordNumber})
	if item == nil {
		errMsg := fmt.Sprintf("reportData *** ERROR: Cannot find index recordNumber: %d", recordNumber)
		println(errMsg)
		return errors.New(errMsg)
	}
	indexRecord := item.(*IndexRecord)

	// Show data.
	rtype := strings.TrimSpace(string(indexRecord.Prefix.Rtype[:]))
	switch rtype {
	case rtypeBeginFrame:
		rbfr := indexRecord.Payload.(RecordBeginFrame)
		fqn := string(rbfr.FQNbytes[:rbfr.FQNsize])
		fmt.Printf("reportData: begin frame: Record %d FQN = %s\n", indexRecord.Key, fqn)
	case rtypeI64Change:
		ri64chg := indexRecord.Payload.(RecordI64Change)
		fmt.Printf("reportData: int64 change: Record %d, old = %d, new = %d\n", indexRecord.Key, ri64chg.ValueOld, ri64chg.ValueNew)
	case rtypeF64Change:
		rf64chg := indexRecord.Payload.(RecordF64Change)
		fmt.Printf("reportData: float64 change: Record %d, old = %f, new = %f\n", indexRecord.Key, rf64chg.ValueOld, rf64chg.ValueNew)
	case rtypeEndFrame:
		refr := indexRecord.Payload.(RecordEndFrame)
		fqn := string(refr.FQNbytes[:refr.FQNsize])
		fmt.Printf("reportData: end frame: Record %d FQN = %s\n", indexRecord.Key, fqn)
	default:
		fmt.Printf("reportData *** ERROR: unrecognizable record type: Record %d, record type = %s\n", indexRecord.Key, rtype)
	}

	return nil
}

// We don't want a random number generator range of [0, n). We actually want (0, n].
func randomPal() int32 {
	var rn int32
	for {
		rn = rand.Int31n(maxValueChanges + 1)
		if rn > 0 {
			break
		}
	}
	return rn
}

// IndexRecord is the wrapper for storing data in the B-bTree
type IndexRecord struct {
	Key     int32 // Record number in the data file
	Prefix  RecordPrefix
	Payload any
}

// Implement the `Less` function for btree.Item interface.
func (this IndexRecord) Less(that btree.Item) bool {
	return this.Key < that.(*IndexRecord).Key
}
