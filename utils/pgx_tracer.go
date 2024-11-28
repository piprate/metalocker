package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ZerologQueryTracer implements pgx.QueryTracer and logs query details using Zerolog logger.
type ZerologQueryTracer struct {
	LogLevel zerolog.Level
}

func NewZerologQueryTracer(lvl zerolog.Level) *ZerologQueryTracer {
	return &ZerologQueryTracer{LogLevel: lvl}
}

type ctxKey int

const (
	_ ctxKey = iota
	tracelogQueryCtxKey
)

type traceQueryData struct {
	startTime time.Time
	sql       string
	args      []any
}

func (tl *ZerologQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, tracelogQueryCtxKey, &traceQueryData{
		startTime: time.Now(),
		sql:       data.SQL,
		args:      data.Args,
	})
}

func (tl *ZerologQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryData := ctx.Value(tracelogQueryCtxKey).(*traceQueryData)

	endTime := time.Now()
	interval := endTime.Sub(queryData.startTime)

	if data.Err != nil {
		if tl.shouldLog(zerolog.ErrorLevel) {
			log.Err(data.Err).
				Str("sql", queryData.sql).
				Any("args", logQueryArgs(queryData.args)).
				Dur("duration", interval).
				Uint32("pid", getPID(conn)).
				Str("module", "pgx").
				Msg("Query failed")
		}
		return
	}

	if tl.shouldLog(zerolog.DebugLevel) {
		log.Info().
			Str("sql", queryData.sql).
			Any("args", logQueryArgs(queryData.args)).
			Dur("time", interval).
			Uint32("pid", getPID(conn)).
			Str("module", "pgx").
			Int64("rowCount", data.CommandTag.RowsAffected()).
			Msg("Query")
	}
}

func (tl *ZerologQueryTracer) shouldLog(lvl zerolog.Level) bool {
	return tl.LogLevel >= lvl
}

func getPID(conn *pgx.Conn) uint32 {
	pgConn := conn.PgConn()
	if pgConn != nil {
		pid := pgConn.PID()
		if pid != 0 {
			return pid
		}
	}
	return 0
}

func logQueryArgs(args []any) []any {
	if len(args) == 0 {
		return nil
	}

	logArgs := make([]any, 0, len(args)-1)

	for _, a := range args[1:] {
		switch v := a.(type) {
		case []byte:
			if len(v) < 64 {
				a = hex.EncodeToString(v)
			} else {
				a = fmt.Sprintf("%x (truncated %d bytes)", v[:64], len(v)-64)
			}
		case string:
			if len(v) > 64 {
				var l, w int
				for ; l < 64; l += w {
					_, w = utf8.DecodeRuneInString(v[l:])
				}
				if len(v) > l {
					a = fmt.Sprintf("%s (truncated %d bytes)", v[:l], len(v)-l)
				}
			}
		}
		logArgs = append(logArgs, a)
	}

	return logArgs
}
