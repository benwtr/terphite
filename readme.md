# Terphite

A terminal [Graphite](http://graphite.readthedocs.org/) browser, loosely
based on Graphite Composer. Browse the metrics tree, build a graph, and
save sets of graphs as dashboards you can view together as a grid.

Written in Go using [bubbletea](https://github.com/charmbracelet/bubbletea),
[lipgloss](https://github.com/charmbracelet/lipgloss), and
[bubbles](https://github.com/charmbracelet/bubbles).

> **This is 100% vibe coded.** Every line of the Go rewrite was written by an
> LLM. It builds, the tests pass, and it's been smoke tested against a mock
> Graphite server — but it has never run against a real production Graphite
> instance, and no human has reviewed the code line by line. Treat it
> accordingly.

### Install

```
go install github.com/benwtr/terphite/cmd/terphite@latest
```

Or download a prebuilt binary from the
[releases page](https://github.com/benwtr/terphite/releases).

### Usage

```
terphite http://your.graphite.com:1234
```

Credentials can be embedded in the URL (`http://user:pass@host:1234`) or
supplied via the `GRAPHITE_USER` / `GRAPHITE_PASS` environment variables.

### Keys

Composer view:

| Key | Action |
| --- | --- |
| `↑`/`↓`, `enter` | move the metrics tree cursor, select a metric or expand/collapse a branch |
| `[` / `]` | decrease / increase the time range by 1 minute |
| `{` / `}` | decrease / increase the time range by 1 hour |
| `t` | set a relative "from" time (e.g. `-1d12h`) |
| `m` | open the selected-metrics popup (`ctrl+a` add, `ctrl+d` delete, `enter` edit, `esc` close) |
| `i` | set the autorefresh interval, in seconds |
| `a` | toggle autorefresh |
| `x` | set max datapoints (`0` = unlimited) |
| `o` | open the current graph in a browser |
| `c` | copy the current graph's URL to the clipboard |
| `S` | save the current graph as a panel on a dashboard |
| `D` | open a saved dashboard |
| `g` | cycle graph style: line / area / stacked |
| `I` | toggle graphical mode: off / iterm2 / kitty (off by default) |
| `l` | redraw the screen |
| `?` | collapse/expand the help bar |
| `q` / `ctrl+c` | quit |

Charts render with braille sub-character resolution for smooth connected
lines. `g` cycles between plain lines, independent filled areas per series,
and a cumulative stacked area (like Graphite's `areaMode=stacked`). Each
saved dashboard panel remembers its own graph style.

Dashboard view (a grid of saved graphs):

| Key | Action |
| --- | --- |
| `←`/`→`/`↑`/`↓` | move focus between panels |
| `enter` | load the focused panel back into the composer view for editing |
| `ctrl+d` | remove the focused panel |
| `a` | toggle autorefresh for all panels |
| `esc` | back to composer view |
| `q` / `ctrl+c` | quit |

Dashboards are saved as JSON files under `$XDG_CONFIG_HOME/terphite/dashboards`
(typically `~/.config/terphite/dashboards` on Linux and
`~/Library/Application Support/terphite/dashboards` on macOS).

### Graphical mode

In terminals that support inline images — iTerm2, WezTerm, Kitty, Ghostty —
terphite can display Graphite's own rendered PNGs instead of ASCII charts,
giving you graphite-web's real axis labels, legends, and gridlines.

It is **off by default**, since whether it works depends on terminal support
that can't be detected reliably (tmux and SSH in particular don't always
relay the escape sequences). Press `I` to turn it on: the first press picks
whichever protocol your terminal looks like it supports, and pressing again
cycles through the others in case that guess is wrong.

The image is painted as an overlay on top of a blank region reserved in the
layout, rather than embedded in it — an image escape sequence is a single
line of text but many rows on screen, and embedding one directly makes
everything below it slide off the bottom of the terminal.

### Developing

```
git clone https://github.com/benwtr/terphite.git
cd terphite
go build ./...
go test ./...
go run ./cmd/terphite http://your.graphite.com:1234
```
