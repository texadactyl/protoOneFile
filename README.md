# protoOneFile
Create a single physical-sequential hand-off file in function `capture` then execute function `analysis` to:
* Load all file records into a B-tree.
* For several randomly selected B-tree items, print the contents of the leaf (record prefix + payload).
* Print the first (begin frame) and last (end frame) items of the B-tree.

To-date, supporting only record types begin frame, end frame, change an int64 primitive, and change a float64 primitive.

See `common.go` for the record specifications and TODOs.
