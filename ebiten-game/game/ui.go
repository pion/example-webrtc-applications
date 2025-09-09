// SPDX-FileCopyrightText: 2024 The Ebitengine Authors
// SPDX-License-Identifier: MIT

package main

import (
	"image"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
	g.logUpdated = true
}

func (g *game) logWindow(ctx *debugui.Context) {
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
			submitOpen := func() {
				g.isHost = true
				g.startConnection()
			}

			submitJoin := func() {
				g.isHost = false
				if g.logSubmitBuf == "" {
					return
				}
				g.lobbyID = g.logSubmitBuf
				g.logSubmitBuf = ""
				g.startConnection()
			}

			ctx.SetGridLayout([]int{-1, -1, -1, -1}, nil)
			ctx.Text("Lobby ID:")
			ctx.TextField(&g.logSubmitBuf).On(func() {
				if ebiten.IsKeyPressed(ebiten.KeyEnter) {
					submitJoin()
					ctx.SetTextFieldValue(g.logSubmitBuf)
				}
			})
			ctx.Button("Host Game").On(func() {
				submitOpen()
			})
			ctx.Button("Join").On(func() {
				submitJoin()
			})
		})
	})
}
