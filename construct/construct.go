package construct

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/1l0/vv2srt/model"
	"github.com/martinlindhe/subtitles"
)

const (
	sec2nanoSec = 1000000000.0
)

func offset2lab(offset float64) int {
	return int(offset * 0.01)
}

func Project2subtitles(projectFilePath string, adjustmentNanoSec float64, isAivis, lab bool) ([]model.SubLab, error) {
	proj, err := loadProject(projectFilePath)
	if err != nil {
		return nil, err
	}
	return project2subtitles(proj, adjustmentNanoSec, isAivis, lab)
}

func loadProject(projectFilePath string) (*model.Project, error) {
	readfile, err := os.Open(projectFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer readfile.Close()
	dec := json.NewDecoder(readfile)
	proj := &model.Project{}
	if err := dec.Decode(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func project2subtitles(proj *model.Project, adjustmentNanoSec float64, isAivis, lab bool) ([]model.SubLab, error) {
	sublabMap := make(map[string]*model.SubLab)
	zero := makeTime(0, 0, 0, 0)
	epoch := zero
	var offset float64 = 0.0

	for i, key := range proj.Talk.AudioKeys {
		it, exist := proj.Talk.AudioItems[key]
		if !exist {
			return nil, fmt.Errorf("audio item not found for the key: %s", key)
		}
		b, err := json.Marshal(it)
		if err != nil {
			return nil, err
		}
		item := &model.AudioItem{}
		if err := json.Unmarshal(b, item); err != nil {
			return nil, err
		}

		sid := item.Voice.SpeakerId
		sublab, exist := sublabMap[sid]
		if !exist {
			sublab = &model.SubLab{
				SpeakerId: sid,
				Subtitles: &subtitles.Subtitle{
					Captions: []subtitles.Caption{},
				},
				Lab: "",
			}
			sublabMap[sid] = sublab
		}

		speedScale := item.Query.SpeedScale

		prePhenome := item.Query.PrePhonemeLength * sec2nanoSec / speedScale
		if lab {
			sublab.Lab += fmt.Sprintf("%d %d %s\n", offset2lab(offset), offset2lab(offset+prePhenome), "pau")
		}
		offset += prePhenome

		for _, acc := range item.Query.AccentPhrases {
			for _, mo := range acc.Moras {
				if mo.Consonant != "" {
					consonant := (mo.ConsonantLength * sec2nanoSec) / speedScale
					if lab {
						sublab.Lab += fmt.Sprintf("%d %d %s\n", offset2lab(offset), offset2lab(offset+consonant), mo.Consonant)
					}
					offset += consonant
				}
				if mo.Vowel != "" {
					vowel := (mo.VowelLength * sec2nanoSec) / speedScale
					if lab {
						sublab.Lab += fmt.Sprintf("%d %d %s\n", offset2lab(offset), offset2lab(offset+vowel), mo.Vowel)
					}
					offset += vowel
				}
			}
			if acc.PauseMora != nil && acc.PauseMora.Vowel != "" {
				pause := (acc.PauseMora.VowelLength * item.Query.PauseLengthScale * sec2nanoSec) / speedScale
				if lab {
					sublab.Lab += fmt.Sprintf("%d %d %s\n", offset2lab(offset), offset2lab(offset+pause), acc.PauseMora.Vowel)
				}
				offset += pause
			}
		}
		postPhenome := item.Query.PostPhonemeLength * sec2nanoSec / speedScale
		if lab {
			sublab.Lab += fmt.Sprintf("%d %d %s\n", offset2lab(offset), offset2lab(offset+postPhenome+adjustmentNanoSec), "pau")
		}
		offset += postPhenome + adjustmentNanoSec

		nextEpoch := zero.Add(time.Duration(offset))
		sublab.Subtitles.Captions = append(sublab.Subtitles.Captions, subtitles.Caption{
			Seq:   i + 1,
			Start: epoch,
			End:   nextEpoch,
			Text:  []string{item.Text},
		})
		epoch = nextEpoch
	}
	platform := func() string {
		if isAivis {
			return "AivisSearch"
		}
		return "VOICEVOX"
	}()

	result := make([]model.SubLab, 0, len(sublabMap))
	for _, sublab := range sublabMap {
		result = append(result, *sublab)
	}
	fmt.Printf(
		"Platform: %s\nDuration: %02d:%02d:%02d.%d\nSpeakers: %d\n",
		platform,
		epoch.Hour(),
		epoch.Minute(),
		epoch.Second(),
		epoch.Nanosecond(),
		len(result),
	)
	return result, nil
}

func makeTime(h int, m int, s int, ms int) time.Time {
	return time.Date(0, 1, 1, h, m, s, ms*1000*1000, time.UTC)
}
