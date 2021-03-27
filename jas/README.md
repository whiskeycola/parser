# json core selector
 ## Use this library
> - if you need high performance
> - if you need a small piece of data
> - if you need to work with a large amount of data

### import
```shell
go get github.com/geniusmerely/parser/jas
```
## Example
testdata/test.json
```json
{
  "header": {
    "text": "Go learn videos"
  },
  "videoList": [
    {
      "video": {
        "name": {
          "text": "go learn 1"
        },
        "duration": {
          "text": "2:22",
          "seconds": 142
        },
        "views": {
          "text": "150K",
          "count": 150000
        },
        "navigation": {
          "id": "go-learn-1",
          "pageType": "VIDEO_PAGE"
        },
        "author": {
          "name": {
            "text": "go production"
          },
          "navigation": {
            "id": "go-production",
            "pageType": "USER_PAGE"
          }
        }
      }
    },
    {
      "video": {
        "name": {
          "text": "go learn 2"
        },
        "duration": {
          "text": "4:02",
          "seconds": 242
        },
        "views": {
          "text": "150K",
          "count": 150000
        },
        "navigation": {
          "id": "go-learn-1",
          "pageType": "VIDEO_PAGE"
        },
        "author": {
          "name": {
            "text": "go production"
          },
          "navigation": {
            "id": "go-production",
            "pageType": "USER_PAGE"
          }
        }
      }
    }
  ],
  "clickTracking": {
    "id": "FA09BC1156DC"
  }
}
```
```go
package main

import (
	"fmt"
	"github.com/geniusmerely/parser/jas"
	"io/ioutil"
	"os"
)

func main() {
    rf, _ := os.Open("./testdata/test.json")
    f, _ := ioutil.ReadAll(rf)
    mainAtom := jas.NewAtom(f)
    mainAtom.Next("videoList", jas.SelectArray)
    for mainAtom.Next("video", jas.SelectMap) != nil {
        // create new atom video block
        videoAtom := mainAtom.Take()
        // move cursor to first selector 
        if videoAtom.Next("name", jas.SelectMap).Next("text", jas.SelectString) == nil {
            fmt.Println("video name not found")
            return
        }
        // take data current cursor
        name := videoAtom.ToString()
        // move cursor next object
        if videoAtom.Next("duration", jas.SelectMap).Next("seconds", jas.SelectNumber) == nil {
            fmt.Println("video duration not found")
        	return
        }
        // take atom and convert to float
        sec := videoAtom.ToFloat()
        // variant two check empty string
        author := videoAtom.End(). // move cursor to end object
            Prev("author", jas.SelectMap). // used reverse algorithm (find last item)
            Next("name", jas.SelectMap).
            Next("text", jas.SelectString).
            ToString()
        // check text
        if author == "" {
            fmt.Println("video author not found")
            return
        }
        fmt.Printf("Name: %s\nDuration: %d\nAuthor: %s\n",name, int(sec), author)
    }
    fmt.Printf("Header: %s\nTrakingId: %s\n",
        mainAtom.Start().Next("header", jas.SelectMap).Next("text", jas.SelectString).ToString(),
        mainAtom.End().Prev("clickTracking", jas.SelectMap).Next("id", jas.SelectString).ToString())
}
```
## Multi thread

---
not use NewAtom(data)  
you need .Root() this will copy link to jnt cache  
cache has mutex  
```go
mainAtom := jas.NewAtom(data)
go func () {
    // free take new duplicate
    newAtom := mainAtom.Root()
}()
```
## Pointer and current object  

---

- atom.pointer - points where to start the next find (default 0)  
- atom.current - points to the current object (default Root)  
### shift atom.pointer
#### atom.Start()
- shift atom.pointer to first byte  
#### atom.End()
- shift atom.pointer to last byte +1
#### atom.Pass()
- shift atom.pointer to last byte+1 in current object  
- `atom.pointer += atom.current + current.Size()`  
- if atom.current == 0 analog .End()  
- if atom.current == -1 keeps position  
#### atom.Next()
- if found: atom.pointer = atom.current  
- if not found pointer keeps position to on the last found  

under this condition, the pointer remained to the Last found "title"
```go
atom.Next("title"). // found
	 Next("runs").  // not found
	 Next("text")   // not found
```
#### atom.Prev()
- if found: pointer = current - size(name object)  
- if not found: pointer keeps position to on the last found -   
### shift atom.current
#### atom.Next()
- shift to the first byte of found data
- if not found atom.current = -1
#### atom.Prev()
- the same behavior as atom.Next()
## Selection

---
### .Next()
find first instance of name in data
if found returns the self or nil
