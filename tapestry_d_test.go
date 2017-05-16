package main

import (
	"fmt"
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

func validInterfaceName(n string) bool {
	if n == "lo0" || n == "en0" || n == "en1" {
		return true
	}
	return false
}

func TestReversePids(t *testing.T) {
	example := []int32{1, 2, 3, 4, 5}
	result := []int32{5, 4, 3, 2, 1}
	reversePids(example)

	if !equal(result, example) {
		t.Errorf("reversePids not working!\n")
	}
}

func TestGetInterfaces(t *testing.T) {
	interfaces := getInterfaces()

	for _, intf := range interfaces {
		name := intf.Name
		fmt.Printf("Interface: %s\n", name)
		if !validInterfaceName(name) {
			t.Errorf("getInterfaces() testing failed\n")
		}
	}
}
