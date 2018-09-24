// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
)

type Properties map[string]string

type entry struct {
	K, V string
}

func (props Properties) MarshalJSON() ([]byte, error) {
	es := props.entries()
	buf := new(bytes.Buffer)
	props.writeJSON(buf, es)
	return buf.Bytes(), nil
}

func (props Properties) writeJSON(w io.Writer, es []entry) {
	enc := json.NewEncoder(w)
	io.WriteString(w, "{")
	for i, e := range es {
		enc.Encode(e.K)
		io.WriteString(w, ":")
		enc.Encode(e.V)
		if i+1 < len(es) {
			io.WriteString(w, ",")
		}
	}
	io.WriteString(w, "}")
}

func (props Properties) entries() []entry {
	es := make([]entry, 0, len(props))
	for k, v := range props {
		es = append(es, entry{k, v})
	}
	sort.Slice(es, func(i, j int) bool { return es[i].K < es[j].K })
	return es
}
