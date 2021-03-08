package utils

import (
	"math/rand"
	"time"
)

var colors = []string{
	"gray",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
}

// GetRandomColor get random color
func GetRandomColor() string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(colors))
	return colors[index]
}