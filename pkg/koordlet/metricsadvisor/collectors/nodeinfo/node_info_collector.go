/*
Copyright 2022 The Koordinator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodeinfo

import (
	"time"

	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/clay-wangzhi/cfs-quota-burst/pkg/koordlet/metriccache"
	"github.com/clay-wangzhi/cfs-quota-burst/pkg/koordlet/metricsadvisor/framework"
	koordletutil "github.com/clay-wangzhi/cfs-quota-burst/pkg/koordlet/util"
)

const (
	CollectorName = "NodeInfoCollector"
)

// TODO more ut is needed for this plugin
type nodeInfoCollector struct {
	collectInterval time.Duration
	storage         metriccache.KVStorage
	started         *atomic.Bool
}

func New(opt *framework.Options) framework.Collector {
	return &nodeInfoCollector{
		collectInterval: opt.Config.CollectNodeCPUInfoInterval,
		storage:         opt.MetricCache,
		started:         atomic.NewBool(false),
	}
}

func (n *nodeInfoCollector) Enabled() bool {
	return true
}

func (n *nodeInfoCollector) Setup(s *framework.Context) {}

func (n *nodeInfoCollector) Run(stopCh <-chan struct{}) {
	go wait.Until(n.collectNodeInfo, n.collectInterval, stopCh)
}

func (n *nodeInfoCollector) Started() bool {
	return n.started.Load()
}

func (n *nodeInfoCollector) collectNodeInfo() {
	started := time.Now()

	err := n.collectNodeCPUInfo()
	if err != nil {
		klog.Warningf("failed to collect node CPU info, err: %s", err)
		return
	}

	n.started.Store(true)
	klog.V(4).Infof("collect node info finished, elapsed %s", time.Since(started).String())
}

func (n *nodeInfoCollector) collectNodeCPUInfo() error {
	klog.V(6).Info("start collect node cpu info")

	localCPUInfo, err := koordletutil.GetLocalCPUInfo()
	if err != nil {
		return err
	}

	nodeCPUInfo := &metriccache.NodeCPUInfo{
		ProcessorInfos: localCPUInfo.ProcessorInfos,
		TotalInfo:      localCPUInfo.TotalInfo,
	}
	klog.V(6).Infof("collect cpu info finished, info: %+v", nodeCPUInfo)

	n.storage.Set(metriccache.NodeCPUInfoKey, nodeCPUInfo)
	klog.V(4).Infof("collectNodeCPUInfo finished, processors num %v", len(nodeCPUInfo.ProcessorInfos))
	return nil
}
