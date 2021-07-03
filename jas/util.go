package jas

import (
	"bytes"
	"strconv"
	"strings"
)

func parseAny(a *atom, s int) *atom {
	if n := a.cache.takeAtom(a.vector, a.current); n != nil {
		return n
	}
	switch getType(a.vector[s]) {
	case Map:
		return parseObject(a, s)
	case Array:
		return parseObject(a, s)
	case String:
		str := parseString(a.vector, s)
		a.cache.Add(s, len(str.vector))
		return str
	case Boolean:
		return parseBoolean(a.vector, s)
	case Number:
		return parseNumber(a.vector, s)
	case Null:
		return parseNull(a.vector, s)
	default:
		return nil
	}
}

func parseObject(a *atom, s int) *atom {
	at := getType(a.vector[s])
	if at&(Array|Map) == 0 {
		return nil
	}
	if na := a.cache.takeAtom(a.vector, s); na != nil {
		return na
	}
	var cs, cf byte = '{', '}'
	if at == Array {
		cs, cf = '[', ']'

	}
	i := s
	na := &atom{
		cache: newCache(),
	}
	st := make(stack, 0)
	st.Push(i)

FOR:
	for i++; i < len(a.vector); i++ {
		if isType(a.vector[i], String) {
			i = seekString(a.vector, i)
			if i >= len(a.vector) {
				break
			}
		}
		//_TODO: recurse parseObject???
		switch a.vector[i] {
		case cs:
			st.Push(i)
		case cf:
			// invalid json syntax
			if len(st) == 0 {
				return nil
			}
			p := st.Pop()
			if len(st) == 0 {
				a.cache.Add(s, i-s)
				break FOR
			}
			a.cache.Add(p, i-p)
			na.cache.Add(p-s, i-p)
		}
	}
	if i < len(a.vector) {
		i++
	}
	na.vector = a.vector[s:i]
	return na
}

func parseBoolean(slice []byte, s int) *atom {
	switch slice[s] {
	case 't':
		return &atom{vector: slice[s : s+4], cache: newCache()}
	case 'f':
		return &atom{vector: slice[s : s+5], cache: newCache()}
	default:
		return nil
	}
}
func parseNull(slice []byte, s int) *atom {
	if slice[s] == 'n' {
		return &atom{vector: slice[s : s+4]}
	}
	return nil
}
func parseString(slice []byte, s int) *atom {
	if !isType(slice[s], String) {
		return nil
	}
	e := seekString(slice, s)
	return &atom{vector: slice[s:e], cache: newCache()}
}

func parseNumber(slice []byte, s int) *atom {
	i := s
FOR:
	for ; i < len(slice); i++ {
		switch slice[i] {
		case '-', '+', 'e', 'E', '.':
			continue
		default:
			if slice[i] >= '0' && slice[i] <= '9' {
				continue
			}
			break FOR
		}
	}
	return &atom{vector: slice[s:i], cache: newCache()}
}

func isType(c byte, t AtomType) bool {
	switch {
	case t&String != 0 && c == '"':
		return true
	case t&Null != 0 && c == 'n':
		return true
	case t&Number != 0 && (c == '-' || (c >= '0' && c <= '9')):
		return true
	case t&Boolean != 0 && (c == 'f' || c == 't'):
		return true
	case t&Array != 0 && c == '[':
		return true
	case t&Map != 0 && c == '{':
		return true
	case t == Undefined:
		return !isType(c, Map|Array|Boolean|Number|String)
	default:
		return false
	}
}
func getType(c byte) AtomType {
	switch {
	case isType(c, Map):
		return Map
	case isType(c, Array):
		return Array
	case isType(c, Boolean):
		return Boolean
	case isType(c, Number):
		return Number
	case isType(c, String):
		return String
	case isType(c, Null):
		return Null
	default:
		return Undefined
	}
}
func selectAtot(t ...AtomType) AtomType {
	var tp AtomType
	for _, v := range t {
		tp |= v
	}
	return tp
}
func seekSpace(slice []byte, i int) int {
	for ; i < len(slice); i++ {
		switch slice[i] {
		case 32, // space
			9,  // tab
			10, // \n
			13: // \r
			continue
		default:
			return i
		}
	}
	return i
}
func atomIndex(slice []byte, start int) (i int) {
	i = seekSpace(slice, start)
	if i < len(slice) && slice[i] == ':' {
		return seekSpace(slice, i+1)
	}
	return -1
}
func seekString(slice []byte, s int) int {
	if !isType(slice[s], String) {
		return s
	}
	for s++; s < len(slice); s++ {
		i := bytes.IndexByte(slice[s:], '"')
		if i == -1 {
			s = len(slice)
		}
		s += i
		a := string(slice[s-1]) + string(slice[s])
		_ = a
		if slice[s-1] != '\\' {
			break
		}
	}
	if s < len(slice) {
		s++
	}
	return s
}
func UnescapeString(str string) string {
	s, err := strconv.Unquote(strings.Replace(strconv.Quote(str), `\\u`, `\u`, -1))
	if err != nil {
		s = str
	}
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\/`, `/`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}
