package main

import "testing"
import "sort"
import "reflect"
import "fmt"

func TestSort(t *testing.T) {
	versions := []string {
		"release/1.2.4",
		"release/1.3.4",
		"release/1.2.3",
		"release/1.29.0",
		"release/1.28.0"}

	sort.Sort(BySemverNumber(versions))

	correctVersions := []string{
		"release/1.2.3",
		"release/1.2.4",
		"release/1.3.4",
		"release/1.28.0",
		"release/1.29.0"}

	if ! reflect.DeepEqual(versions, correctVersions) {
		t.Error(fmt.Printf("Expected %v, got %v", correctVersions, versions))
	}
}

func TestParse(t *testing.T) {
	// The below is a truncated output from cloc. The SUM won't add up. We don't care, it's not our job to verify that
	// this adds up. We only need to find the last column entries for each language and the total LOC count.
	clocOutput := []byte(`9641 text files.
    6978 unique files.
    6416 files ignored.

http://cloc.sourceforge.net v 1.60  T=45.69 s (105.9 files/s, 18208.7 lines/s)
--------------------------------------------------------------------------------
Language                      files          blank        comment           code
--------------------------------------------------------------------------------
Java                           1269          34118          13669         141385
Javascript                      163          18733          27769         100695
Scala                           355           9107          10880          32138
--------------------------------------------------------------------------------
SUM:                           4836         163326          63095         605448
--------------------------------------------------------------------------------`)

	parsed, sum := parseClocOutput(clocOutput)
	expected := map[string]int64 {
		"Java": 141385,
		"Javascript": 100695,
		"Scala": 32138,
	}

	if sum != 605448 {
		t.Errorf("Found sum %v but expected %v", sum, 605448)
	}
	if ! reflect.DeepEqual(parsed, expected) {
		t.Errorf("Expected %v, got %v", expected, parsed)
	}
}
