package jas

type stack []int

func (s *stack) Push(i int) *stack {
	if s == nil {
		s = &stack{i}
		return s
	}
	*s = append(*s, i)
	return s
}
func (s *stack) Pop() int {
	r := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return r
}
