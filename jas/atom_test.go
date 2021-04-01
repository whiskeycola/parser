package jas

import (
	"encoding/json"
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
	//types := []AtomType{Number, Boolean, Array, Map, String}
	c := []struct {
		c byte
		t AtomType
	}{
		{'t', Boolean},
		{'f', Boolean},
		{'0', Number},
		{'1', Number | Array},
		{'2', Map | Array | Boolean | Number | String},
		{'9', Number},
		{'-', Number},
		{'[', Array},
		{'{', Map},
		{'"', String},
		{'z', Undefined},
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

	mainAtom.Next("videoList", Array)
	wg := sync.WaitGroup{}
	for mainAtom.Next("video", Map) != nil {
		// create new atom video block (take slice bytes)
		videoAtom := mainAtom.Take()
		wg.Add(1)
		go func() {
			defer wg.Done()

			// find first
			if videoAtom.Next("name", Map).Next("text", String) == nil {
				fmt.Println("video name not found")
				return
			}
			// take and parsing object current cursor
			name := videoAtom.ToString()
			// move cursor next object
			if videoAtom.Next("duration", Map).Next("seconds", Number) == nil {
				return
			}
			// take atom and convert to float
			sec := videoAtom.ToFloat()
			// variant two check empty string
			author := videoAtom.End(). // move cursor to end object
							Prev("author", Map). // used reverse algorithm (find last string)
							Next("name", Map).
							Next("text", String).
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
		mainAtom.Start().Next("header", Map).Next("text", String).ToString(),
		mainAtom.End().Prev("clickTracking", Map).Next("id", String).ToString())

	j := NewAtom(bigData)
	fmt.Println(j.End().Prev("header").Next("musicImmersiveHeaderRenderer").Next("title").Next("text").ToString())

	NewAtom(bigData).End().Prev("musicTwoRowItemRenderer").Next("url")
}
func TestAtom_Parent(t *testing.T) {
	rf, _ := os.Open("./testdata/test.json")
	f, _ := ioutil.ReadAll(rf)
	mainAtom := NewAtom(f)
	if mainAtom.Next("duration", Map) != nil {
		p := mainAtom.Parent().Parent().Parent().Take()
		fmt.Println(p)

	}
}

func BenchmarkAtom_Next(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := NewAtom(bigData)
		j.End().Prev("musicTwoRowItemRenderer").Next("url")

	}
}

type A struct {
	Name string `json:"name"`
}

func TestAtom_ToString(t *testing.T) {
	a := A{
		Name: "[]№-\\/\":,'._<header></header>?%!;#@!\n$%^&*()山上的人\u00e5\u00b1\u00b1\u00e4\u00b8\u008a\u00e7\u009a\u0084\u00e4\u00ba\u00ba_+~` 😀\U0001F9BF 😈🕶☂",
	}
	b, _ := json.Marshal(a)
	atom := NewAtom(b).Next("name")
	str1 := atom.ToString()
	a2 := A{Name: str1}
	b2, _ := json.Marshal(a2)
	result := string(b) == string(b2)
	if !result {
		t.Fatal()
	}

}
