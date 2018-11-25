package entity

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
)

// GenerateEnemies seeds the map with enemies
func GenerateEnemies(enemies int) []*Entity {
	monsters := []string{"1_ORK", "2_ORK", "3_ORK", "1_TROLL", "1_TROLL", "2_TROLL", "3_TROLL"}
	rand.Seed(time.Now().Unix())
	enemySlice := make([]*Entity, enemies)
	for i := 0; i < enemies; i++ {
		sprite := monsters[rand.Intn(len(monsters))]
		enemyOpts := &Options{
			SpriteDir:    "./assets/_PNG/" + sprite,
			AttackPower:  rand.Float64() * 2,
			Health:       rand.Float64() * 100,
			Name:         sprite,
			StartingV:    pixel.V(rand.Float64()*1800+200, 325),
			Scaling:      0.12,
			CanMove:      true,
			CanCombat:    true,
			DefaultState: Static,
			Facing:       FacingLeft,
		}
		enemySlice[i] = enemyOpts.New()
	}
	return enemySlice
}
