package view_test

import (
	"fmt"
	"os"
	"testing"
	"testing/fstest"

	"github.com/hexops/valast"
	"github.com/livebud/bud/internal/is"
	"github.com/livebud/bud/runtime/view"
)

func TestIndex(t *testing.T) {
	is := is.New(t)
	fsys := fstest.MapFS{
		"index.gohtml": &fstest.MapFile{Data: []byte("Hello {{ .Planet }}!")},
	}
	// Find the pages
	pages, err := view.Find(fsys)
	is.NoErr(err)
	is.Equal(len(pages), 1)
	is.True(pages["index"] != nil)
	is.Equal(pages["index"].Path, "index.gohtml")
	is.Equal(len(pages["index"].Frames), 0)
	is.Equal(pages["index"].Layout, nil)
	is.Equal(pages["index"].Error, nil)
}

func TestNested(t *testing.T) {
	is := is.New(t)
	fsys := fstest.MapFS{
		"layout.svelte":      &fstest.MapFile{Data: []byte(`<slot />`)},
		"frame.svelte":       &fstest.MapFile{Data: []byte(`<slot />`)},
		"posts/frame.svelte": &fstest.MapFile{Data: []byte(`<slot />`)},
		"posts/index.svelte": &fstest.MapFile{Data: []byte(`<h1>Hello {planet}!</h1>`)},
	}
	// Find the pages
	pages, err := view.Find(fsys)
	is.NoErr(err)
	is.Equal(len(pages), 1)
	is.True(pages["posts/index"] != nil)
	is.Equal(pages["posts/index"].Path, "posts/index.svelte")

	// Frames
	is.Equal(len(pages["posts/index"].Frames), 2)
	is.Equal(pages["posts/index"].Frames[0].Key, "frame")
	is.Equal(pages["posts/index"].Frames[0].Path, "frame.svelte")
	is.Equal(pages["posts/index"].Frames[1].Key, "posts/frame")
	is.Equal(pages["posts/index"].Frames[1].Path, "posts/frame.svelte")

	is.Equal(pages["posts/index"].Error, nil)

	// Layout
	is.True(pages["posts/index"].Layout != nil)
	is.Equal(pages["posts/index"].Layout.Key, "layout")
	is.Equal(pages["posts/index"].Layout.Path, "layout.svelte")
}

func TestDocs(t *testing.T) {
	is := is.New(t)
	fsys := os.DirFS("../../example/docs/view")
	// Find the pages
	pages, err := view.Find(fsys)
	is.NoErr(err)
	fmt.Println(valast.String(pages))
}