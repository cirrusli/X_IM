package middleware

import (
	x "X_IM"
	"X_IM/logger"
	"X_IM/wire/pkt"
	"fmt"
	"runtime"
	"strings"
)

// Recover 捕获panic并且返回错误信息
func Recover() x.HandlerFunc {
	return func(ctx x.Context) {
		//如果Next()中出现panic，会被这里defer捕获
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈信息
				var callers []string
				for i := 1; ; i++ {
					_, file, line, got := runtime.Caller(i)
					if !got {
						break
					}
					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}
				logger.WithFields(logger.Fields{
					"ChannelID": ctx.Header().ChannelID,
					"Command":   ctx.Header().Command,
					"Seq":       ctx.Header().Sequence,
				}).Error(err, strings.Join(callers, "\n"))

				_ = ctx.Resp(pkt.Status_SystemException, &pkt.ErrorResp{Message: "Internal server error"})
			}
		}()
		ctx.Next()
	}
}
