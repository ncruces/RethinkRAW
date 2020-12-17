// Copyright (c) 2020 Nuno Cruces
// SPDX-License-Identifier: MIT

package xmp

import (
	"encoding/xml"
	"io"
	"strings"
)

// IsSidecarForExt checks if a reader is the sidecar for a given file extension.
func IsSidecarForExt(r io.Reader, ext string) bool {
	test := func(name xml.Name) bool {
		return name.Local == "SidecarForExtension" &&
			(name.Space == "http://ns.adobe.com/photoshop/1.0/" || name.Space == "photoshop")
	}

	dec := xml.NewDecoder(r)
	for {
		t, err := dec.Token()
		if err != nil {
			return err == io.EOF // assume yes
		}

		if s, ok := t.(xml.StartElement); ok {
			if test(s.Name) {
				t, _ := dec.Token()
				v, ok := t.(xml.CharData)
				return ok && strings.EqualFold(ext, "."+string(v))
			}
			for _, a := range s.Attr {
				if test(a.Name) {
					return strings.EqualFold(ext, "."+a.Value)
				}
			}
		}
	}
}
