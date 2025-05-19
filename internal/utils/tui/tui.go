package tui

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jroimartin/gocui"

	"github.com/merzzzl/cisco-socks5/internal/service"
	"github.com/merzzzl/cisco-socks5/internal/utils/log"
)

type LogWriter struct {
	logs chan string
}

// Write interface for writing to a log.
func (l *LogWriter) Write(p []byte) (n int, err error) {
	l.logs <- string(p)

	return len(p), nil
}

// CreateTUI creates a TUI for the given service.
func CreateTUI(svc *service.Service, useFun bool) error {
	l := &LogWriter{logs: make(chan string, 100)}

	log.SetOutput(l)

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		return err
	}

	if useFun {
		go fun(g)
	} else {
		g.FgColor = gocui.Attribute(232)
	}

	g.BgColor = gocui.Attribute(235)

	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error {
		return layout(g, svc, l.logs)
	})

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		return err
	}

	return nil
}

func layout(g *gocui.Gui, svc *service.Service, logs <-chan string) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("logs", 0, 0, maxX-21, maxY-1); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Logs"

		go func() {
			for logMsg := range logs {
				g.Update(func(*gocui.Gui) error {
					fmt.Fprint(v, logMsg)

					lines := len(v.BufferLines()) - 1
					_, vy := v.Size()

					if lines > vy {
						ox, _ := v.Origin()

						if err := v.SetOrigin(ox, lines-vy); err != nil {
							return err
						}
					}

					return nil
				})
			}
		}()
	}

	if v, err := g.SetView("stats", maxX-20, 0, maxX-1, maxX-5); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Status"

		go func() {
			for range time.NewTicker(time.Millisecond * 100).C {
				g.Update(func(*gocui.Gui) error {
					v.Clear()

					fmt.Fprintf(v, "Cisco: %s\n", boolToStr(svc.GetState().CiscoConnected))
					fmt.Fprintf(v, "Filter:    %s\n", boolToStr(svc.GetState().PFDisabled))
					fmt.Fprintf(v, "Proxy: %s\n", boolToStr(svc.GetState().ProxyStarted))

					return nil
				})
			}
		}()
	}

	if v, err := g.SetView("uptime", maxX-20, maxY-3, maxX-1, maxY-1); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Uptime"
		start := time.Now()

		go func() {
			for range time.NewTicker(time.Millisecond * 1000).C {
				g.Update(func(*gocui.Gui) error {
					v.Clear()

					up := time.Unix(0, 0).UTC().Add(time.Since(start)).Format("15:04:05")

					fmt.Fprintf(v, "%8s %9s", "", up)

					return nil
				})
			}
		}()
	}

	return nil
}

func fColor() func() int {
	currentColor := 51
	color := func() int {
		if currentColor == 231 {
			currentColor = 51
		}

		currentColor++

		return currentColor
	}

	return color
}

func fArt() string {
	ts := []string{
		"⊂(◉‿◉)つ──",
		"( ✜︵ ✜ )─",
		"ʕっ •ᴥ•ʔっ─",
		"(｡◕‿‿◕｡)─",
		"(っ ´ω`c)♡",
		"(ʘ‿ʘ)╯────",
	}
	art := ts[rand.Intn(6)]

	return fmt.Sprintf("─%s─", art)
}

func boolToStr(b bool) string {
	if b {
		return colorize("OK", 10)
	}

	return colorize("FAIL", 9)
}

func colorize(s string, c int) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c, s)
}

func fun(g *gocui.Gui) {
	cl := fColor()

	v, err := g.SetView("fun", 0, -1, 11, 1)
	if !errors.Is(err, gocui.ErrUnknownView) {
		return
	}

	v.Frame = false

	_, _ = fmt.Fprint(v, fArt())

	for range time.NewTicker(time.Millisecond * 75).C {
		fgc := cl() + 1

		g.FgColor = gocui.Attribute(fgc)

		if _, err := g.View("fun"); err == nil {
			v.FgColor = gocui.Attribute(fgc)
		}

		if v, err := g.View("uptime"); err == nil {
			v.FgColor = gocui.Attribute(fgc)
		}
	}
}
