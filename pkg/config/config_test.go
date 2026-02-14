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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/livekit/egress/pkg/types"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/rpc"
)

func TestSegmentNaming(t *testing.T) {
	t.Cleanup(func() {
		_ = os.RemoveAll("conf_test/")
	})

	for _, test := range []struct {
		filenamePrefix               string
		playlistName                 string
		livePlaylistName             string
		expectedStorageDir           string
		expectedPlaylistFilename     string
		expectedLivePlaylistFilename string
		expectedSegmentPrefix        string
	}{
		{
			filenamePrefix: "", playlistName: "playlist", livePlaylistName: "",
			expectedStorageDir: "", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "playlist",
		},
		{
			filenamePrefix: "", playlistName: "conf_test/playlist", livePlaylistName: "conf_test/live_playlist",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "live_playlist.m3u8", expectedSegmentPrefix: "playlist",
		},
		{
			filenamePrefix: "filename", playlistName: "", livePlaylistName: "live_playlist2.m3u8",
			expectedStorageDir: "", expectedPlaylistFilename: "filename.m3u8", expectedLivePlaylistFilename: "live_playlist2.m3u8", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "filename", playlistName: "playlist", livePlaylistName: "",
			expectedStorageDir: "", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "filename", playlistName: "conf_test/", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "filename.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "filename", playlistName: "conf_test/playlist", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "conf_test/", playlistName: "playlist", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "playlist",
		},
		{
			filenamePrefix: "conf_test/filename", playlistName: "playlist", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "conf_test/filename", playlistName: "conf_test/playlist", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "filename",
		},
		{
			filenamePrefix: "conf_test_2/filename", playlistName: "conf_test/playlist", livePlaylistName: "",
			expectedStorageDir: "conf_test/", expectedPlaylistFilename: "playlist.m3u8", expectedLivePlaylistFilename: "", expectedSegmentPrefix: "conf_test_2/filename",
		},
	} {
		p := &PipelineConfig{Info: &livekit.EgressInfo{EgressId: "egress_ID"}}
		o, err := p.getSegmentConfig(&livekit.SegmentedFileOutput{
			FilenamePrefix:   test.filenamePrefix,
			PlaylistName:     test.playlistName,
			LivePlaylistName: test.livePlaylistName,
		})
		require.NoError(t, err)

		require.Equal(t, test.expectedStorageDir, o.StorageDir)
		require.Equal(t, test.expectedPlaylistFilename, o.PlaylistFilename)
		require.Equal(t, test.expectedLivePlaylistFilename, o.LivePlaylistFilename)
		require.Equal(t, test.expectedSegmentPrefix, o.SegmentPrefix)
	}
}

func TestValidateAndUpdateOutputParamsRejectsHLSMP3(t *testing.T) {
	p := &PipelineConfig{
		Outputs: map[types.EgressType][]OutputConfig{
			types.EgressTypeSegments: {
				&SegmentConfig{outputConfig: outputConfig{OutputType: types.OutputTypeHLS}},
			},
		},
	}

	p.AudioEnabled = true
	p.VideoEnabled = false
	p.AudioOutCodec = types.MimeTypeMP3
	p.Info = &livekit.EgressInfo{}

	err := p.validateAndUpdateOutputParams()
	require.Error(t, err)
	require.ErrorContains(t, err, "format application/x-mpegurl incompatible with codec audio/mpeg")
}

func TestValidateAndUpdateOutputParamsRejectsVideoFileMP3(t *testing.T) {
	p := &PipelineConfig{
		Outputs: map[types.EgressType][]OutputConfig{
			types.EgressTypeFile: {
				&FileConfig{outputConfig: outputConfig{OutputType: types.OutputTypeMP3}},
			},
		},
	}

	p.AudioEnabled = true
	p.VideoEnabled = true
	p.AudioOutCodec = types.MimeTypeMP3
	p.VideoOutCodec = types.MimeTypeH264
	p.Info = &livekit.EgressInfo{}

	err := p.validateAndUpdateOutputParams()
	require.Error(t, err)
	require.ErrorContains(t, err, "format audio/mpeg incompatible with codec video/h264")
}

func TestPipelineConfigUpdateExternalConferenceRequest(t *testing.T) {
	p := &PipelineConfig{BaseConfig: BaseConfig{WsUrl: "ws://localhost"}, Outputs: make(map[types.EgressType][]OutputConfig)}
	req := &rpc.StartEgressRequest{
		EgressId: "eg_external",
		Token:    "token",
		Request: &rpc.StartEgressRequest_Web{Web: &livekit.WebEgressRequest{
			Url: "https://example.com/ingest?requestType=external_conference&sessionId=s1&conferenceId=c1&participantId=p1&trackId=t1&role=audio",
			FileOutputs: []*livekit.EncodedFileOutput{{
				Filepath: "external.mp4",
			}},
		}},
	}

	err := p.Update(req)
	require.NoError(t, err)
	require.Equal(t, types.RequestTypeExternalConference, p.RequestType)
	require.Equal(t, types.SourceTypeExternalIngest, p.SourceType)
	require.Equal(t, "s1", p.SessionID)
	require.Equal(t, "c1", p.ConferenceID)
	require.Equal(t, "p1", p.ParticipantID)
	require.Equal(t, "t1", p.IngestTrackID)
	require.True(t, p.AudioEnabled)
	require.False(t, p.VideoEnabled)
}

func TestPipelineConfigUpdateExternalConferenceRequestRejectsInvalidRole(t *testing.T) {
	p := &PipelineConfig{BaseConfig: BaseConfig{WsUrl: "ws://localhost"}, Outputs: make(map[types.EgressType][]OutputConfig)}
	req := &rpc.StartEgressRequest{
		EgressId: "eg_external",
		Token:    "token",
		Request: &rpc.StartEgressRequest_Web{Web: &livekit.WebEgressRequest{
			Url: "https://example.com/ingest?requestType=external_conference&sessionId=s1&conferenceId=c1&participantId=p1&trackId=t1&role=admin",
			FileOutputs: []*livekit.EncodedFileOutput{{
				Filepath: "external.mp4",
			}},
		}},
	}

	err := p.Update(req)
	require.Error(t, err)
	require.ErrorContains(t, err, "role")
}

func TestPipelineConfigUpdateExternalConferenceRequestRejectsInvalidIdentifiers(t *testing.T) {
	p := &PipelineConfig{BaseConfig: BaseConfig{WsUrl: "ws://localhost"}, Outputs: make(map[types.EgressType][]OutputConfig)}
	req := &rpc.StartEgressRequest{
		EgressId: "eg_external",
		Token:    "token",
		Request: &rpc.StartEgressRequest_Web{Web: &livekit.WebEgressRequest{
			Url: "https://example.com/ingest?requestType=external_conference&sessionId=s/1&conferenceId=c1&participantId=p1&trackId=t1&role=video",
			FileOutputs: []*livekit.EncodedFileOutput{{
				Filepath: "external.mp4",
			}},
		}},
	}

	err := p.Update(req)
	require.Error(t, err)
	require.ErrorContains(t, err, "sessionId")
}
