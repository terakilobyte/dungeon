package entity

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
)

// GenerateEnemies seeds the map with enemies
func GenerateEnemies() []*Entity {
	enemies := 10
	genFuncs := []func(float64) *Entity{
		GenerateOrk, GenerateTroll, GenerateOrk, GenerateKnight, GenerateTroll,
		GenerateOrk, GenerateOrk, GenerateTroll, GenerateTroll, GenerateKnight,
	}
	rand.Seed(time.Now().Unix())
	enemySlice := make([]*Entity, enemies)
	for i := 0; i < enemies; i++ {
		enemySlice[i] = genFuncs[rand.Intn(len(genFuncs))](float64(i))
	}
	return enemySlice
}

func GenerateOrk(idx float64) *Entity {
	monsters := []string{"1_ORK", "2_ORK", "3_ORK"}
	sprite := monsters[rand.Intn(len(monsters))]
	enemyOpts := &Options{
		SpriteDir:    "./assets/_PNG/" + sprite,
		AttackPower:  rand.Float64() * 2,
		Health:       rand.Float64() * 100,
		Name:         sprite,
		StartingV:    pixel.V((2000/10)*idx+200, 325),
		Scaling:      0.10,
		CanMove:      true,
		CanCombat:    true,
		DefaultState: Static,
		Facing:       FacingLeft,
	}
	enemy := enemyOpts.New()
	if rand.Float64() < .5 {
		enemy.State = Remove
		enemy.SetRemoved = true
	}
	return enemy
}

func GenerateTroll(idx float64) *Entity {
	monsters := []string{"1_TROLL", "2_TROLL", "3_TROLL"}
	sprite := monsters[rand.Intn(len(monsters))]
	enemyOpts := &Options{
		SpriteDir:    "./assets/_PNG/" + sprite,
		AttackPower:  rand.Float64() * 2,
		Health:       rand.Float64() * 100,
		Name:         sprite,
		StartingV:    pixel.V((2000/15)*idx+200, 325),
		Scaling:      0.20,
		CanMove:      true,
		CanCombat:    true,
		DefaultState: Static,
		Facing:       FacingLeft,
	}
	enemy := enemyOpts.New()
	if rand.Float64() < .5 {
		enemy.State = Remove
		enemy.SetRemoved = true
	}
	return enemy
}

func GenerateKnight(idx float64) *Entity {
	monsters := []string{"1_KNIGHT", "2_KNIGHT", "3_KNIGHT"}
	sprite := monsters[rand.Intn(len(monsters))]
	enemyOpts := &Options{
		SpriteDir:    "./assets/_PNG/" + sprite,
		AttackPower:  rand.Float64() * 2,
		Health:       rand.Float64() * 100,
		Name:         sprite,
		StartingV:    pixel.V((2000/15)*idx+200, 325),
		Scaling:      0.20,
		CanMove:      true,
		CanCombat:    true,
		DefaultState: Static,
		Facing:       FacingLeft,
	}
	enemy := enemyOpts.New()
	if rand.Float64() < .5 {
		enemy.State = Remove
		enemy.SetRemoved = true
	}
	return enemy
}
