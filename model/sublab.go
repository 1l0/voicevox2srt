package model

import (
	"github.com/martinlindhe/subtitles"
)

type SubLab struct {
	SpeakerId string
	Subtitles *subtitles.Subtitle
	Lab       string
}
