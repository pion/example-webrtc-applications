// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"image"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
	g.logUpdated = true
}

func (g *Game) logWindow(ctx *debugui.Context) {
	ctx.Window("Log Window", image.Rect(350, 40, 650, 290), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{-1}, []int{-1, 0})
		ctx.Panel(func(layout debugui.ContainerLayout) {
			ctx.SetGridLayout([]int{-1}, []int{-1})
			ctx.Text(g.logBuf)
			if g.logUpdated {
				ctx.SetScroll(image.Pt(layout.ScrollOffset.X, layout.ContentSize.Y))
				g.logUpdated = false
			}
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			submit_open := func() {
				g.isHost = true
				startConnection(g)
			}

			submit_join := func() {
				g.isHost = false
				if g.logSubmitBuf == "" {
					return
				}
				g.lobby_id = g.logSubmitBuf
				g.logSubmitBuf = ""
				startConnection(g)
			}

			ctx.SetGridLayout([]int{-3, -1, -1}, nil)
			ctx.TextField(&g.logSubmitBuf).On(func() {
				if ebiten.IsKeyPressed(ebiten.KeyEnter) {
					submit_join()
					ctx.SetTextFieldValue(g.logSubmitBuf)
				}
			})
			ctx.Button("Open").On(func() {
				submit_open()
			})
			ctx.Button("Join").On(func() {
				submit_join()
			})
		})
	})
}
