package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/sabariramc/goserverbase/log"
	glog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type gormLogger struct {
	config   *glog.Config
	l        *log.Logger
	LogLevel glog.LogLevel
}

func NewLogger(log *log.Logger, config *glog.Config) *gormLogger {
	l := &gormLogger{l: log, config: config}
	return l
}

// LogMode log mode
func (l *gormLogger) LogMode(level glog.LogLevel) glog.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glog.Info {
		l.l.Info(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glog.Warn {
		l.l.Warning(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glog.Error {
		l.l.Error(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Trace print sql message
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= glog.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glog.Error && (!errors.Is(err, glog.ErrRecordNotFound) || !l.config.IgnoreRecordNotFoundError):
		l.l.Error(ctx, "SQL ERROR", l.getMessage(begin, elapsed, fc, err))
	case elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 && l.LogLevel >= glog.Warn:
		l.l.Notice(ctx, "SQL SLOW QUERY", l.getMessage(begin, elapsed, fc, err))
	case l.LogLevel == glog.Info:
		l.l.Debug(ctx, "SQL LOG", l.getMessage(begin, elapsed, fc, err))
	}
}

func (l *gormLogger) getMessage(begin time.Time, elapsed time.Duration, fc func() (string, int64), err error) map[string]interface{} {
	sql, rows := fc()
	return map[string]interface{}{
		"error":        err,
		"fileTrace":    utils.FileWithLineNum(),
		"runTime":      float64(elapsed.Nanoseconds()) / 1e6,
		"sql":          sql,
		"rowsReturned": rows,
	}
}
