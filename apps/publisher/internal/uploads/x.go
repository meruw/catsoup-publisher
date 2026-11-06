package uploads

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const xAPI = "https://api.x.com/2/"

var xID = regexp.MustCompile(`^[0-9]{1,25}$`)

// X requires an exact registered callback, including its port. Only a local
// listener is supported; developer settings cannot turn this into a relay.
func validXCallback(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.User != nil || u.Path != "/callback/" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return false
	}
	port, err := strconv.Atoi(u.Port())
	return err == nil && port >= 1024 && port <= 65535 && u.Host == "127.0.0.1:"+strconv.Itoa(port)
}

func CheckXClient(t Token) error {
	cfg, err := desktopConfig("x")
	if err != nil {
		return err
	}
	if t.Client != cfg.fingerprint() {
		return errors.New("X developer client changed; reconnect from Accounts")
	}
	return nil
}

// Conservative local bound: avoids undercounting wide characters and emoji.
// X remains authoritative for URL shortening and combined emoji sequences.
func xTextWeight(text string) int {
	// Even a short URL receives X's t.co length; longer URLs deliberately
	// keep their original length instead of requiring a URL parser dependency.
	for _, match := range regexp.MustCompile(`https?://[^\s]+`).FindAllString(text, -1) {
		if len(match) < 23 {
			text += string(bytes.Repeat([]byte(" "), 23-len(match)))
		}
	}
	n := 0
	for _, r := range text {
		if r <= 0x10ff || (r >= 0x2000 && r <= 0x200d) || (r >= 0x2010 && r <= 0x201f) || (r >= 0x2032 && r <= 0x2037) {
			n++
		} else {
			n += 2
		}
	}
	return n
}

type xMediaResponse struct {
	Data struct {
		ID         string `json:"id"`
		Processing *struct {
			State      string `json:"state"`
			CheckAfter int    `json:"check_after_secs"`
		} `json:"processing_info"`
	} `json:"data"`
}

func (c *Client) x(ctx context.Context, token string, m Media, record func(string) error) (string, error) {
	auth := "Bearer " + token
	var out xMediaResponse
	_, err := c.request(ctx, "POST", xAPI+"media/upload/initialize", auth, jsonHeaders, jsonBody(map[string]any{"media_type": "video/mp4", "total_bytes": m.Size, "media_category": "tweet_video"}), -1, &out)
	if err != nil {
		return "", err
	}
	id := out.Data.ID
	if !xID.MatchString(id) {
		return "", errors.New("X did not return a media ID")
	}
	reference := "media:" + id
	if err = record(reference); err != nil {
		return reference, err
	}
	// Each buffer is bounded to one 4 MiB chunk, independent of video size.
	const chunkSize int64 = 4 << 20
	for offset, segment := int64(0), 0; offset < m.Size; offset, segment = offset+chunkSize, segment+1 {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, e := writer.CreateFormFile("media", "video.mp4")
		if e != nil {
			return reference, e
		}
		if _, e = io.CopyN(part, io.NewSectionReader(m.File, offset, min(chunkSize, m.Size-offset)), min(chunkSize, m.Size-offset)); e != nil {
			return reference, errors.New("could not read staged X video")
		}
		if e = writer.WriteField("segment_index", strconv.Itoa(segment)); e != nil {
			return reference, e
		}
		if e = writer.Close(); e != nil {
			return reference, e
		}
		if _, e = c.request(ctx, "POST", xAPI+"media/upload/"+id+"/append", auth, map[string]string{"Content-Type": writer.FormDataContentType()}, &body, int64(body.Len()), nil); e != nil {
			return reference, e
		}
	}
	out = xMediaResponse{}
	if _, err = c.request(ctx, "POST", xAPI+"media/upload/"+id+"/finalize", auth, nil, nil, 0, &out); err != nil {
		return reference, err
	}
	for attempts := 0; out.Data.Processing != nil; attempts++ {
		p := out.Data.Processing
		if p.State == "succeeded" {
			break
		}
		if p.State != "pending" && p.State != "in_progress" {
			return reference, errors.New("X could not process the video; check its format and account limits")
		}
		if attempts >= 60 {
			return reference, errors.New("X video processing did not finish; check the destination before retrying")
		}
		delay := max(p.CheckAfter, 1)
		if delay > 300 {
			return reference, errors.New("X requested a longer processing wait; check the destination before retrying")
		}
		timer := time.NewTimer(time.Duration(delay) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return reference, ctx.Err()
		case <-timer.C:
		}
		out = xMediaResponse{}
		if _, err = c.request(ctx, "GET", xAPI+"media/upload?command=STATUS&media_id="+id, auth, nil, nil, 0, &out); err != nil {
			return reference, err
		}
		if out.Data.Processing == nil {
			return reference, errors.New("X returned no video processing status")
		}
	}
	var post struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_, err = c.request(ctx, "POST", xAPI+"tweets", auth, jsonHeaders, jsonBody(map[string]any{"text": m.Caption, "media": map[string]any{"media_ids": []string{id}}}), -1, &post)
	if err != nil {
		return reference, err
	}
	if !xID.MatchString(post.Data.ID) {
		return reference, errors.New("X did not return a post ID; check the account before retrying")
	}
	return fmt.Sprintf("https://x.com/i/status/%s", post.Data.ID), nil
}
