package jas

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type AtomType int

const (
	Undefined AtomType = 0
	Any                = Undefined
	String    AtomType = 1 << iota
	Array
	Map
	Boolean
	Number
	Null
)

type atom struct {
	vector  []byte
	pointer int
	current int
	cache   *cache
}

func NewAtom(vector []byte) *atom {
	if vector == nil {
		vector = []byte{}
	}
	return &atom{
		vector: vector,
		cache:  newCache(),
	}
}

func (a *atom) Next(name string, atp ...AtomType) *atom {
	if a == nil {
		return nil
	}
	sep := []byte(fmt.Sprintf(`"%s"`, name))
	if a.pointer >= len(a.vector)-len(sep) {
		a.current = -1
		return nil
	}
	tp := selectAtot(atp...)
	for start := a.pointer; start < len(a.vector); {
		i := bytes.Index(a.vector[start:], sep)
		if i == -1 {
			a.current = -1
			return nil
		}

		if i > 0 && a.vector[start+i-1] == '\\' {
			start += len(sep)
			continue
		}

		c := atomIndex(a.vector, start+i+len(sep))
		if c == -1 {
			start += i + len(sep)
			continue
		}
		if c == a.current {
			start = c
			continue
		}
		if c >= len(a.vector) {
			return nil
		}
		if c != -1 {
			if tp == Any || isType(a.vector[c], AtomType(tp)) {
				a.pointer = c
				a.current = c
				return a
			}
		}
		start = c
	}
	a.current = -1
	return nil
}

func (a *atom) Prev(name string, atp ...AtomType) *atom {
	if a == nil {
		return nil
	}
	sep := []byte(fmt.Sprintf(`"%s"`, name))
	if a.pointer < 0 && a.pointer > len(a.vector) {
		a.current = -1
		return nil
	}
	tp := selectAtot(atp...)
	for end := a.pointer; end >= 0; {
		i := bytes.LastIndex(a.vector[:end], sep)
		if i == -1 {
			a.current = -1
			return nil
		}

		if i > 0 && a.vector[i-1] == '\\' {
			end = i
			continue

		}
		c := atomIndex(a.vector, i+len(sep))
		if c == a.current {
			end = i
			continue
		}
		if c >= len(a.vector) {
			a.current = -1
			return nil
		}
		if c != -1 {
			if tp == Any || isType(a.vector[c], AtomType(tp)) {
				a.pointer = i
				a.current = c
				return a
			}
		}
		end = i
	}
	a.current = -1
	return nil
}
func (a *atom) Parent() *atom {
	if a == nil {
		return nil
	}
	i := a.current
	aStack := 0
	mStack := 0
	for i--; i > 0; i-- {
		switch a.vector[i] {
		case '{':
			if i > 0 && a.vector[i-1] == '\\' {
				continue
			}
			if mStack == 0 {
				a.current = i
				a.pointer = i
				return a
			}
			mStack--
		case '}':
			if i > 0 && a.vector[i-1] == '\\' {
				continue
			}
			mStack++
			continue
		case '[':
			if i > 0 && a.vector[i-1] == '\\' {
				continue
			}
			if aStack == 0 {
				a.current = i
				a.pointer = i
				return a
			}
			aStack--
		case ']':
			if i > 0 && a.vector[i-1] == '\\' {
				continue
			}
			aStack++
		}
	}
	a.current = -1
	return nil
}
func (a *atom) Pointer() int {
	return a.pointer
}
func (a *atom) Move(i int) *atom {
	switch {
	case a == nil:
		return nil
	case i > len(a.vector):
		a.pointer = len(a.vector)
	case i < 0:
		a.pointer = 0
	default:
		a.pointer = i
	}
	return a
}
func (a *atom) Start() *atom {
	if a == nil {
		return nil
	}
	a.pointer = 0
	return a
}
func (a *atom) End() *atom {
	if a == nil {
		return nil
	}
	a.pointer = len(a.vector)
	return a
}
func (a *atom) Pass() *atom {
	if a.Type() == Undefined {
		return nil
	}
	if a.current == 0 {
		a.pointer = len(a.vector)
		return a
	}
	a.pointer = a.current + a.Size()
	return a
}

func (a *atom) Root() *atom {
	if a.Type() == Undefined {
		return nil
	}
	return &atom{
		vector: a.vector,
		cache:  a.cache,
	}
}

func (a *atom) Take() *atom {
	if a.Type() == Undefined {
		return nil
	}
	if a.current == 0 {
		return &atom{
			vector: a.vector,
			cache:  a.cache,
		}
	}
	return parseAny(a, a.current)
}
func (a *atom) ToArray() []*atom {
	arr := make([]*atom, 0, 10)
	if a.Type() != Array {
		return arr
	}
	for i := a.current + 1; i <= len(a.vector); i++ {
		i = seekSpace(a.vector, i)
		if a.vector[i] == ',' {
			continue
		}
		z := string(a.vector[i])
		_ = z
		v := parseAny(a, i)
		if v == nil {
			return arr
		}
		i += len(v.vector)
		arr = append(arr, v)
	}
	return arr
}
func (a *atom) ToMap() map[string]*atom {
	arr := make(map[string]*atom, 0)
	if a.Type() != Map {
		return arr
	}
	for i := a.current + 1; i <= len(a.vector); i++ {
		i = seekSpace(a.vector, i)
		if a.vector[i] == ',' {
			continue
		}
		name := parseString(a.vector, i)
		if name == nil {
			return arr
		}
		i += len(name.vector)
		i = atomIndex(a.vector, i)
		if i == -1 {
			return arr
		}
		v := parseAny(a, i)
		if v == nil {
			return arr
		}
		i += len(v.vector)
		arr[name.ToString()] = v
	}
	return arr
}
func (a *atom) ToString() string {
	t := a.Type()
	if t&(Number|String|Boolean) == 0 {
		return ""
	}
	switch t {
	case Boolean:
		if a.vector[a.current] == 't' {
			return "true"
		} else {
			return "false"
		}
	case String:
		if a.current == 0 {
			if len(a.vector) <= 2 {
				return ""
			}
			// skip quotes
			str := string(a.vector[1 : len(a.vector)-1])
			str2, err := strconv.Unquote(strings.Replace(strconv.Quote(str), `\\u`, `\u`, -1))
			if err != nil {
				str2 = str
			}
			return strings.NewReplacer(`\n`, "\n", `\"`, `"`, `\/`, `/`, `\\`, `\`).Replace(str2)
		} else {
			return a.Take().ToString()
		}
	case Number:
		if a.current == 0 {
			return string(a.vector)
		} else {
			return a.Take().ToString()
		}
	default:
		return ""
	}
}

// If you need precision
// Before call if atom.Type() == Boolean
func (a *atom) ToBoolean() bool {
	if a.Type() != Boolean {
		return false
	}
	if a.vector[a.current] == 't' {
		return true
	}
	return false
}
func (a *atom) ToFloat() float64 {
	t := a.Type()
	if t&(Number|String) == 0 {
		return 0
	}
	str := ""
	switch t {
	case Number:
		if a.current == 0 {
			str = a.ToString()
		} else {
			str = parseNumber(a.vector, a.current).ToString()
		}
	case String:
		str = a.ToString()
	}

	f, _ := strconv.ParseFloat(str, 64)
	return f
}

func (a *atom) Type() AtomType {
	if a == nil || a.current >= len(a.vector) || a.current < 0 {
		return Undefined
	}
	return getType(a.vector[a.current])
}

func (a *atom) Byte() []byte {
	if a.Type() == Undefined {
		return []byte{}
	}
	if a.current == 0 {
		return a.vector
	}
	return a.Take().Byte()
}

func (a *atom) Size() int {
	if a.Type() == Undefined {
		return 0
	}
	if a.current == 0 {
		return len(a.vector)
	}
	if size, ok := a.cache.Get(a.current); ok {
		return size
	}
	return a.Take().Size()
}
