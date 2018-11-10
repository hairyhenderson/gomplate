// Package json implements functions to load the Public key data from an EJSON
// file, and to walk that data file, encrypting or decrypting any keys which,
// according to the specification, are marked as encryptable (see README.md for
// details).
//
// It may be non-obvious why this is implemented using a scanner and not by
// loading the structure, manipulating it, then dumping it. Since Go's maps are
// explicitly randomized, that would cause the entire structure to be randomized
// each time the file was written, rendering diffs over time essentially
// useless.
package json

import (
	"bytes"
	"fmt"

	"github.com/dustin/gojson"
)

// Walker takes an Action, which will run on fields selected by EJSON for
// encryption, and provides a Walk method, which iterates on all the fields in
// a JSON text, running the Action on all selected fields. Fields are selected
// if they are a Value (not a Key) of type string, and their referencing Key did
// *not* begin with an Underscore. Note that this
// underscore-to-disable-encryption syntax does not propagate down the hierarchy
// to children.
// That is:
//   * In {"_a": "b"}, Action will not be run at all.
//   * In {"a": "b"}, Action will be run with "b", and the return value will
//      replace "b".
//   * In {"k": {"a": ["b"]}, Action will run on "b".
//   * In {"_k": {"a": ["b"]}, Action run on "b".
//   * In {"k": {"_a": ["b"]}, Action will not run.
type Walker struct {
	Action func([]byte) ([]byte, error)
}

// Walk walks an entire JSON structure, running the ejsonWalker.Action on each
// actionable node. A node is actionable if it's a string *value*, and its
// referencing key doesn't begin with an underscore. For each actionable node,
// the contents are replaced with the result of Action. Everything else is
// unchanged.
func (ew *Walker) Walk(data []byte) ([]byte, error) {
	var (
		inLiteral    bool
		literalStart int
		isComment    bool
		scanner      json.Scanner
	)
	scanner.Reset()
	pline := newPipeline()
	for i, c := range data {
		switch v := scanner.Step(&scanner, int(c)); v {
		case json.ScanContinue, json.ScanSkipSpace:
			// Uninteresting byte. Just advance to next.
		case json.ScanBeginLiteral:
			inLiteral = true
			literalStart = i
		case json.ScanObjectKey:
			// The literal we just finished reading was a Key. Decide whether it was a
			// encryptable by checking whether the first byte after the '"' was an
			// underscore, then append it verbatim to the output buffer.
			inLiteral = false
			isComment = data[literalStart+1] == '_'
			pline.appendBytes(data[literalStart:i])
		case json.ScanError:
			// Some error happened; just bail.
			pline.flush()
			return nil, fmt.Errorf("invalid json")
		case json.ScanEnd:
			// We successfully hit the end of input.
			pline.appendByte(c)
			return pline.flush()
		default:
			if inLiteral {
				inLiteral = false
				// We finished reading some literal, and it wasn't a Key, meaning it's
				// potentially encryptable. If it was a string, and the most recent Key
				// encountered didn't begin with a '_', we are to encrypt it. In any
				// other case, we append it verbatim to the output buffer.
				if isComment || data[literalStart] != '"' {
					pline.appendBytes(data[literalStart:i])
				} else {
					res := make(chan promiseResult)
					go func(subData []byte) {
						actioned, err := ew.runAction(subData)
						res <- promiseResult{actioned, err}
						close(res)
					}(data[literalStart:i])
					pline.appendPromise(res)
				}
			}
		}
		if !inLiteral {
			// If we're in a literal, we save up bytes because we may have to encrypt
			// them. Outside of a literal, we simply append each byte as we read it.
			pline.appendByte(c)
		}
	}
	if scanner.EOF() == json.ScanError {
		// Unexpected EOF => malformed JSON
		pline.flush()
		return nil, fmt.Errorf("invalid json")
	}
	return pline.flush()
}

func (ew *Walker) runAction(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	unquoted, ok := json.UnquoteBytes(trimmed)
	if !ok {
		return nil, fmt.Errorf("invalid json")
	}
	done, err := ew.Action(unquoted)
	if err != nil {
		return nil, err
	}
	quoted, err := quoteBytes(done)
	if err != nil {
		return nil, err
	}
	return append(quoted, data[len(trimmed):]...), nil
}

// probably a better way to do this, but...
func quoteBytes(in []byte) ([]byte, error) {
	data := []string{string(in)}
	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return out[1 : len(out)-1], nil
}
