// Copyright 2017 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reader

import (
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	Path int = iota
	Filename
	FilenameSize
	FilenameMod
	FilenameHash
	Hash
)

func keygen(join int, fi File) string {
	switch join {
	default:
		return fi.Path
	case Filename:
		return filepath.Base(fi.Path)
	case FilenameSize:
		return filepath.Base(fi.Path) + strconv.FormatInt(fi.Size, 10)
	case FilenameMod:
		return filepath.Base(fi.Path) + fi.Mod
	case FilenameHash:
		return filepath.Base(fi.Path) + string(fi.Hash)
	case Hash:
		return string(fi.Hash)
	}
}

func idStr(fi File) string {
	ids := make([]string, len(fi.IDs))
	for i, id := range fi.IDs {
		ids[i] = id.String()
	}
	sort.Strings(ids)
	return strings.Join(ids, ";")
}

func matches(res []string) bool {
	if len(res) < 3 {
		return false
	}
	m := res[1]
	for _, r := range res[2:] {
		if r != m {
			return false
		}
	}
	return true
}

func Compare(w io.Writer, join int, paths ...string) error {
	if len(paths) < 2 {
		return fmt.Errorf("at least two results files must be provided for comparison; got %d", len(paths))
	}
	readers := make([]Reader, len(paths))
	for i, v := range paths {
		rdr, err := Open(v)
		if err != nil {
			return err
		}
		readers[i] = rdr
	}
	files := make([]string, 0, 1000)
	results := make(map[string][]string)
	for i, rdr := range readers {
		for f, e := rdr.Next(); e == nil; f, e = rdr.Next() {
			key := keygen(join, f)
			_, ok := results[key]
			if !ok {
				files = append(files, key)
				def := make([]string, len(readers)+1)
				def[0] = f.Path
				for i := range def[1:] {
					def[i+1] = "MISSING"
				}
				results[key] = def
			}
			results[key][i+1] = idStr(f)
		}
	}
	wrt := csv.NewWriter(w)
	var complete bool = true
	for _, f := range files {
		if !matches(results[f]) {
			complete = false
			if err := wrt.Write(results[f]); err != nil {
				return err
			}
		}
	}
	wrt.Flush()
	if complete {
		fmt.Fprint(w, "COMPLETE MATCH")
	}
	return nil
}
