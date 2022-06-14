package internal

import (
	"github.com/rs/zerolog/log"
	"runtime"
	"time"
)

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		log.Error().Msg("can not get context.")
	}
	funcName := runtime.FuncForPC(pc).Name()
	return funcName
}

func TimeTrack(start time.Time) {
	elapsed := time.Since(start)
	log.Info().Stringer("Response Time", elapsed).Msg("")
}
