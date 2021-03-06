// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proto

import (
	"time"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

type ServerConfig struct {
	MaxHeaderBytes int64 `json:"maxHeaderBytes"`
	MaxBodyBytes   int64 `json:"maxBodyBytes"`

	ReadHeaderTimeout string `json:"readHeaderTimeout"`
	ReadTimeout       string `json:"readTimeout"`
	IdleTimeout       string `json:"idleTimeout"`
	WriteTimeout      string `json:"writeTimeout"`

	LimitTTLUnit     string `json:"limitTTLUnit"`
	LimitConnections int64  `json:"limitConnections"`
	LimitIPLookup    string `json:"limitIPLookup"`

	SslEnabled    bool   `json:"sslEnabled,string"`
	SslMinVersion string `json:"sslMinVersion"`
	SslVerifyPeer bool   `json:"sslVerifyPeer,string"`
	SslCiphers    string `json:"sslCiphers"`

	AutoSyncInterval  string `json:"-"`
	CompactIndexDelta int64  `json:"-"`
	CompactInterval   string `json:"-"`

	EnablePProf bool `json:"enablePProf"`
	EnableCache bool `json:"enableCache"`

	LogRotateSize   int64  `json:"-"`
	LogBackupCount  int64  `json:"-"`
	LogFilePath     string `json:"-"`
	LogLevel        string `json:"-"`
	LogFormat       string `json:"-"`
	LogSys          bool   `json:"-"`
	EnableAccessLog bool   `json:"-"`
	AccessLogFile   string `json:"-"`

	PluginsDir string          `json:"-"`
	Plugins    util.JSONObject `json:"plugins"`

	SelfRegister bool `json:"selfRegister"`

	//clear no-instance services
	ServiceClearEnabled  bool          `json:"serviceClearEnabled"`
	ServiceClearInterval time.Duration `json:"serviceClearInterval"`
	//if a service's existence time reaches this value, it can be cleared
	ServiceTTL time.Duration `json:"serviceTTL"`
	//CacheTTL is the ttl of cache
	CacheTTL time.Duration `json:"cacheTTL"`
}

type ServerInformation struct {
	Version string       `json:"version"`
	Config  ServerConfig `json:"-"`
}

func NewServerInformation() *ServerInformation {
	return &ServerInformation{Config: ServerConfig{Plugins: make(util.JSONObject)}}
}
