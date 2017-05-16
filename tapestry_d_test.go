package main

import (
	"testing"
)

func equal(a []int32, b []int32) bool {
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestReversePids(t *testing.T) {
	example := []int32{1, 2, 3, 4, 5}
	result := []int32{5, 4, 3, 2, 1}
	reversePids(example)

	if !equal(result, example) {
		t.Errorf("reversePids not working!\n")
	}
}
