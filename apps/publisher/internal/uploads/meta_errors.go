package uploads

import (
	"encoding/json"
	"fmt"
)

// Parse identifiers only. In particular, rupload can wrap a Graph response as
// JSON inside debug_info.message; neither message is safe to store verbatim.
func metaErrorIdentifiers(raw []byte) (code, subcode int, kind string) {
	var envelope struct {
		Error struct {
			Code    int
			Subcode int `json:"error_subcode"`
		}
		Debug struct {
			Type    string
			Message string
		} `json:"debug_info"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return
	}
	code, subcode = envelope.Error.Code, envelope.Error.Subcode
	if code == 0 && subcode == 0 {
		var nested struct {
			Error struct {
				Code    int
				Subcode int `json:"error_subcode"`
			}
		}
		if json.Unmarshal([]byte(envelope.Debug.Message), &nested) == nil {
			code, subcode = nested.Error.Code, nested.Error.Subcode
		}
	}
	switch envelope.Debug.Type {
	case "ProcessingFailedError":
		kind = "meta_processing_failed"
	case "NotAuthorizedError":
		kind = "meta_not_authorized"
	case "InvalidParameterError":
		kind = "meta_invalid_parameter"
	case "PartialRequestError":
		kind = "meta_partial_request"
	}
	return
}

// Keep only provider identifiers; response text can contain secrets and URLs.
type metaResponseError struct {
	status, code, subcode int
	kind                  string
}

func (e *metaResponseError) Error() string {
	return fmt.Sprintf("Meta returned HTTP %d (code %d, subcode %d, type %s); check the saved upload result", e.status, e.code, e.subcode, e.kind)
}
