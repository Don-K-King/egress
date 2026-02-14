// Copyright 2023 LiveKit, Inc.
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

package source

import (
	"context"

	"github.com/livekit/egress/pkg/config"
	"github.com/livekit/egress/pkg/gstreamer"
	"github.com/livekit/egress/pkg/types"
)

// ExternalIngestSource encapsulates ingest-oriented conference capture while
// preserving the controller contract and reusing the SDK transport machinery.
type ExternalIngestSource struct {
	inner *SDKSource
}

func NewExternalIngestSource(ctx context.Context, p *config.PipelineConfig, callbacks *gstreamer.Callbacks) (*ExternalIngestSource, error) {
	adapted := *p
	adapted.RequestType = types.RequestTypeTrackComposite
	adapted.Info.RoomName = p.ConferenceID

	sdkSource, err := NewSDKSource(ctx, &adapted, callbacks)
	if err != nil {
		return nil, err
	}

	return &ExternalIngestSource{inner: sdkSource}, nil
}

func (s *ExternalIngestSource) StartRecording() <-chan struct{} { return s.inner.StartRecording() }
func (s *ExternalIngestSource) EndRecording() <-chan struct{}   { return s.inner.EndRecording() }
func (s *ExternalIngestSource) GetStartedAt() int64             { return s.inner.GetStartedAt() }
func (s *ExternalIngestSource) GetEndedAt() int64               { return s.inner.GetEndedAt() }
func (s *ExternalIngestSource) Close()                          { s.inner.Close() }
func (s *ExternalIngestSource) SetTimeProvider(tp gstreamer.TimeProvider) {
	s.inner.SetTimeProvider(tp)
}
