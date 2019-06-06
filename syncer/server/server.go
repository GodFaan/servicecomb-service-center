/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package server

import (
	"context"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	"github.com/apache/servicecomb-service-center/syncer/pkg/syssig"
	"github.com/apache/servicecomb-service-center/syncer/pkg/ticker"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/serf"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter"
)

// Server struct for syncer
type Server struct {
	// Syncer configuration
	conf *config.Config

	// Ticker for Syncer
	tick *ticker.TaskTicker

	// Wrap the servicecenter
	servicecenter servicecenter.Servicecenter

	etcd *etcd.Agent

	// Wraps the serf agent
	agent *serf.Agent

	// Wraps the grpc server
	grpc *grpc.Server
}

// NewServer new server with Config
func NewServer(conf *config.Config) *Server {
	return &Server{
		conf: conf,
	}
}

// Run syncer Server
func (s *Server) Run(ctx context.Context) {
	s.initPlugin()

	if err := s.runEtcd(); err != nil {
		return
	}

	if err := s.runSerf(); err != nil {
		return
	}

	if err := s.runGrpc(); err != nil {
		return
	}

	if err := s.createServiceCenter(); err != nil {
		return
	}

	if err := s.runTicker(); err != nil {
		return
	}

	s.waitQuit(ctx)
}

// Stop Syncer Server
func (s *Server) Stop() {
	if s.tick != nil {
		s.tick.Stop()
	}

	if s.agent != nil {
		// removes the serf eventHandler
		s.agent.DeregisterEventHandler(s)
		//Leave from Serf
		s.agent.Leave()
		// closes this serf agent
		s.agent.Shutdown()
	}

	if s.grpc != nil {
		s.grpc.Stop()
	}

	if s.etcd != nil {
		s.etcd.Stop()
	}

	// Closes all goroutines in the pool
	gopool.CloseAndWait()
}

// initPlugin Initialize the plugin and load the external plugin according to the configuration
func (s *Server) initPlugin() {
	plugins.SetPluginConfig(plugins.PluginServicecenter.String(), s.conf.ServicecenterPlugin)
	plugins.LoadPlugins()
}

func (s *Server) runEtcd() (err error) {
	s.etcd = etcd.NewAgent(etcd.DefaultConfig())
	if err = s.etcd.Run(); err != nil {
		log.Errorf(err, "Run etcd failed, %s", err)
	}
	return
}

func (s *Server) runSerf() (err error) {
	s.agent, err = serf.Create(s.conf.Config, createLogFile(s.conf.LogFile))
	if err != nil {
		log.Errorf(err, "Create serf failed, %s", err)
		return
	}

	s.agent.RegisterEventHandler(s)

	if err = s.agent.Start(); err != nil {
		log.Errorf(err, "Start serf failed, %s", err)
		return
	}

	if s.conf.JoinAddr != "" {
		_, err = s.agent.Join([]string{s.conf.JoinAddr}, false)
		if err != nil {
			log.Errorf(err, "Join serf cluster failed, %s", err)
		}
	}
	return
}

func (s *Server) runGrpc() (err error) {
	s.grpc = grpc.NewServer(s.conf.RPCAddr, s)
	if err = s.grpc.Run(); err != nil {
		log.Errorf(err, "Run grpc failed, %s", err)
	}
	return err
}

func (s *Server) createServiceCenter() (err error) {
	s.servicecenter, err = servicecenter.NewServicecenter(strings.Split(s.conf.SCAddr, ","), s.etcd.Storage())
	if err != nil {
		log.Errorf(err, "Create service center failed, %s", err)
	}
	return
}

func (s *Server) runTicker() error {
	s.tick = ticker.NewTaskTicker(s.conf.TickerInterval, s.tickHandler)
	gopool.Go(s.tick.Start)
	return nil
}

// waitQuit Waiting for system quit signal
func (s *Server) waitQuit(ctx context.Context) {
	err := syssig.AddSignalsHandler(func() {
		s.Stop()
	}, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	if err != nil {
		log.Errorf(err, "Syncer add signals handler failed")
		return
	}
	syssig.Run(ctx)
}

// createLogFile create log file
func createLogFile(logFile string) (fw io.Writer) {
	fw = os.Stderr
	if logFile == "" {
		return
	}

	f, err := utils.OpenFile(logFile)
	if err != nil {
		log.Errorf(err, "Syncer open log file %s failed", logFile)
		return
	}
	return f
}
