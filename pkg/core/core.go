// Copyright 2014 Richard Lehane. All rights reserved.
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

// Package core defines the Siegfried struct and Identifier/Identification interfaces.
// The packages within core (bytematcher and namematcher) provide a toolkit for building identifiers based on different signature formats (such as PRONOM).
package core

import (
	"errors"

	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/signature"
)

// Identifier describes different signature formats. E.g. there is a PRONOM identifier that implements the TNA's format.
type Identifier interface {
	Recorder() Recorder
	Describe() [2]string
	Save(*signature.LoadSaver)
	String() string
	Recognise(MatcherType, int) (bool, string) // do you recognise this index
}

// Add additional identifier types here
const (
	Pronom byte = iota
)

type IdentifierLoader func(*signature.LoadSaver) Identifier

var loaders = [8]IdentifierLoader{nil, nil, nil, nil, nil, nil, nil, nil}

func RegisterIdentifier(id byte, l IdentifierLoader) {
	loaders[int(id)] = l
}

func LoadIdentifier(ls *signature.LoadSaver) Identifier {
	id := ls.LoadByte()
	l := loaders[int(id)]
	if l == nil {
		if ls.Err == nil {
			ls.Err = errors.New("bad identifier loader")
		}
		return nil
	}
	return l(ls)
}

// Recorder is a mutable object generated by an identifier. It records match results and sends identifications
type Recorder interface {
	Record(MatcherType, Result) bool // Record results for each matcher; return true if match recorded (siegfried will iterate through the identifiers until an identifier returns true)
	Satisfied() bool                 // Called after matchers - should we continue with further matchers?
	Report(chan Identification)      // Now send results
	Compress() bool                  // Is this a compressed format?
}

// Identification is sent by an identifier when a format matches
type Identification interface {
	String() string // short text that is displayed to indicate the format match
	Yaml() string   // long text that should be displayed to indicate the format match
	Json() string   // JSON match response
	Csv() []string  // CSV match response
}

// Matcher does the matching (against the name or the byte stream) and sends results
type Matcher interface {
	Identify(string, siegreader.Buffer) (chan Result, error)
	Add(SignatureSet, priority.List) (int, error) // add a signature set, return total number of signatures in a matcher
	String() string
	Save(*signature.LoadSaver)
}

// MatcherType is used by recorders to tell which type of matcher has sent a result
type MatcherType int

const (
	ExtensionMatcher MatcherType = iota
	ContainerMatcher
	ByteMatcher
)

// SignatureSet is added to a matcher. It can take any form, depending on the matcher
type SignatureSet interface{}

// Result is a raw hit that matchers pass on to Identifiers
type Result interface {
	Index() int
	Basis() string
}
