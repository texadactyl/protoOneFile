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

// Maximums for controlling analysis behaviour.
const maxValueChanges = 1000 //
const maxDisplaySamples = 20

type BtreeLeaf struct {
	Key     int32 // Record number
	Prefix  RecordPrefix
	Payload any
}

// Implement the `Less` function for btree.Item interface.
func (this BtreeLeaf) Less(that btree.Item) bool {
	return this.Key < that.(*BtreeLeaf).Key
}
func analysis(pathData string) error {

	eofFlag := false
	tracing := false
	var recordNumber int32
	var recordPrefix RecordPrefix
	var rbfr PayloadBeginFrame
	var ri64chg PayloadI64Change
	var rf64chg PayloadF64Change
	var refr PayloadEndFrame

	// Open the data file.
	dataFile, err := os.Open(pathData)
	if err != nil {
		fmt.Println("analysis *** ERROR: os.Open(%s) failed, err: %v", pathData, err)
		return err
	}
	defer dataFile.Close()

	// Initialise the B-tree.
	bTree := btree.New(2)

	// Read every record from the data file.
	for recordNumber = 0; ; recordNumber++ {

		// Read record prefix.
		err = binary.Read(dataFile, binary.LittleEndian, &recordPrefix)
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			// no more records left
			eofFlag = true
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

		// Retrieved the record prefix.
		// Read the payload.
		// Finally, add both prefix and payload to the B-tree as an BtreeLeaf.
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
			bTree.ReplaceOrInsert(&BtreeLeaf{Key: recordNumber, Prefix: recordPrefix, Payload: ri64chg})
		case rtypeF64Change:
			err = binary.Read(dataFile, binary.LittleEndian, &rf64chg)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeF64Change) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&BtreeLeaf{Key: recordNumber, Prefix: recordPrefix, Payload: rf64chg})
		case rtypeBeginFrame:
			err = binary.Read(dataFile, binary.LittleEndian, &rbfr)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeBeginFrame) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&BtreeLeaf{Key: recordNumber, Prefix: recordPrefix, Payload: rbfr})
		case rtypeEndFrame:
			err = binary.Read(dataFile, binary.LittleEndian, &refr)
			if err != nil {
				fmt.Printf("analysis *** ERROR: binary.Read(rtypeEndFrame) failed, recordNumber=%d, err: %v\n", recordNumber, err)
				return err
			}
			bTree.ReplaceOrInsert(&BtreeLeaf{Key: recordNumber, Prefix: recordPrefix, Payload: refr})
		default:
			fmt.Printf("analysis *** ERROR: binary.Read ==> unknown record type: %s, recordNumber=%d\n", rtype, recordNumber)
			return err
		}

	}

	fmt.Printf("analysis: Loaded %d records into the B-tree\n", recordNumber)

	// Retrieve random records and report their contents.
	for ix := 0; ix < maxDisplaySamples; ix++ {
		err = reportData(randomPal(), bTree)
		if err != nil {
			return err
		}
	}

	// Report the first record and the last record.
	reportData(int32(0), bTree)
	reportData(maxValueChanges+1, bTree)

	return nil
}

// Given a key, report the associated data record.
func reportData(recordNumber int32, loadedTree *btree.BTree) error {
	var fname string

	// Get B-tree leaf and set indexRecord to its components.
	item := loadedTree.Get(&BtreeLeaf{Key: recordNumber})
	if item == nil {
		errMsg := fmt.Sprintf("reportData *** ERROR: Cannot find index recordNumber: %d", recordNumber)
		println(errMsg)
		return errors.New(errMsg)
	}
	indexRecord := item.(*BtreeLeaf)

	// Get record type.
	rtype := strings.TrimSpace(string(indexRecord.Prefix.Rtype[:]))

	// Show data.
	switch rtype {
	case rtypeBeginFrame:
		rbfr := indexRecord.Payload.(PayloadBeginFrame)
		fqn := string(rbfr.FQNbytes[:rbfr.FQNsize])
		fmt.Printf("reportData: begin frame: Record %d FQN = %s\n", indexRecord.Key, fqn)
	case rtypeI64Change:
		ri64chg := indexRecord.Payload.(PayloadI64Change)
		if ri64chg.NameSize == 0 {
			fname = ""
		} else {
			fname = strings.TrimSpace(string(ri64chg.NameBytes[:ri64chg.NameSize]))
		}
		fmt.Printf("reportData: int64 change: Record %d, fname=\"%s\", ftype=%d, old = %d, new = %d\n",
			indexRecord.Key, fname, ri64chg.FieldType, ri64chg.ValueOld, ri64chg.ValueNew)
	case rtypeF64Change:
		rf64chg := indexRecord.Payload.(PayloadF64Change)
		if rf64chg.NameSize == 0 {
			fname = ""
		} else {
			fname = strings.TrimSpace(string(rf64chg.NameBytes[:rf64chg.NameSize]))
		}
		fmt.Printf("reportData: float64 change: Record %d, fname=\"%s\", ftype=%d, old = %f, new = %f\n",
			indexRecord.Key, fname, rf64chg.FieldType, rf64chg.ValueOld, rf64chg.ValueNew)
	case rtypeEndFrame:
		refr := indexRecord.Payload.(PayloadEndFrame)
		fqn := string(refr.FQNbytes[:refr.FQNsize])
		fmt.Printf("reportData: end frame: Record %d FQN = %s\n", indexRecord.Key, fqn)
	default:
		fmt.Printf("reportData *** ERROR: unrecognizable record type: Record %d, record type = %s\n", indexRecord.Key, rtype)
	}

	return nil
}

// We don't want a random number in the range of [0, n). We actually want the range = (0, n].
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
