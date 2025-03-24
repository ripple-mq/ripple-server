package utils

import (
	"fmt"
	"math/rand"
)

const (
	minPort = 1024
	maxPort = 49150
)

func RandLocalAddr() string {
	randomNumber := rand.Intn(maxPort-minPort) + minPort
	return fmt.Sprintf(":%d", randomNumber)
}
