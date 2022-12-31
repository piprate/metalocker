// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// GracefulWarden listens for system signals and invokes Close() methods
// to all 'subscribed' io.Closers (use CloseOnShutdown method to subscribe a closer)
type GracefulWarden struct {
	closers []io.Closer
}

func (gw *GracefulWarden) closeAll() (err error) {
	for _, closer := range gw.closers {
		if closerErr := closer.Close(); err == nil && closerErr != nil {
			err = closerErr
		}
	}
	return
}

// Given Closer will have Close() method invoked when the server process
// receives some system signals (see implementation below for full list)
func (gw *GracefulWarden) CloseOnShutdown(closer io.Closer) {
	gw.closers = append(gw.closers, closer)
}

func (gw *GracefulWarden) handleSystemSignals(patienceTimeout int64) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		sig := <-signalCh
		systemSignal, ok := sig.(syscall.Signal)
		if !ok {
			log.Error().Msg("Not a unix signal")
		}
		switch systemSignal {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			log.Info().Int("signal", int(systemSignal)).Msg("Got system signal: shutting down")
			allClosedCh := make(chan bool)
			go func() {
				if err := gw.closeAll(); err != nil {
					log.Err(err).Msg("Error shutting down")
					os.Exit(1)
				}
				allClosedCh <- true
			}()
			select {
			case <-allClosedCh:
				log.Info().Msg("Shut down.")
				os.Exit(0)
			case <-time.After(time.Second * time.Duration(patienceTimeout)):
				log.Error().Msg("Timeout shutting down. Exiting uncleanly.")
				os.Exit(1)
			}
		default:
			log.Error().Msg("Received another signal, should not happen.")
		}
	}
}

var wardenLock sync.Mutex
var warden *GracefulWarden

// Create a warden. Only one copy of GracefulWarden can be present in the system.
func NewGracefulWarden(patienceTimeout int64) *GracefulWarden {
	if warden != nil {
		log.Warn().Msg("Attempt to create a new GracefulWarden instance. Returning the existing warden")
		return warden
	} else {
		wardenLock.Lock()

		warden = &GracefulWarden{}

		go warden.handleSystemSignals(patienceTimeout)

		wardenLock.Unlock()

		return warden
	}
}
