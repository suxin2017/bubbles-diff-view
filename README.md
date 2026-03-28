# bubbles-diff-view

A reusable diff view component for Bubble Tea applications, written in Go.

## Features

- Render line-based diffs in terminal UI.
- Colorized added, removed, and unchanged lines.
- Optional line numbers for left and right sides.
- Scroll support via Bubbles viewport.
- Built-in key bindings and help panel.

## Install

```bash
go get github.com/suxin2017/bubbles-diff-view
```

## Run Demo

```bash
go run ./cmd/demo
```

Or compare two files:

```bash
go run ./cmd/demo path/to/old.txt path/to/new.txt
```

## Charmtone Web Palette

Live page: <https://suxin2017.cc/bubbles-diff-view/>

Run the palette web server:

```bash
go run ./cmd/charmtone-web
```

Open <http://localhost:5173> to view color cards.

For real-time preview in VS Code:

1. Install one of the recommended extensions in [.vscode/extensions.json](.vscode/extensions.json).
2. Start the server with `go run ./cmd/charmtone-web`.
3. Open the URL in VS Code's Live Preview panel, or your browser.

## Component Usage

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/suxin2017/bubbles-diff-view/diffview"
)

func main() {
    m := diffview.New(diffview.Options{
        Title:           "Diff",
        LeftTitle:       "old",
        RightTitle:      "new",
        ShowLineNumbers: true,
        Width:           120,
        Height:          32,
    })

    m.SetDiffStrings("hello\nworld\n", "hello\nbubbles\nworld\n")

    tea.NewProgram(m).Run()
}
```

## Key Bindings

- `up` / `k`: move up
- `down` / `j`: move down
- `pgup` / `b`: page up
- `pgdn` / `f`: page down
- `g`: jump to top
- `G`: jump to bottom
- `?`: toggle help
- `q`: quit

## Publish To GitHub

1. Create a new repository on GitHub. Example name: `bubbles-diff-view`.
2. In local project, set the remote URL:

```bash
git init
git add .
git commit -m "feat: initial bubbles diff view component"
git branch -M main
git remote add origin git@github.com:suxin2017/bubbles-diff-view.git
git push -u origin main
```

3. If you use HTTPS:

```bash
git remote set-url origin https://github.com/suxin2017/bubbles-diff-view.git
git push -u origin main
```

4. Optional version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```
