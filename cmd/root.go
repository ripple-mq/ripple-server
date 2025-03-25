package cmd

import (
	"fmt"
	"math/rand"

	"os"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/lighthouse"
	"github.com/ripple-mq/ripple-server/internal/lighthouse/utils"
	"github.com/ripple-mq/ripple-server/pkg/utils/config"
)

const (
	minPort = 1024
	maxPort = 49150
)

func RandLocalAddr() string {
	randomNumber := rand.Intn(maxPort-minPort) + minPort
	return fmt.Sprintf(":%d", randomNumber)
}

func Execute() {
	cfg := config.Conf

	log.Info(cfg)

	lh := lighthouse.GetLightHouse()

	addr := RandLocalAddr()

	lh.StartElectLoop(utils.Path{Cmp: []string{"topics", "topic-0", "bucket-0"}}, addr, func(path utils.Path) {})

	// These functions demonstrate two separate checks to detect if the code is being
	// run inside a docker container in debug mode, or production mode!
	//
	// Note: Valid only for docker containers generated using the Makefile command
	FirstCheck()
	SecondCheck()

	select {}

}

func FirstCheck() bool {
	/*
	 * The `debug_mode` environment variable exists only in debug builds, likewise,
	 * `production_mode` variable exists selectively in production builds - use the
	 * existence of these variables to detect container build type (and not values)
	 *
	 * Exactly one of these - `production_mode` or `debug_mode` - is **guaranteed** to
	 * exist for docker builds generated using the Makefile commands!
	 */

	if _, ok := os.LookupEnv("production_mode"); ok {
		fmt.Println("(Check 01): Running in `production` mode!")
		return true
	} else if _, ok := os.LookupEnv("debug_mode"); ok {
		// Could also use a simple `else` statement (above) for docker builds!
		fmt.Println("(Check 01): Running in `debug` mode!")
		return true
	}
	fmt.Println("\nP.S. Try running a build generated with the Makefile :)")
	return false
}

func SecondCheck() bool {
	/*
	 * There's also a central `__BUILD_MODE__` variable for a dirty checks -- guaranteed
	 * to exist for docker containers generated by the Makefile commands!
	 * The variable will have a value of `production` or `debug` (no capitalization)
	 *
	 * Note: Relates to docker builds generated using the Makefile
	 */

	value := os.Getenv("__BUILD_MODE__")

	// Yes, this if/else could have been written better
	switch value {
	case "production":
		fmt.Println("(Check 02): Running in `production` mode!")
		return true

	case "debug":
		fmt.Println("(Check 02): Running in `debug` mode!")
		return true

	default:
		// Flow ends up here for non-docker builds, or docker builds generated directly
		fmt.Println("Non-makefile build detected :(")
		return false
	}
}
