package calendar

import (
	"fmt"
	"testing"
	"time"
)

func Test_randSequence(t *testing.T) {
	s1 := randSequence(10)
	s2 := randSequence(10)

	fmt.Println(s1)
	fmt.Println(s2)
}

func Test_randSequenceV2(t *testing.T) {
	seed := time.Now().UnixNano()
	s1 := randSequenceV2(10, seed)
	s2 := randSequenceV2(10, seed)

	fmt.Println(s1)
	fmt.Println(s2)
}
