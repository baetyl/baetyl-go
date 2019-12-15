package link

import (
	"github.com/baetyl/baetyl-go/log"
	"google.golang.org/grpc/grpclog"
)

type grpcLogger struct {
	*log.Logger
}

func init() {
	grpclog.SetLoggerV2(&grpcLogger{log.With(log.Any("grpc", "log"))})
}

func (l *grpcLogger) Info(args ...interface{}) {
	l.Sugar().Info(args)
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	l.Sugar().Info(args...)
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
	l.Sugar().Infof(format, args...)
}

func (l *grpcLogger) Warning(args ...interface{}) {
	l.Sugar().Warn(args...)
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	l.Sugar().Warn(args...)
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	l.Sugar().Warnf(format, args...)
}

// Error returns
func (l *grpcLogger) Error(args ...interface{}) {
	l.Sugar().Error(args...)
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	l.Sugar().Error(args...)
}

func (l *grpcLogger) Errorf(format string, args ...interface{}) {
	l.Sugar().Errorf(format, args...)
}

func (l *grpcLogger) Fatal(args ...interface{}) {
	l.Sugar().Fatal(args...)
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	l.Sugar().Fatal(args...)
}

func (l *grpcLogger) Fatalf(format string, args ...interface{}) {
	l.Sugar().Fatalf(format, args...)
}

func (l *grpcLogger) V(v int) bool {
	return false
}
