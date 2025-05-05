package pen

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
)

func InitLog() {
	styles := log.DefaultStyles()
	log.SetFormatter(log.TextFormatter)
	log.SetStyles(styles)
	log.SetLevel(log.InfoLevel)
}

func Loader(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Prefix = suffix
	s.FinalMSG = ""
	s.Start()
	return s
}

func Complete(s *spinner.Spinner, msg string) {
	s.Stop()
	log.Info(msg)
}

func SpinWheel(start string, finish string) {
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)

	s.Prefix = start
	s.FinalMSG = ""
	s.Start()

	time.Sleep(1 * time.Second)
	s.Stop()
	log.Info(finish)
}

func SpinBar(start string, finish string) {
	s := spinner.New(spinner.CharSets[36], 120*time.Millisecond)

	s.Prefix = start
	s.FinalMSG = ""
	s.Start()

	time.Sleep(1 * time.Second)
	s.Stop()
	log.Info(finish)
}
