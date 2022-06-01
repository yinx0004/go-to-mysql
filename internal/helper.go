package internal

import (
	"github.com/rs/zerolog/log"
	"runtime"
)

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		log.Error().Msg("can not get context.")
	}
	funcName := runtime.FuncForPC(pc).Name()
	return funcName
}