package gohipernetFake

import (
	"go.uber.org/zap/zapcore"
	"time"
)

var NTELIB_LOG_DEBUG func(msg string, fields ...zapcore.Field)
var NTELIB_LOG_INFO func(msg string, fields ...zapcore.Field)
var NTELIB_LOG_ERROR func(msg string, fields ...zapcore.Field)

// 서버 실행 여부
var _server_state_running bool = true

func wrapLoggerFunc() {
	NTELIB_LOG_DEBUG = Logger.Debug
	NTELIB_LOG_INFO = Logger.Info
	NTELIB_LOG_ERROR = Logger.Error
}

func NetLib_StopServer() {
	_server_state_running = false
}

func NetLib_IsRunningServer() bool {
	return _server_state_running
}

// 유닉스 타임 스탬프 시간
func NetLib_GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}