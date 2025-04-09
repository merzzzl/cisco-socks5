package ui

import (
	"context"
	"errors"
	"fmt"
	"github.com/jroimartin/gocui"
	"strings"
	"warp-server/pkg/controlloop"
	"warp-server/pkg/log"
)

type LogWriter struct {
	Logs chan string
}

// Write interface for writing to a log.
func (l *LogWriter) Write(p []byte) (n int, err error) {
	l.Logs <- string(p)

	return len(p), nil
}

// CreateTUI creates a TUI for the given service.
func CreateTUI(cancel context.CancelFunc, g *gocui.Gui, l *LogWriter, conditions <-chan []controlloop.Condition) error {

	g.FgColor = gocui.Attribute(232)
	g.BgColor = gocui.Attribute(235)

	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error {
		return layout(g, l.Logs, conditions)
	})

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
		cancel()
		return nil
	}); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		return err
	}

	return nil
}

func layout(g *gocui.Gui, logs <-chan string, conditions <-chan []controlloop.Condition) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("Logs", 0, maxY-15, maxX-21, maxY-1); err != nil {
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

	// 0, maxY-15, maxX-21, maxY-1
	if v, err := g.SetView("conditions", 0, 0, maxX-21, maxY-16); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Conditions"

		go func() {
			for c := range conditions {
				g.Update(func(*gocui.Gui) error {
					v.Clear()

					for _, currentCondition := range c {
						fmt.Fprintf(v, "%s %s %s %s \n",
							log.Colorize(currentCondition.Type, 7),
							log.Colorize(strings.ToUpper(currentCondition.Reason), 11),
							currentCondition.Status,
							currentCondition.Message,
						)
					}

					return nil
				})
			}
		}()
	}

	return nil
}
