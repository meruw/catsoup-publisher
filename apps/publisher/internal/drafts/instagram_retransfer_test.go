package drafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstagramRetransferGuard(t *testing.T) {
	for _, scenario := range []string{"ok", "hash", "reference", "approval", "revision", "account", "published", "check_publish", "sending"} {
		t.Run(scenario, func(t *testing.T) {
			s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
			ctx := context.Background()
			d := sample()
			d.Key.Platform, d.Key.Account, d.DeliveryMode = "instagram", "account:ig1", "publish"
			if e := s.PutProfile(Profile{Platform: "instagram", Language: "es", ID: "ig1", Name: "IG"}, []byte("fixture"), ""); e != nil {
				t.Fatal(e)
			}
			d, e := s.Save(ctx, d)
			if e != nil {
				t.Fatal(e)
			}
			d, e = s.SetPostApproval(ctx, d.Key, d.Revision, true)
			if e != nil {
				t.Fatal(e)
			}
			if e = s.Reserve(d, "ig1"); e != nil {
				t.Fatal(e)
			}
			hash := strings.Repeat("b", 64)
			if e = s.SaveInstagramCheckpoint(d, "ig1", "original", hash); e != nil {
				t.Fatal(e)
			}
			state := "check"
			if scenario == "published" || scenario == "check_publish" || scenario == "sending" {
				state = scenario
			}
			s.SaveReceipt(d, "ig1", state, "")
			ref, account := "original", "ig1"
			switch scenario {
			case "hash":
				hash = strings.Repeat("c", 64)
			case "reference":
				ref = "other"
			case "account":
				account = "ig2"
			case "approval":
				s.SetPostApproval(ctx, d.Key, d.Revision, false)
			case "revision":
				other := d
				other.Title = "edited"
				s.Save(ctx, other)
			}
			e = s.ReserveInstagramRetransfer(d, account, ref, hash)
			if (e == nil) != (scenario == "ok") {
				t.Fatal(e)
			}
			if scenario == "ok" {
				if s.ReserveInstagramRetransfer(d, account, ref, hash) == nil {
					t.Fatal("concurrent recovery")
				}
				r, _ := s.Receipt("instagram", "ig1", d.VideoHash)
				if r.Reference != "original" || r.State != "sending" {
					t.Fatal(r)
				}
			}
		})
	}
}
