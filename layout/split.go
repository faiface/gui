package layout

// SplitFunc represents a way to split a space among a number of elements.
// The space of the returned slice must be equal to the number of elements.
// The sum of all elements of the returned slice must be eqal to the space.
type SplitFunc func(elements int, space int) []int

// EvenSplit implements SplitFunc to split a space (almost) evenly among the elements.
// It is almost evenly because width may not be divisible by elements.
func EvenSplit(elements int, width int) []int {
	ret := make([]int, 0, elements)
	for elements > 0 {
		v := width / elements
		width -= v
		elements -= 1
		ret = append(ret, v)
	}
	return ret
}
