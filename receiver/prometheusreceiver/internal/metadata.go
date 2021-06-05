// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"errors"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
)

// MetadataCache is an adapter to prometheus' scrape.Target  and provide only the functionality which is needed
type MetadataCache interface {
	Metadata(metricName string) (scrape.MetricMetadata, bool)
	SharedLabels() labels.Labels
}

type ScrapeManager interface {
	TargetsAll() map[string][]*scrape.Target
}

type metadataService struct {
	sm ScrapeManager
}

func (s *metadataService) Get(job, instance string) (MetadataCache, error) {
	// Lookup against unaltered job label
	if targetGroup, ok := s.sm.TargetsAll()[job]; ok {
		for _, target := range targetGroup {
			if target.Labels().Get(model.InstanceLabel) == instance {
				return &mCache{target}, nil
			}
		}
	}
	// Fallback to lookup through all target groups for cases where job was relabeled
	for _, targetGroup := range s.sm.TargetsAll() {
		for _, target := range targetGroup {
			if target.Labels().Get(model.InstanceLabel) == instance && target.Labels().Get(model.JobLabel) == job {
				return &mCache{target}, nil
			}
		}
	}
	return nil, errors.New("unable to find a target with job=" + job + ", and instance=" + instance)
}

// adapter to get metadata from scrape.Target
type mCache struct {
	t *scrape.Target
}

func (m *mCache) Metadata(metricName string) (scrape.MetricMetadata, bool) {
	return m.t.Metadata(metricName)
}

func (m *mCache) SharedLabels() labels.Labels {
	return m.t.DiscoveredLabels()
}
