package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/terakilobyte/dungeon/chatbot"
	"github.com/terakilobyte/dungeon/entity"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func run() {
	go func() {
		chatbot.Run()
	}()
	chatTimer := time.NewTimer(2 * time.Second)
	<-chatTimer.C

	chat, err := chatbot.GetClient()

	if err != nil {
		fmt.Println(err)
	}
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

	hero, enemies, numEnemiesAlive := initGame(chat)

	last := time.Now()
	gameOver := false
	setScore := false
	gameResetTimer := 0
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if gameOver {
				gameResetTimer++
			}
		}
	}()
	numEnemiesDead := 0
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		background := pixel.NewSprite(pic, pic.Bounds())
		backgroundMat := pixel.IM.Moved(win.Bounds().Center())
		backgroundMat = backgroundMat.ScaledXY(win.Bounds().Center(), pixel.V(2, 2))

		win.Clear(colornames.Skyblue)
		background.Draw(win, backgroundMat)
		if gameOver && gameResetTimer >= 5 {
			gameOver = false
			gameResetTimer = 0
			hero.ResetHero(chat)
			hero.Health = rand.Float64()*150 + 50
			for _, e := range enemies {
				e.ResetEntity(false)
			}
			numEnemiesDead = 0
			heroName, err := getHeroName(chat)
			go func(chat *chatbot.ChatClient, hero *entity.Entity) {
				res := &bson.D{}
				elem := chat.Collection.FindOne(context.Background(), bson.D{{"user", hero.Name}})
				if err := elem.Decode(res); err != nil {
					fmt.Println(err, "error decoding document")
				}
				doc := res.Map()
				if losses, ok := doc["dungeonLosses"]; ok {
					hero.Health = hero.Health + float64(losses.(int64))
				}
				if wins, ok := doc["dungeonWins"]; ok {
					hero.Health = hero.Health + (float64(wins.(int64) * 2))
				}
			}(chat, hero)

			if err != nil {
				heroName = "hero"
			}
			hero.Name = heroName
			setScore = false
			numEnemiesAlive = 0
			for _, e := range enemies {
				if e.State != entity.Remove {
					numEnemiesAlive++
				}
			}
		} else if numEnemiesDead == numEnemiesAlive {
			if !setScore {
				go func(chat *chatbot.ChatClient, hero string) {
					opts := options.Update()
					opts.SetUpsert(true)
					filter := bson.D{{"user", hero}}
					update := bson.D{{"$inc", bson.D{{"dungeonWins", 1}}}}
					chat.Collection.UpdateOne(context.TODO(), filter, update, opts)
				}(chat, hero.Name)
				setScore = true
			}
			gameOver = true
			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			centerWindow := win.Bounds().Center()
			centerWindow.X = centerWindow.X - 200
			basicTxt := text.New(centerWindow, basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "YOU WIN")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 7))
		} else if !gameOver {
			hero.Update(win, dt)
			for _, e := range enemies {
				e.Update(win, dt)
				if e.State == entity.Remove && e.SetRemoved == false {
					e.SetRemoved = true
					numEnemiesDead++
				}
			}
			hero.DetectCollision(enemies, win)
			// health text
			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			basicTxt := text.New(pixel.V(100, 895), basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "Health")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 3))

			heroName := text.New(pixel.V(hero.PosX-50, hero.PosY+80), basicAtlas)
			_, err = fmt.Fprintln(heroName, hero.Name)
			if err != nil {
				fmt.Println(err)
			}

			heroName.Draw(win, pixel.IM.Scaled(heroName.Orig, 2))
			// health bar
			imd := imdraw.New(nil)
			imd.Color = colornames.Red
			imd.EndShape = imdraw.RoundEndShape
			imd.Push(pixel.V(100, 885), pixel.V(100+hero.Health, 885))
			imd.Line(15)
			imd.Draw(win)
			if hero.Health <= 0 {
				gameOver = true
			}
		} else if hero.Health <= 0 && gameOver {
			if !setScore {
				go func(chat *chatbot.ChatClient, hero string) {
					opts := options.Update()
					opts.SetUpsert(true)
					filter := bson.D{{"user", hero}}
					update := bson.D{{"$inc", bson.D{{"dungeonLosses", 1}}}}
					chat.Collection.UpdateOne(context.TODO(), filter, update, opts)
				}(chat, hero.Name)
			}
			setScore = true

			basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
			centerWindow := win.Bounds().Center()
			centerWindow.X = centerWindow.X - 200
			basicTxt := text.New(centerWindow, basicAtlas)
			_, err := fmt.Fprintln(basicTxt, "YOU LOSE")
			if err != nil {
				fmt.Println(err)
			}
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 7))
		}
		win.Update()
	}
}

func getHeroName(client *chatbot.ChatClient) (string, error) {
	users, err := client.Client.Userlist("swarmlogic")
	var heroName = ""
	if err != nil {
		return "", fmt.Errorf("hero")
	}
	heroName = users[rand.Intn(len(users))]
	return heroName, nil
}

func initGame(client *chatbot.ChatClient) (*entity.Entity, []*entity.Entity, int) {
	heroName, err := getHeroName(client)
	if err != nil {
		heroName = "hero"
	}

	heroOpts := &entity.Options{
		SpriteDir:    "./assets/_PNG/3_KNIGHT",
		AttackPower:  rand.Float64()*2 + 1,
		Health:       rand.Float64()*150 + 50,
		Name:         heroName,
		StartingV:    pixel.V(100, 320),
		Scaling:      0.12,
		CanMove:      true,
		CanCombat:    true,
		DefaultState: entity.Moving,
		Facing:       entity.FacingRight,
	}
	hero := heroOpts.New()
	go func(chat *chatbot.ChatClient, hero *entity.Entity) {
		res := &bson.D{}
		elem := chat.Collection.FindOne(context.Background(), bson.D{{"user", hero.Name}})
		if err := elem.Decode(res); err != nil {
			fmt.Println(err, "No Document Found", hero.Name)
		}
		doc := res.Map()
		dungeonRuns := 0.0
		if losses, ok := doc["dungeonLosses"]; ok {
			lossPool := float64(losses.(int64))
			dungeonRuns += lossPool
			hero.Health = hero.Health + lossPool*.25
		}
		if wins, ok := doc["dungeonWins"]; ok {
			winPool := float64(wins.(int64))
			dungeonRuns += winPool
			hero.Health = hero.Health + winPool*0.5
		}
		apAddition := math.Log10(dungeonRuns) / math.Log10(60.0)
		hero.AttackPower = hero.AttackPower + apAddition
	}(client, hero)
	enemies := entity.GenerateEnemies()
	numEnemiesAlive := 0
	for _, e := range enemies {
		if e.State != entity.Remove {
			numEnemiesAlive++
		}
	}
	return hero, enemies, numEnemiesAlive
}

func main() {
	pixelgl.Run(run)
}
