// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"strings"
	"text/template"
)

type assetTree struct {
	Asset    binAsset
	Children map[string]*assetTree
	Depth    int
}

func newAssetTree() *assetTree {
	return &assetTree{
		Children: make(map[string]*assetTree),
	}
}

func (node *assetTree) child(name string) *assetTree {
	rv, ok := node.Children[name]
	if !ok {
		rv = newAssetTree()
		rv.Depth = node.Depth + 1
		node.Children[name] = rv
	}

	return rv
}

func init() {
	template.Must(baseTemplate.New("bintree").Funcs(template.FuncMap{
		"maxKeyLength": func(children map[string]*assetTree) int {
			l := 0
			for k := range children {
				if len(k) > l {
					l = len(k)
				}
			}

			return l
		},
	}).Parse(`bintree{
{{- if .Children}}
{{$max := maxKeyLength .Children -}}
{{range $k, $v := .Children -}}
{{repeat "\t" $v.Depth}}{{printf "%q" $k}}:
	{{- repeat " " (sub $max (len $k))}} {{template "bintree" $v}},
{{end -}}
{{- end -}}
{{- if .Children -}}
	{{repeat "\t" .Depth}}
{{- end -}}
}`))

	template.Must(baseTemplate.New("tree").Funcs(template.FuncMap{
		"tree": func(toc []binAsset) *assetTree {
			tree := newAssetTree()
			for _, asset := range toc {
				node := tree
				for _, name := range strings.Split(asset.Name, "/") {
					node = node.child(name)
				}

				node.Asset = asset
			}

			return tree
		},
	}).Parse(`
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree

	if name != "" {
		var ok bool
		canonicalName := strings.Replace(name, "\\", "/", -1)
		for _, p := range strings.Split(canonicalName, "/") {
			if node, ok = node[p]; !ok {
				return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
			}
		}
	}

	if len(node) == 0 {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	rv := make([]string, 0, len(node))
	for name := range node {
		rv = append(rv, name)
	}

	return rv, nil
}

type bintree map[string]bintree

var _bintree = {{template "bintree" (tree .Assets)}}
`))
}
