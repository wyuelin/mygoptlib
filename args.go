package pt

import (
	"bytes"
	"errors"
	"fmt"
)

// Key–value mappings for the representation of client and server options.

// Args maps a string key to a list of values. It is similar to url.Values.
type Args map[string][]string

// Get the first value associated with the given key. If there are any value
// associated with the key, the ok return value is true; otherwise it is false.
// If you need access to multiple values, use the map directly.
func (args Args) Get(key string) (value string, ok bool) {
	if args == nil {
		return "", false
	}
	vals, ok := args[key]
	if !ok || len(vals) == 0 {
		return "", false
	}
	return vals[0], true
}

// Append value to the list of values for key.
func (args Args) Add(key, value string) {
	args[key] = append(args[key], value)
}

// Return the index of the next unescaped byte in s that is in the term set, or
// else the length of the string if not terminators appear. Additionally return
// the unescaped string up to the returned index.
func indexUnescaped(s string, term []byte) (int, string, error) {
	var i int
	unesc := make([]byte, 0)
	for i = 0; i < len(s); i++ {
		b := s[i]
		// A terminator byte?
		if bytes.IndexByte(term, b) != -1 {
			break
		}
		if b == '\\' {
			i++
			if i >= len(s) {
				return 0, "", errors.New(fmt.Sprintf("nothing following final escape in %q", s))
			}
			b = s[i]
		}
		unesc = append(unesc, b)
	}
	return i, string(unesc), nil
}

// Parse a name–value mapping as from an encoded SOCKS username/password.
//
// "If any [k=v] items are provided, they are configuration parameters for the
// proxy: Tor should separate them with semicolons ... If a key or value value
// must contain [an equals sign or] a semicolon or a backslash, it is escaped
// with a backslash."
func parseClientParameters(s string) (args Args, err error) {
	args = make(Args)
	if len(s) == 0 {
		return
	}
	i := 0
	for {
		var key, value string
		var offset, begin int

		begin = i
		// Read the key.
		offset, key, err = indexUnescaped(s[i:], []byte{'=', ';'})
		if err != nil {
			return
		}
		i += offset
		// End of string or no equals sign?
		if i >= len(s) || s[i] != '=' {
			err = errors.New(fmt.Sprintf("no equals sign in %q", s[begin:i]))
			return
		}
		// Skip the equals sign.
		i++
		// Read the value.
		offset, value, err = indexUnescaped(s[i:], []byte{';'})
		if err != nil {
			return
		}
		i += offset
		if len(key) == 0 {
			err = errors.New(fmt.Sprintf("empty key in %q", s[begin:i]))
			return
		}
		args.Add(key, value)
		if i >= len(s) {
			break
		}
		// Skip the semicolon.
		i++
	}
	return args, nil
}
