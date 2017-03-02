// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a Modified
// BSD License that can be found in the LICENSE file.

package bindata

import (
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

func init() {
	template.Must(baseTemplate.New("header").Funcs(template.FuncMap{
		"trimright": func(s string) string {
			return strings.TrimRightFunc(s, unicode.IsSpace)
		},
		"toslash": filepath.ToSlash,
	}).Parse(`{{- /* This makes e.g. Github ignore diffs in generated files. */ -}}
// Code generated by go-bindata.
{{if $.Dev -}}
	//  debug: dev
{{else if $.Debug -}}
	//  debug: true
{{end -}}
{{- if $.MemCopy -}}
	//  memcopy: true
{{end -}}
{{- if $.Compress -}}
	//  compress: true
{{end -}}
{{- if and $.Compress $.DecompressOnce -}}
	//  decompress: once
{{end -}}
{{- if $.Metadata -}}
	//  metadata: true
{{end -}}
{{- if $.Mode -}}
	//  mode: {{printf "%04o" $.Mode}}
{{end -}}
{{- if $.ModTime -}}
	//  modtime: {{$.ModTime}}
{{end -}}
{{- if $.AssetDir -}}
	//  asset-dir: true
{{end -}}
{{- if $.Restore -}}
	//  restore: true
{{end -}}
{{- if $.Hash -}}
{{- if $.HashFormat -}}
	//  hash-format: {{$.HashFormat}}
{{else -}}
	//  hash-format: unchanged
{{end -}}
{{- if and $.HashFormat $.HashLength (ne $.HashLength 16) -}}
	//  hash-length: {{$.HashLength}}
{{end -}}
{{- if and $.HashFormat $.HashEncoding -}}
	//  hash-encoding: {{$.HashEncoding}}
{{end -}}
{{- end -}}
// sources:
{{range .Assets -}}
	//  {{toslash (trimright .Path)}}
{{end -}}
// DO NOT EDIT!

{{if $.Tags -}}
	// +build {{$.Tags}}

{{end -}}

package {{$.Package}}`))
}
