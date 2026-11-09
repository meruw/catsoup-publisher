package uploads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"catsoup/publisher/internal/diagnostics"
)

// A manual recovery can retransmit the exact saved bytes to this same container.
// It must never create a replacement or infer that a publish request failed.
func CanResumeInstagramBytes(state string, err error) bool {
	var failure *instagramProcessingError
	return state == "IN_PROGRESS" && errors.As(err, &failure) && failure.code == 1363008 && failure.emptyUpload
}

func (c *Client) ResumeInstagramBytes(ctx context.Context, token, container string, m Media) error {
	state, err := c.InstagramContainerState(ctx, token, container)
	if err == nil && state == "FINISHED" {
		return nil
	}
	if !CanResumeInstagramBytes(state, err) {
		return errors.New("Instagram no longer reports an empty failed transfer; its container was retained")
	}
	if m.File == nil || m.Size <= 0 {
		return errors.New("missing prepared Instagram video")
	}
	var accepted struct{ Success bool }
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_started", diagnostics.Details{})
	_, err = c.request(ctx, "POST", "https://rupload.facebook.com/ig-api-upload/v25.0/"+container, "OAuth "+token,
		map[string]string{"Content-Type": "application/octet-stream", "offset": "0", "file_size": fmt.Sprint(m.Size)}, io.NewSectionReader(m.File, 0, m.Size), m.Size, &accepted)
	if err != nil || !accepted.Success {
		state, checkErr := c.InstagramContainerState(ctx, token, container)
		if checkErr == nil && state == "FINISHED" {
			return nil
		}
		if checkErr != nil {
			return checkErr
		}
		if err != nil {
			return err
		}
		return errors.New("Meta did not confirm the recovered transfer; its container was retained")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_finished", diagnostics.Details{})
	return c.waitInstagramContainer(ctx, "Bearer "+token, container, time.Minute)
}

type instagramProcessingError struct {
	container   string
	code        int
	emptyUpload bool
}

func (e *instagramProcessingError) Error() string {
	return fmt.Sprintf("Instagram confirmed an upload/processing error (code %d): %v", e.code, ErrProcessingFailed)
}
func (e *instagramProcessingError) Unwrap() error { return ErrProcessingFailed }

func pendingInstagramFailure(state string, err error) bool {
	var failure *instagramProcessingError
	return state == "IN_PROGRESS" && errors.As(err, &failure) && failure.code == 1363008
}

func (c *Client) InstagramContainerState(ctx context.Context, token, container string) (string, error) {
	if !safeID.MatchString(container) {
		return "", errors.New("missing Instagram container")
	}
	return c.instagramContainerStatus(ctx, "Bearer "+token, container)
}

func (c *Client) PublishReadyInstagram(ctx context.Context, token, account, container string) (string, error) {
	if !safeID.MatchString(account) {
		return "", errors.New("invalid Instagram account")
	}
	state, err := c.InstagramContainerState(ctx, token, container)
	if err != nil {
		return "", err
	}
	if state != "FINISHED" {
		return "", errors.New("Instagram container is not ready; its receipt has been retained")
	}
	return c.publishInstagramContainer(ctx, "Bearer "+token, account, container)
}

// Retry only the observed empty-upload failure, before media_publish. An HTTP
// error, unknown status or interrupted publish can never enter this path.
func (c *Client) instagramWithRetry(ctx context.Context, token string, a Account, m Media, record func(string) error, delay time.Duration) (string, error) {
	ref, err := c.instagramAttempt(ctx, token, a, m, record)
	var failed *instagramProcessingError
	if !errors.As(err, &failed) || failed.code != 1363008 || !failed.emptyUpload {
		return ref, err
	}
	diagnostics.Record(ctx, slog.LevelWarn, "upload_retry_wait", diagnostics.Details{ProviderCode: failed.code})
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
	}
	// Re-read after the delay: a late success must reuse the original container.
	state, checkErr := c.instagramContainerStatus(ctx, "Bearer "+token, failed.container)
	if pendingInstagramFailure(state, checkErr) {
		checkErr = nil
	}
	var checked *instagramProcessingError
	if state == "ERROR" && errors.As(checkErr, &checked) && checked.code == 1363008 && checked.emptyUpload {
		diagnostics.Record(ctx, slog.LevelWarn, "upload_retry_started", diagnostics.Details{ProviderCode: checked.code})
		// Exactly one replacement per send. Its checkpoint precedes all bytes.
		return c.instagramAttempt(ctx, token, a, m, record)
	}
	if checkErr != nil {
		return "", checkErr
	}
	if state == "IN_PROGRESS" {
		if err = c.waitInstagramContainer(ctx, "Bearer "+token, failed.container, time.Minute); err != nil {
			return "", err
		}
		state = "FINISHED"
	}
	if state == "FINISHED" {
		return c.publishInstagramContainer(ctx, "Bearer "+token, a.ID, failed.container)
	}
	return "", errors.New("Instagram has not confirmed a failed upload; check the saved container before retrying")
}

// ConfirmInstagramFailure is read-only. It does not infer failure from a local
// receipt, status text, network failure or a pending/published container.
func (c *Client) ConfirmInstagramFailure(ctx context.Context, token, container string) error {
	if !safeID.MatchString(container) {
		return errors.New("missing Instagram container")
	}
	state, err := c.instagramContainerStatus(ctx, "Bearer "+token, container)
	if state == "IN_PROGRESS" {
		return errors.New("Instagram still reports processing; keep this container and check again later")
	}
	if (state == "ERROR" || state == "EXPIRED") && errors.Is(err, ErrProcessingFailed) {
		return nil
	}
	if err == nil && (state == "ERROR" || state == "EXPIRED") {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("Meta has not confirmed this container failed; keep its receipt and check the result")
}
