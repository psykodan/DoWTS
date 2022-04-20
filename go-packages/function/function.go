package function

//Package for generic serverelss function. Takes ID (0,1,2,...,n), memory allocation in MB and its execution time in ms
//Platform packages are then used to calculate price
type Function struct {
	ID      int
	Memory  uint64
	Runtime uint64
	Trigger string
}
