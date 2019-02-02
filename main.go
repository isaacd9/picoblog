package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday/v2"
	flag "github.com/spf13/pflag"
)

const VERSION = "0.1"

type outputType string

const (
	HTML outputType = "html"
	RSS  outputType = "rss"
	ATOM outputType = "atom"
)

var (
	title = flag.String("title", "Picoblog", "Title for blog")
	list  = flag.String("list", "", "List of blog posts, sorted by display order")
	mode  = flag.String("mode", "html", "Render in HTML or RSS mode. If RSS mode is set, the \"url\" flag must also be.")
	link  = flag.String("url", "", "URL to this blog to use in RSS feed")

	blogTemplate = template.Must(template.New("blog").Parse(
		`
<!DOCTYPE html>
<head>
<title>{{ .Title }}</title>
<style>
body {
	margin: 0 auto;
	padding: 2em 0px;
	max-width: 800px;
	color: #888;
	font-family: -apple-system,system-ui,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;
	font-size: 14px;
	line-height: 1.4em;
}
h1,h2,h3,h4   {color: #000;}
a {color: #000;}
a:visited {color: #888;}
</style>
</head>
<body>
<h4 style="padding-bottom: 2em">{{ .Title }}</h4>
{{ range .Posts }}
  <hr style="margin: 2em 0" />
  <div>
  <div style="text-align: right">
    <h3 id="{{ .Title }}" style="margin-bottom: .5em">{{ .Title }}</h3>
    <b>Updated {{ .Timestamp.Format "January 2nd, 2006" }}</b>
  </div>
  {{ .HTML }}
  </div>
{{ end }}
</body>
`))
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  picoblog takes a file of post paths

  Examples:
	picoblog first.md second.md
	picoblog --list file.txt

  Flags:
`)
		flag.PrintDefaults()
	}

	flag.ErrHelp = fmt.Errorf("")
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 && args[0] == "version" {
		fmt.Fprintf(os.Stderr, "%s\n", VERSION)
		return
	}

	var postNames []string
	if *list != "" {
		lis, err := os.Open(*list)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening post list: %+v\n", err)
		}

		s := bufio.NewScanner(lis)
		for s.Scan() {
			postNames = append(postNames, s.Text())
		}
	} else {
		postNames = args
	}

	if len(postNames) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No post provided\n\n")
		flag.Usage()
		os.Exit(1)
	}

	posts := getPosts(postNames)

	if *list == "" {
		sort.SliceStable(posts, func(i, j int) bool {
			return posts[i].Timestamp.After(posts[j].Timestamp)
		})
	}

	mode := strings.ToLower(*mode)
	switch outputType(mode) {
	case HTML:
		renderHTML(os.Stdout, posts)
	case RSS:
		if *link == "" {
			fmt.Fprintf(os.Stderr, "ERROR: URL must be specified in RSS mode\n\n")
			os.Exit(1)
		}
		renderFeed(os.Stdout, posts, *link, RSS)
	case ATOM:
		if *link == "" {
			fmt.Fprintf(os.Stderr, "ERROR: URL must be specified in Atom mode\n\n")
			os.Exit(1)
		}
		renderFeed(os.Stdout, posts, *link, ATOM)
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Unsupported mode %q\n\n", mode)
	}
}

type post struct {
	Title     string
	Timestamp time.Time
	Contents  string
}

func (p *post) HTML() string {
	htmlContents := blackfriday.Run([]byte(p.Contents))
	return string(htmlContents)
}

func (p *post) FeedItem(baseURL string) *feeds.Item {
	return &feeds.Item{
		Title: p.Title,
		Link: &feeds.Link{
			Href: fmt.Sprintf("%s#%s", baseURL, url.PathEscape(p.Title)),
		},
	}
}

func getPosts(filenames []string) (posts []*post) {
	for _, filename := range filenames {
		post, err := getPost(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building post: %v", err)
		}

		posts = append(posts, post)
	}
	return posts
}

func getPost(filename string) (*post, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %q: %v", filename, err)
	}

	name := strings.TrimRight(info.Name(), path.Ext(info.Name()))

	p := &post{
		Title:     name,
		Timestamp: info.ModTime(),
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %v", filename, err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(bufio.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %v", filename, err)
	}

	p.Contents = string(b)
	return p, nil
}

func renderHTML(w io.Writer, posts []*post) {
	blogTemplate.Execute(w, struct {
		Title string
		Posts []*post
	}{
		*title,
		posts,
	})
}

func renderFeed(w io.Writer, posts []*post, baseURL string, t outputType) {
	feed := &feeds.Feed{
		Title: *title,
		Link:  &feeds.Link{Href: baseURL},
		Items: []*feeds.Item{},
	}

	for _, post := range posts {
		feed.Items = append(feed.Items, post.FeedItem(baseURL))
	}

	var xmlFeed feeds.XmlFeed
	switch t {
	case RSS:
		xmlFeed = (&feeds.Rss{Feed: feed}).RssFeed()
	case ATOM:
		xmlFeed = (&feeds.Atom{Feed: feed}).AtomFeed()
	}

	if err := feeds.WriteXML(xmlFeed, w); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not render feed %v\n\n", err)
	}
}
