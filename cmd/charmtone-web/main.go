package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/x/exp/charmtone"
)

type swatch struct {
  Key       string `json:"key"`
	Name      string `json:"name"`
	Label     string `json:"label"`
	Hex       string `json:"hex"`
	RGB       string `json:"rgb"`
	Group     string `json:"group"`
	HueFamily string `json:"hueFamily"`
	Tone      string `json:"tone"`
	Safe256   bool   `json:"safe256"`
	VSCode    string `json:"vscodeTerminal"`
	IDEA      string `json:"ideaTerminal"`
}

func main() {
	staticMode := flag.Bool("static", false, "write a standalone static HTML palette file and exit")
	staticFile := flag.String("out", "./charmtone-palette.html", "output file path used with -static")
	flag.Parse()

	if *staticMode {
		if err := writeStaticHTML(*staticFile, palette()); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote static palette: %s\n", *staticFile)
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
	})

	mux.HandleFunc("/api/colors", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(palette()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5173"
	}

	addr := ":" + port
	fmt.Printf("charmtone web palette: http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func palette() []swatch {
	out := make([]swatch, 0, int(charmtone.NeueZinc-charmtone.Cumin)+1)
	for k := charmtone.Cumin; k <= charmtone.NeueZinc; k++ {
		r, g, b := rgb8(k)
		hex := k.Hex()
		out = append(out, swatch{
      Key:       keyIdent(k),
			Name:      k.String(),
			Label:     k.String(),
			Hex:       hex,
			RGB:       fmt.Sprintf("rgb(%d, %d, %d)", r, g, b),
			Group:     colorGroup(k),
			HueFamily: hueFamily(r, g, b),
			Tone:      toneLabel(r, g, b),
			Safe256:   isXterm256Hex(hex),
			VSCode:    "yes (truecolor)",
			IDEA:      "yes (truecolor)",
		})
	}
	return out
}

func keyIdent(k charmtone.Key) string {
  switch k {
  case charmtone.Mochi:
    return "Mochi"
  case charmtone.NeueGuac:
    return "NeueGuac"
  case charmtone.NeueZinc:
    return "NeueZinc"
  default:
    return strings.ReplaceAll(k.String(), " ", "")
  }
}

func writeStaticHTML(path string, colors []swatch) error {
	data, err := json.Marshal(colors)
	if err != nil {
		return err
	}
	html := strings.ReplaceAll(staticHTMLTemplate, "__COLORS_JSON__", string(data))
	return os.WriteFile(path, []byte(html), 0644)
}

func isXterm256Hex(hex string) bool {
	_, ok := xterm256HexSet[strings.ToUpper(hex)]
	return ok
}

func hexFromRGB(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func buildXterm256HexSet() map[string]struct{} {
	set := make(map[string]struct{}, 256)

	standard := [16][3]int{
		{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
		{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
		{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
		{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
	}
	for _, c := range standard {
		set[hexFromRGB(c[0], c[1], c[2])] = struct{}{}
	}

	vals := []int{0, 95, 135, 175, 215, 255}
	for _, r := range vals {
		for _, g := range vals {
			for _, b := range vals {
				set[hexFromRGB(r, g, b)] = struct{}{}
			}
		}
	}

	for i := 232; i <= 255; i++ {
		v := 8 + (i-232)*10
		set[hexFromRGB(v, v, v)] = struct{}{}
	}

	return set
}

var xterm256HexSet = buildXterm256HexSet()

func rgb8(c color.Color) (r, g, b uint8) {
	r16, g16, b16, _ := c.RGBA()
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8)
}

func colorGroup(k charmtone.Key) string {
	switch k {
	case charmtone.Pickle, charmtone.Gator, charmtone.Spinach:
		return "Diff additions"
	case charmtone.Pom, charmtone.Steak, charmtone.Toast:
		return "Diff deletions"
	case charmtone.NeueGuac, charmtone.NeueZinc:
		return "Provisional"
	}

	if k.IsPrimary() {
		return "Primary"
	}
	if k.IsSecondary() {
		return "Secondary"
	}
	if k.IsTertiary() {
		return "Tertiary"
	}
	return "Core"
}

func hueFamily(r, g, b uint8) string {
	h, s, _ := rgbToHsl(r, g, b)
	if s < 0.08 {
		return "Neutral"
	}

	switch {
	case h < 15 || h >= 345:
		return "Red"
	case h < 45:
		return "Orange"
	case h < 70:
		return "Yellow"
	case h < 165:
		return "Green"
	case h < 200:
		return "Cyan"
	case h < 255:
		return "Blue"
	case h < 290:
		return "Purple"
	case h < 345:
		return "Pink"
	default:
		return "Unknown"
	}
}

func toneLabel(r, g, b uint8) string {
	_, s, l := rgbToHsl(r, g, b)
	bucket := "mid"
	if l < 0.22 {
		bucket = "deep"
	} else if l < 0.42 {
		bucket = "dark"
	} else if l < 0.7 {
		bucket = "mid"
	} else if l < 0.86 {
		bucket = "light"
	} else {
		bucket = "very light"
	}

	if s < 0.1 {
		return bucket + " muted"
	}
	if s > 0.75 {
		return bucket + " vivid"
	}
	return bucket + " balanced"
}

func rgbToHsl(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	maxV := math.Max(rf, math.Max(gf, bf))
	minV := math.Min(rf, math.Min(gf, bf))
	l = (maxV + minV) / 2.0

	if maxV == minV {
		return 0, 0, l
	}

	delta := maxV - minV
	if l > 0.5 {
		s = delta / (2.0 - maxV - minV)
	} else {
		s = delta / (maxV + minV)
	}

	switch maxV {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/delta + 2
	case bf:
		h = (rf-gf)/delta + 4
	}

	h *= 60
	return h, s, l
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Charmtone Web Palette</title>
  <style>
    :root {
      --bg: #f4f4f1;
      --ink: #212121;
      --card: #fffffd;
      --muted: #6b6b6b;
      --ring: #d7d7cf;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      font-family: "IBM Plex Sans", "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at 0% 0%, #e9efe7 0, transparent 35%),
        radial-gradient(circle at 100% 0%, #f3e5de 0, transparent 28%),
        radial-gradient(circle at 100% 100%, #e6ecf4 0, transparent 35%),
        var(--bg);
      min-height: 100vh;
    }

    .wrap {
      width: min(980px, 92vw);
      margin: 3rem auto;
    }

    h1 {
      margin: 0;
      font-size: clamp(1.8rem, 3vw, 2.8rem);
      letter-spacing: 0.02em;
      font-weight: 700;
    }

    p {
      margin: 0.5rem 0 0.9rem;
      color: var(--muted);
    }

    .toolbar {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
      margin: 0.9rem 0 1.2rem;
    }

    .toolbar input,
    .toolbar select,
    .toolbar button {
      border: 1px solid var(--ring);
      border-radius: 999px;
      background: #fff;
      color: var(--ink);
      padding: 0.45rem 0.85rem;
      font-family: inherit;
      font-size: 0.92rem;
    }

    .toolbar input {
      min-width: 230px;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      gap: 14px;
    }

    .swatch {
      border: 1px solid var(--ring);
      border-radius: 16px;
      background: var(--card);
      overflow: hidden;
      box-shadow: 0 8px 28px rgba(0,0,0,0.06);
      transform: translateY(8px);
      opacity: 0;
      animation: rise .45s ease forwards;
      cursor: copy;
    }

    @keyframes rise {
      to {
        transform: translateY(0);
        opacity: 1;
      }
    }

    .chip {
      height: 86px;
      border-bottom: 1px solid var(--ring);
    }

    .meta {
      padding: 10px 12px 12px;
    }

    .name {
      font-weight: 700;
      font-size: 0.98rem;
    }

    .hex {
      color: var(--muted);
      font-size: 0.87rem;
      margin-top: 4px;
      font-family: "IBM Plex Mono", Consolas, monospace;
    }

    .meta2 {
      margin-top: 8px;
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .tag {
      border: 1px solid var(--ring);
      border-radius: 999px;
      padding: 0.12rem 0.5rem;
      font-size: 0.74rem;
      color: #333;
      background: #fafaf7;
      white-space: nowrap;
    }

    .hint {
      margin-top: 1rem;
      font-size: 0.9rem;
      color: var(--muted);
    }

    button {
      border: 1px solid var(--ring);
      border-radius: 999px;
      background: #fff;
      color: var(--ink);
      padding: 0.4rem 0.85rem;
      cursor: pointer;
      font-family: inherit;
    }

    button:hover {
      background: #f7f7f2;
    }
  </style>
</head>
<body>
  <main class="wrap">
    <h1>Charmtone Palette</h1>
    <p>All colors are generated from Go charmtone constants.</p>
    <p>VS Code terminal and JetBrains IDEA terminal are marked as supported under truecolor mode.</p>
    <div class="toolbar">
      <input id="q" type="search" placeholder="Search name / hex / group" />
      <select id="group"></select>
      <button id="reload" type="button">Reload</button>
    </div>
    <div id="grid" class="grid"></div>
    <div id="stat" class="hint"></div>
    <div id="copyTip" class="hint"></div>
    <div class="hint">Tip: use VS Code Live Preview / Live Server for auto refresh while editing.</div>
  </main>

  <script>
    const grid = document.getElementById('grid');
    const stat = document.getElementById('stat');
    const copyTip = document.getElementById('copyTip');
    const reloadBtn = document.getElementById('reload');
    const qInput = document.getElementById('q');
    const groupSelect = document.getElementById('group');
    let allColors = [];

    function esc(s) {
      return String(s)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
    }

    function renderGroups(colors) {
      const groups = ['All', ...new Set(colors.map(c => c.group))];
      groupSelect.innerHTML = groups
        .map(g => '<option value="' + esc(g) + '">' + esc(g) + '</option>')
        .join('');
    }

    function applyFilters() {
      const q = qInput.value.trim().toLowerCase();
      const group = groupSelect.value;

      const filtered = allColors.filter(item => {
        if (group && group !== 'All' && item.group !== group) {
          return false;
        }
        if (!q) {
          return true;
        }

        return [item.name, item.label, item.hex, item.rgb, item.group, item.hueFamily, item.tone]
          .join(' ')
          .toLowerCase()
          .includes(q);
      });

      renderGrid(filtered);
      stat.textContent = 'Showing ' + filtered.length + ' / ' + allColors.length + ' colors';
    }

    function renderGrid(colors) {
      grid.innerHTML = '';

      colors.forEach((item, i) => {
        const card = document.createElement('article');
        card.className = 'swatch';
        card.tabIndex = 0;
        card.setAttribute('role', 'button');
        card.setAttribute('aria-label', 'Copy charmtone token for ' + item.name);
        card.style.animationDelay = (i * 35) + 'ms';

        card.innerHTML =
          '<div class="chip" style="background:' + item.hex + '"></div>' +
          '<div class="meta">' +
            '<div class="name">' + esc(item.name) + '</div>' +
            '<div class="hex">' + esc(item.hex) + ' / ' + esc(item.rgb) + '</div>' +
            '<div class="meta2">' +
              '<span class="tag">' + esc(item.group) + '</span>' +
              '<span class="tag">' + esc(item.hueFamily) + '</span>' +
              '<span class="tag">' + esc(item.tone) + '</span>' +
              '<span class="tag">VS Code: ' + esc(item.vscodeTerminal) + '</span>' +
              '<span class="tag">IDEA: ' + esc(item.ideaTerminal) + '</span>' +
              '<span class="tag">xterm256: ' + (item.safe256 ? 'yes' : 'no') + '</span>' +
            '</div>' +
          '</div>';

        const copyHandler = () => {
          copyColorRef(item).catch(err => {
            showCopyTip('Copy failed: ' + err);
          });
        };
        card.addEventListener('click', copyHandler);
        card.addEventListener('keydown', (e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            copyHandler();
          }
        });

        grid.appendChild(card);
      });
    }

    function showCopyTip(text) {
      copyTip.textContent = text;
      clearTimeout(showCopyTip._timer);
      showCopyTip._timer = setTimeout(() => {
        copyTip.textContent = '';
      }, 1800);
    }

    async function copyText(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        try {
          await navigator.clipboard.writeText(text);
          return;
        } catch (err) {
          // Some hosts (for example embedded previews) block Clipboard API.
          // Fall back to document.execCommand("copy").
        }
      }

      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.focus();
      ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      if (!ok) {
        throw new Error('clipboard unavailable');
      }
    }

    async function copyColorRef(item) {
      const text = 'charmtone.' + item.key;
      await copyText(text);
      showCopyTip('Copied ' + text);
    }

    async function loadColors() {
      const res = await fetch('/api/colors', { cache: 'no-store' });
      allColors = await res.json();
      renderGroups(allColors);
      applyFilters();
    }

    reloadBtn.addEventListener('click', () => {
      loadColors().catch(console.error);
    });
    qInput.addEventListener('input', applyFilters);
    groupSelect.addEventListener('change', applyFilters);

    loadColors().catch(console.error);
  </script>
</body>
</html>
`

const staticHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Charmtone Static Palette</title>
  <style>
    :root {
      --bg: #f4f4f1;
      --ink: #212121;
      --card: #fffffd;
      --muted: #6b6b6b;
      --ring: #d7d7cf;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "IBM Plex Sans", "Segoe UI", sans-serif;
      color: var(--ink);
      background: var(--bg);
      min-height: 100vh;
    }
    .wrap { width: min(980px, 92vw); margin: 2rem auto; }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
      gap: 12px;
    }
    .swatch {
      border: 1px solid var(--ring);
      border-radius: 14px;
      background: var(--card);
      overflow: hidden;
      cursor: copy;
    }
    .chip { height: 80px; }
    .meta { padding: 10px; }
    .name { font-weight: 700; }
    .hex { color: var(--muted); font-family: "IBM Plex Mono", Consolas, monospace; font-size: 0.85rem; margin-top: 4px; }
    .tags { margin-top: 8px; display: flex; flex-wrap: wrap; gap: 6px; }
    .tag {
      border: 1px solid var(--ring);
      border-radius: 999px;
      padding: 0.1rem 0.45rem;
      font-size: 0.74rem;
      background: #fafaf7;
    }
  </style>
</head>
<body>
  <main class="wrap">
    <h1>Charmtone Static Palette</h1>
    <p>Generated by Go at build time from charmtone constants.</p>
    <p>VS Code terminal and JetBrains IDEA terminal are marked as supported under truecolor mode.</p>
    <div id="grid" class="grid"></div>
    <p id="copyTip"></p>
  </main>
  <script>
    const data = __COLORS_JSON__;
    const grid = document.getElementById('grid');
    const copyTip = document.getElementById('copyTip');

    function showCopyTip(text) {
      copyTip.textContent = text;
      clearTimeout(showCopyTip._timer);
      showCopyTip._timer = setTimeout(() => {
        copyTip.textContent = '';
      }, 1800);
    }

    async function copyText(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        try {
          await navigator.clipboard.writeText(text);
          return;
        } catch (err) {
          // Some hosts (for example embedded previews) block Clipboard API.
          // Fall back to document.execCommand("copy").
        }
      }
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.focus();
      ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      if (!ok) {
        throw new Error('clipboard unavailable');
      }
    }

    async function copyColorRef(item) {
      const text = 'charmtone.' + item.key;
      await copyText(text);
      showCopyTip('Copied ' + text);
    }

    for (const item of data) {
      const card = document.createElement('article');
      card.className = 'swatch';
      card.tabIndex = 0;
      card.setAttribute('role', 'button');
      card.setAttribute('aria-label', 'Copy charmtone token for ' + item.name);
      card.innerHTML =
        '<div class="chip" style="background:' + item.hex + '"></div>' +
        '<div class="meta">' +
          '<div class="name">' + item.name + '</div>' +
          '<div class="hex">' + item.hex + ' / ' + item.rgb + '</div>' +
          '<div class="tags">' +
            '<span class="tag">' + item.group + '</span>' +
            '<span class="tag">' + item.hueFamily + '</span>' +
            '<span class="tag">' + item.tone + '</span>' +
            '<span class="tag">VS Code: ' + item.vscodeTerminal + '</span>' +
            '<span class="tag">IDEA: ' + item.ideaTerminal + '</span>' +
            '<span class="tag">xterm256: ' + (item.safe256 ? 'yes' : 'no') + '</span>' +
          '</div>' +
        '</div>';
      const copyHandler = () => {
        copyColorRef(item).catch(err => {
          showCopyTip('Copy failed: ' + err);
        });
      };
      card.addEventListener('click', copyHandler);
      card.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          copyHandler();
        }
      });
      grid.appendChild(card);
    }
  </script>
</body>
</html>
`
