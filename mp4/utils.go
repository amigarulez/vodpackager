package mp4

import (
	"fmt"
	"github.com/Eyevinn/mp4ff/mp4"
)

func getTrackType(hdlr *mp4.HdlrBox) (string, error) {
	switch hdlrType := hdlr.HandlerType; hdlrType {
	case "vide":
		return "video", nil
	case "soun":
		return "audio", nil
	default:
		return "", fmt.Errorf("hdlr type %q not supported", hdlrType)
	}
}
