package jas

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
)

var bigData []byte
var trashObj []byte
var err error

func init() {
	bigData, err = ioutil.ReadFile("testdata/bigdata.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	trashObj, err = ioutil.ReadFile("testdata/trash.json")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func TestIsType(t *testing.T) {
	//types := []AtomType{AtomTypeNumber, AtomTypeBoolean, AtomTypeArray, AtomTypeMap, AtomTypeString}
	c := []struct {
		c byte
		t AtomType
	}{
		{'t', AtomTypeBoolean},
		{'f', AtomTypeBoolean},
		{'0', AtomTypeNumber},
		{'1', AtomTypeNumber | AtomTypeArray},
		{'2', AtomTypeMap | AtomTypeArray | AtomTypeBoolean | AtomTypeNumber | AtomTypeString},
		{'9', AtomTypeNumber},
		{'-', AtomTypeNumber},
		{'[', AtomTypeArray},
		{'{', AtomTypeMap},
		{'"', AtomTypeString},
		{'z', AtomTypeUndefined},
	}
	for _, o := range c {
		if !isType(o.c, o.t) {
			t.Fail()
		}
	}

}

func TestVideos(t *testing.T) {
	rf, _ := os.Open("./testdata/test.json")
	f, _ := ioutil.ReadAll(rf)
	mainAtom := NewAtom(f)

	mainAtom.Next("videoList", SelectArray)
	wg := sync.WaitGroup{}
	for mainAtom.Next("video", SelectMap) != nil {
		// create new atom video block (take slice bytes)
		videoAtom := mainAtom.Take()
		wg.Add(1)
		go func() {
			defer wg.Done()

			// find first
			if videoAtom.Next("name", SelectMap).Next("text", SelectString) == nil {
				fmt.Println("video name not found")
				return
			}
			// take and parsing object current cursor
			name := videoAtom.ToString()
			// move cursor next object
			if videoAtom.Next("duration", SelectMap).Next("seconds", SelectNumber) == nil {
				return
			}
			// take atom and convert to float
			sec := videoAtom.ToFloat()
			// variant two check empty string
			author := videoAtom.End(). // move cursor to end object
							Prev("author", SelectMap). // used reverse algorithm (find last string)
							Next("name", SelectMap).
							Next("text", SelectString).
							ToString()
			// check text
			if author == "" {
				return
			}
			fmt.Printf("Name: %s\nDuration: %d\nAuthor: %s\n", name, int(sec), author)
		}()
	}
	wg.Wait()
	fmt.Printf("Header: %s\nTrakingId: %s\n",
		mainAtom.Start().Next("header", SelectMap).Next("text", SelectString).ToString(),
		mainAtom.End().Prev("clickTracking", SelectMap).Next("id", SelectString).ToString())

	j := NewAtom(bigData)
	fmt.Println(j.End().Prev("header").Next("musicImmersiveHeaderRenderer").Next("title").Next("text").ToString())

	NewAtom(bigData).End().Prev("musicTwoRowItemRenderer").Next("url")
}
func BenchmarkAtom_Next(b *testing.B) {
	j := NewAtom(bigData)
	for i := 0; i < b.N; i++ {
		j.End().Prev("musicTwoRowItemRenderer").Next("url")

	}

}
