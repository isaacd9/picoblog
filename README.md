# Picoblog

Read more at [blog.isaac.diamonds](https://blog.isaac.diamonds).

Picoblog is a minimal static blog creator. It's heavily inspired by
[picofeed](https://github.com/seenaburns/picofeed) and usees the same styling.
It looks nice, it's dead simple to use, and it gets out of your way. What's not
to like, dude.

Things you don't need with picoblog (or picofeed):

- An account
- A subscription
- Any state at all

Honestly it's like a fancy markdown compiler.

```
  Examples:
	picoblog first.md second.md
	picoblog --list file.txt
```

```sh
# Use whatever click to open your terminal supports, like cmd+double click in OSX's Terminal.app
./picooblog blog-posts.txt
```

<p align="center">
      <img alt="picofeed local browser rss" src="https://i.imgur.com/2HQcHYF.jpg"/>
</p>



#### Install

From source, with go 1.11 just run `go build`

Or there are precompiled binaries in the [releases page](https://github.com/isaacd9/picoblog/releases/latest)


#### Other

Picoblog is built on top of [blackfriday](https://github.com/russross/blackfriday)
