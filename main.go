package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/russross/blackfriday/v2"
	flag "github.com/spf13/pflag"
)

const VERSION = "0.1"

var (
	title = flag.String("title", "Picoblog", "Title for blog")
	list  = flag.String("list", "", "List of blog posts, sorted by display order")

	postTemplate = template.Must(template.New("post").Parse(
		`
<hr style="margin: 2em 0" />
<div>
<div style="text-align: right">
  <h3 id="{{ .Title }}" style="margin-bottom: .5em">{{ .Title }}</h3>
  <b>{{ .Timestamp.Format "January 2nd, 2006" }}</b>
</div>
{{ .Contents }}
</div>
`))

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
			fmt.Fprintf(os.Stderr, "error opening post list: %+v", err)
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

	renderHtml(os.Stdout, posts)
}

type post struct {
	Title     string
	Timestamp time.Time
	Contents  string
}

func (p *post) Render(w io.Writer) {
	htmlContents := blackfriday.Run([]byte(p.Contents))
	err := postTemplate.Execute(w, post{
		p.Title,
		p.Timestamp,
		string(htmlContents),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "error rendering post %s\n: %+v", p.Title, err)
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

func renderHtml(w io.Writer, posts []*post) {
	blogTemplate.Execute(w, struct{ Title string }{
		*title,
	})

	for _, p := range posts {
		fmt.Fprintf(os.Stderr, "rendering post %s\n", p.Title)
		p.Render(w)
	}
	fmt.Fprintf(w, "</body>")
}
