package responsemodifers

import (
	"bytes"
	"embed"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"html/template"
	"io"
	"ladder/proxychain"
	"log"

	//"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
)

//go:embed generate_readable_outline.html
var templateFS embed.FS

// GenerateReadableOutline creates an reader-friendly distilled representation of the article.
// This is a reliable way of bypassing soft-paywalled articles, where the content is hidden, but still present in the DOM.
func GenerateReadableOutline() proxychain.ResponseModification {

	// get template only once, and resuse for subsequent calls
	f := "generate_readable_outline.html"
	tmpl, err := template.ParseFS(templateFS, f)
	if err != nil {
		panic(fmt.Errorf("tx.GenerateReadableOutline Error: %s not found", f))
	}

	return func(chain *proxychain.ProxyChain) error {

		// ===========================================================
		// 1. extract dom contents using reading mode algo
		// ===========================================================
		opts := trafilatura.Options{
			IncludeImages:      true,
			IncludeLinks:       true,
			FavorRecall:        true,
			Deduplicate:        true,
			FallbackCandidates: nil, // TODO: https://github.com/markusmobius/go-trafilatura/blob/main/examples/chained/main.go
			// implement fallbacks from	"github.com/markusmobius/go-domdistiller" and 	"github.com/go-shiori/go-readability"
			OriginalURL: chain.Request.URL,
		}

		extract, err := trafilatura.Extract(chain.Response.Body, opts)
		if err != nil {
			return err
		}

		// ============================================================================
		// 2. render generate_readable_outline.html template using metadata from step 1
		// ============================================================================

		// render DOM to string without H1 title
		removeFirstH1(extract.ContentNode)
		var b bytes.Buffer
		html.Render(&b, extract.ContentNode)
		distilledHTML := b.String()

		// populate template parameters
		data := map[string]interface{}{
			"Success":     true,
			"Footer":      extract.Metadata.License,
			"Image":       extract.Metadata.Image,
			"Description": extract.Metadata.Description,
			"Hostname":    extract.Metadata.Hostname,
			"Url":         chain.Request.URL,
			"Title":       extract.Metadata.Title, // todo: modify CreateReadableDocument so we don't have <h1> titles duplicated?
			"Date":        extract.Metadata.Date.String(),
			"Author":      extract.Metadata.Author,
			"Body":        distilledHTML,
		}

		// ============================================================================
		// 3. queue sending the response back to the client by replacing the response body
		// (the response body will be read as a stream in proxychain.Execute() later on.)
		// ============================================================================
		pr, pw := io.Pipe() // pipe io.writer contents into io.reader

		// Use a goroutine for writing to the pipe so we don't deadlock the request
		go func() {
			defer pw.Close()

			err := tmpl.Execute(pw, data) // <- render template

			if err != nil {
				log.Printf("WARN: GenerateReadableOutline template rendering error: %s\n", err)
			}
		}()

		chain.Context.Set("content-type", "text/html")
		chain.Response.Body = pr // <- replace reponse body reader with our new reader from pipe
		return nil
	}
}

func removeFirstH1(n *html.Node) {
	var recurse func(*html.Node) bool
	recurse = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.DataAtom == atom.H1 {
			return true // Found the first H1, return true to stop
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if recurse(c) {
				n.RemoveChild(c)
				return false // Removed first H1, no need to continue
			}
		}
		return false
	}
	recurse(n)
}
