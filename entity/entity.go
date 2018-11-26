package entity

import (
	"context"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/terakilobyte/dungeon/chatbot"

	"github.com/faiface/pixel"
)

// State represents the state of the entity
type State int

// FacingDirection is the direction the entity is facing
type FacingDirection int

const (
	// FacingLeft facing left
	FacingLeft FacingDirection = -1
	// FacingRight facing right
	FacingRight FacingDirection = 1
)

const (
	// Moving state
	Moving State = 0
	// Static state
	Static State = 1
	// Combat state
	Combat State = 2
	// Dead state
	Dead State = 3
	// Remove state. The entity will be removed from the game world
	Remove State = 4
)

// Options are the valid options to construct an Entity with
type Options struct {
	SpriteDir    string
	AttackPower  float64
	Health       float64
	Name         string
	StartingV    pixel.Vec
	Scaling      float64
	CanMove      bool
	CanCombat    bool
	DefaultState State
	Facing       FacingDirection
}

// Entity represents a game object
type Entity struct {
	walkFrames   []*pixel.Sprite
	idleFrames   []*pixel.Sprite
	attackFrames []*pixel.Sprite
	deadFrames   []*pixel.Sprite
	currentFrame *pixel.Sprite
	AttackPower  float64
	Health       float64
	Name         string

	matrix         pixel.Matrix
	startingMatrix pixel.Matrix
	canMove        bool
	canCombat      bool
	State          State
	defaultState   State
	PosX           float64
	PosY           float64
	facing         FacingDirection
	animationFrame int
	deathFrame     int
	SetRemoved     bool
	startingHealth float64
	startingV      pixel.Vec
}

// Update is called in the game loop
func (e *Entity) Update(win *pixelgl.Window, dt float64) {
	switch e.State {
	case Moving:
		e.PosX += dt * 100
		e.matrix = e.matrix.Moved(pixel.V(dt*100, 0))
		e.currentFrame = e.walkFrames[e.animationFrame%len(e.walkFrames)]
		e.currentFrame.Draw(win, e.matrix)
	case Static:
		e.currentFrame = e.idleFrames[e.animationFrame%len(e.idleFrames)]
		e.currentFrame.Draw(win, e.matrix)
	case Combat:
		e.currentFrame = e.attackFrames[e.animationFrame%len(e.attackFrames)]
		e.currentFrame.Draw(win, e.matrix)
	case Dead:
		e.currentFrame = e.deadFrames[e.deathFrame%len(e.deadFrames)]
		e.deathFrame++
		if e.deathFrame == len(e.deadFrames)+122 {
			e.State = Remove
			break
		}
		if e.deathFrame >= len(e.deadFrames)-1 {
			e.currentFrame = e.deadFrames[len(e.deadFrames)-1]
		}
		e.currentFrame.Draw(win, e.matrix)
	}
}

// DetectCollision determines if our hero has collided with an object
func (e *Entity) DetectCollision(entities []*Entity, win *pixelgl.Window) {
	mid := e.currentFrame.Frame().Center().ScaledXY(pixel.V(0.12, 0.12)).X / 2
	for _, o := range entities {
		if e.State == Dead || e.State == Remove {
			break
		}
		if o.State == Dead || o.State == Remove {
			continue
		}
		oMid := o.currentFrame.Frame().Center().ScaledXY(pixel.V(0.12, 0.12)).X / 2
		left1 := e.PosX - mid
		right1 := e.PosX + mid
		left2 := o.PosX - oMid
		right2 := o.PosX + oMid
		if left1 > right2 || left2 > right1 || !o.canCombat {
			e.State = e.defaultState
			continue
		}
		e.State = Combat
		o.State = Combat
		eAttack := rand.Float64() * e.AttackPower
		oAttack := rand.Float64() * o.AttackPower
		e.Health -= oAttack
		o.Health -= eAttack
		if e.Health <= 0 {
			e.State = Dead
			o.State = o.defaultState
			break
		}
		if o.Health <= 0 {
			o.State = Dead
			e.State = e.defaultState
			break
		}
		break
	}

}

// New creates a new Entity given supplied Options
func (o *Options) New() *Entity {

	walkingPath := path.Join(o.SpriteDir, "WALK")
	attackPath := path.Join(o.SpriteDir, "ATTACK")
	idlePath := path.Join(o.SpriteDir, "IDLE")
	deadPath := path.Join(o.SpriteDir, "DIE")
	walkFrames, err := gatherAssets(walkingPath)
	if err != nil {
		panic(err)
	}
	attackFrames, err := gatherAssets(attackPath)
	if err != nil {
		panic(err)
	}
	idleFrames, err := gatherAssets(idlePath)
	if err != nil {
		panic(err)
	}
	deadFrames, err := gatherAssets(deadPath)
	if err != nil {
		panic(err)
	}
	mat := pixel.IM
	mat = mat.ScaledXY(pixel.ZV, pixel.V(float64(o.Facing)*o.Scaling, o.Scaling))
	mat = mat.Moved(o.StartingV)
	e := &Entity{
		canCombat:      o.CanCombat,
		canMove:        o.CanMove,
		State:          o.DefaultState,
		defaultState:   o.DefaultState,
		matrix:         mat,
		walkFrames:     walkFrames,
		idleFrames:     idleFrames,
		attackFrames:   attackFrames,
		deadFrames:     deadFrames,
		PosX:           o.StartingV.X,
		PosY:           o.StartingV.Y,
		Health:         o.Health,
		AttackPower:    o.AttackPower,
		Name:           o.Name,
		startingMatrix: mat,
		startingHealth: o.Health,
		startingV:      o.StartingV,
	}
	ticker := time.NewTicker(33 * time.Millisecond)
	go func() {
		for range ticker.C {
			e.animationFrame++
		}
	}()
	return e
}

func gatherAssets(spriteDir string) ([]*pixel.Sprite, error) {
	pics, err := ioutil.ReadDir(spriteDir)
	if err != nil {
		return nil, err
	}
	frames := make([]*pixel.Sprite, len(pics))

	for i, f := range pics {
		pic, err := LoadPicture(path.Join(spriteDir, f.Name()))
		if err != nil {
			return nil, err
		}
		sprite := pixel.NewSprite(pic, pic.Bounds())
		frames[i] = sprite
	}
	return frames, nil
}

func (hero *Entity) ResetHero(chat *chatbot.ChatClient) {
	hero.ResetEntity(true)
	go func(chat *chatbot.ChatClient, hero *Entity) {
		res := &bson.D{}
		elem := chat.Collection.FindOne(context.Background(), bson.D{{"user", hero.Name}})
		if err := elem.Decode(res); err != nil {
			fmt.Println(err, "No document found", hero.Name)
		}
		doc := res.Map()
		dungeonRuns := 0.0
		if losses, ok := doc["dungeonLosses"]; ok {
			lossPool := float64(losses.(int64))
			dungeonRuns += lossPool
			hero.Health = hero.Health + lossPool*.5
		}
		if wins, ok := doc["dungeonWins"]; ok {
			winPool := float64(wins.(int64))
			dungeonRuns += winPool
			hero.Health = hero.Health + winPool*.25
		}
		apAddition := math.Log10(dungeonRuns) / math.Log10(80.0)

		hero.AttackPower = hero.AttackPower + apAddition

	}(chat, hero)

}

// ResetEntity resets the entity
func (e *Entity) ResetEntity(force bool) {
	e.matrix = e.startingMatrix
	e.State = e.defaultState
	if force || rand.Float64() < .5 {
		e.State = e.defaultState
		e.SetRemoved = false
	} else {
		e.State = Remove
		e.SetRemoved = true
	}
	e.Health = rand.Float64()*100 + 50
	e.PosX, e.PosY = e.startingV.XY()
	e.deathFrame = 0
	e.AttackPower = rand.Float64() * 3
}

func LoadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
