package player

import (
	"io"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/minimp3"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/pkg/errors"
	minimp3pkg "github.com/tosone/minimp3"
)

func DecodeSong(t SongType, r io.ReadSeekCloser) (streamer beep.StreamSeekCloser, format beep.Format, err error) {
	switch t {
	case Mp3:
		switch configs.ConfigRegistry.Player.BeepMp3Decoder {
		case types.BeepMiniMp3Decoder:
			minimp3pkg.BufferSize = 1024 * 50
			streamer, format, err = minimp3.Decode(r)
		default:
			streamer, format, err = mp3.Decode(r)
		}
	case Wav:
		streamer, format, err = wav.Decode(r)
	case Ogg:
		streamer, format, err = vorbis.Decode(r)
	case Flac:
		streamer, format, err = flac.Decode(r)
	default:
		err = errors.Errorf("Unknown song type(%d)", t)
	}
	return
}
