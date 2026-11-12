package drafts

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCoversPersistSeparatelyAndRevokeApproval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 16, 16)))
	id, err := s.ImportCover(ctx, b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	d := sample()
	d.Cover = &Cover{Kind: "image", Hash: id}
	d, err = s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	s.db.Exec("UPDATE drafts SET approved_revision=revision")
	d.Cover = &Cover{Kind: "frame", Hash: id, TimeMS: 1200}
	d, err = s.Save(ctx, d)
	if err != nil || d.ApprovedRevision != 0 {
		t.Fatal(d, err)
	}
	s.Close()
	s = openTest(t, path)
	got, err := s.Get(ctx, d.Key)
	if err != nil || !reflect.DeepEqual(d, got) {
		t.Fatal(got, err)
	}
	raw, mime, err := s.CoverImage(ctx, id)
	if err != nil || mime != "image/png" || !bytes.Equal(raw, b.Bytes()) {
		t.Fatal(mime, err)
	}
	if _, err = s.ImportCover(ctx, []byte("not a picture")); err == nil {
		t.Fatal("invalid image accepted")
	}
	if _, err = s.ImportCover(ctx, make([]byte, (2<<20)+1)); err == nil {
		t.Fatal("oversized image accepted")
	}
	if _, _, err = s.CoverImage(ctx, "../../private"); err == nil {
		t.Fatal("invalid address accepted")
	}
	for _, platform := range []string{"instagram", "tiktok", "facebook", "x"} {
		d.Key.Platform = platform
		d.Cover = &Cover{Kind: "image", Hash: id}
		if _, err = s.Save(ctx, d); err == nil {
			t.Fatal("unsupported image", platform)
		}
	}
	s.db.Exec("UPDATE cover_assets SET image=?", []byte("corrupted"))
	if _, _, err = s.CoverImage(ctx, id); err == nil {
		t.Fatal("corrupt asset accepted")
	}
}

func TestStaleInstagramRecoveryRetainsReceipt(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "db"))
	for _, state := range []string{"sending", "publishing", "published", "check_publish"} {
		for _, old := range []bool{false, true} {
			s.db.Exec("DELETE FROM upload_attempts")
			started := time.Now().Unix()
			if old {
				started -= 3600
			}
			s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('instagram','account','hash',?,'original',?)", state, started)
			if err := s.RecoverStaleInstagram("account", "hash"); err != nil {
				t.Fatal(err)
			}
			got, err := s.Receipt("instagram", "account", "hash")
			want := state
			if old && state == "sending" {
				want = "check"
			}
			if old && state == "publishing" {
				want = "check_publish"
			}
			if err != nil || got.State != want || got.Reference != "original" {
				t.Fatal(state, old, got, err)
			}
		}
	}
}
