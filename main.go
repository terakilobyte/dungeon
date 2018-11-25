package main

import (
	"fmt"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/terakilobyte/dungeon/entity"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Viewer Dungeon",
		Bounds: pixel.R(0, 0, 2048, 1280),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(true)

	pic, err := entity.LoadPicture("./assets/Fight_Background_test.png")
	if err != nil {
		panic(err)
	}

	heroOpts := &entity.Options{
		SpriteDir:    "./assets/_PNG/3_KNIGHT",
		AttackPower:  1.5,
		Health:       300,
		Name:         "hero",
		StartingV:    pixel.V(100, 320),
		Scaling:      0.12,
		CanMove:      true,
		CanCombat:    true,
		DefaultState: entity.Moving,
		Facing:       entity.FacingRight,
	}
	hero := heroOpts.New()
	enemies := entity.GenerateEnemies(5)
	last := time.Now()

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		// if time.Since(last) < 1*time.Second {
		// 	continue
		// }
		last = time.Now()

		background := pixel.NewSprite(pic, pic.Bounds())
		backgroundMat := pixel.IM.Moved(win.Bounds().Center())
		backgroundMat = backgroundMat.ScaledXY(win.Bounds().Center(), pixel.V(2, 2))

		// remember to update iterators here dumbass

		win.Clear(colornames.Skyblue)
		background.Draw(win, backgroundMat)
		if len(enemies) == 0 {
			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			centerWindow := win.Bounds().Center()
			centerWindow.X = centerWindow.X - 200
			basicTxt := text.New(centerWindow, basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "YOU WIN")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 7))
		}

		if hero != nil {
			newEnemies := make([]*entity.Entity, 0)
			hero.Update(win, dt)
			for _, e := range enemies {
				e.Update(win, dt)
				if e.State != entity.Remove {
					newEnemies = append(newEnemies, e)
				}
			}
			enemies = newEnemies
			hero.DetectCollision(enemies, win)
			// health text
			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			basicTxt := text.New(pixel.V(100, 895), basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "Health")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 3))

			// health bar
			imd := imdraw.New(nil)
			imd.Color = colornames.Red
			imd.EndShape = imdraw.RoundEndShape
			imd.Push(pixel.V(100, 885), pixel.V(100+hero.Health, 885))
			imd.Line(15)
			imd.Draw(win)
			if hero.Health <= 0 {
				hero = nil
			}
		}
		if hero == nil {
			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			centerWindow := win.Bounds().Center()
			centerWindow.X = centerWindow.X - 200
			basicTxt := text.New(centerWindow, basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "GAME OVER")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 7))
		}
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
