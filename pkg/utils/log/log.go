package log

import (
	"github.com/charmbracelet/log"
)

func InitLog() {
	styles := log.DefaultStyles()
	log.SetFormatter(log.TextFormatter)
	log.SetStyles(styles)
	log.SetLevel(log.InfoLevel)
}
