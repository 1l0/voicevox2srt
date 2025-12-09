package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/1l0/vv2srt/construct"
)

const (
	//  adjustments for the sync problem
	adjustmentVoicevoxDefault = -30000000.0
	adjustmentAivisDefault    = 60000000.0
)

var (
	outputFilename string
	isAivis        bool = false
	lab            bool = false

	// adjustment for the sync problem
	adjustmentNanoSec = 0.0
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&outputFilename, "o", "", "Output file path.")
	adjReadme := fmt.Sprintf(
		"Duration adjustment in nano seconds per item. If you don't specify, following are set:\n\tVOICEVOX:\t%0.1f\n\tAivisSpeech:\t%0.1f",
		adjustmentVoicevoxDefault,
		adjustmentAivisDefault,
	)
	flag.Float64Var(&adjustmentNanoSec, "d", 0, adjReadme)
	flag.BoolVar(&lab, "lab", false, "Generate also .lab file.")
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln(fmt.Errorf("missing project file path (.vvproj or .aisp)"))
	}
	projPath := args[0]
	exp, err := regexp.Compile(`.+\.(aisp|vvproj)$`)
	if err != nil {
		log.Fatalln(err)
	}
	m := exp.FindStringSubmatch(projPath)
	if len(m) < 2 {
		log.Fatalln(fmt.Errorf("unsupported project file"))
	}
	if outputFilename == "" {
		outputFilename = m[0]
	}
	if m[1] == "aisp" {
		isAivis = true
	}

	if adjustmentNanoSec == 0.0 {
		if isAivis {
			adjustmentNanoSec = adjustmentAivisDefault
		} else {
			adjustmentNanoSec = adjustmentVoicevoxDefault
		}
	}

	sublabs, err := construct.Project2subtitles(projPath, adjustmentNanoSec, isAivis, lab)
	if err != nil {
		log.Fatalln(err)
	}
	for _, sublab := range sublabs {
		filename := fmt.Sprintf("%s_%s.srt", sublab.SpeakerId, outputFilename)
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()

		if err := f.Truncate(0); err != nil {
			log.Fatalln(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			log.Fatalln(err)
		}
		if _, err := f.WriteString(sublab.Subtitles.AsSRT()); err != nil {
			log.Fatalln(err)
		}

		if lab {
			filename := fmt.Sprintf("%s_%s.lab", sublab.SpeakerId, outputFilename)
			f, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalln(err)
			}
			if err := f.Truncate(0); err != nil {
				log.Fatalln(err)
			}
			if _, err := f.Seek(0, 0); err != nil {
				log.Fatalln(err)
			}
			if _, err := f.WriteString(sublab.Lab); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
