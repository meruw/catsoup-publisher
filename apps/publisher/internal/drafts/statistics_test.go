package drafts

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"catsoup/publisher/internal/statistics"
)

func TestStatisticsPersistenceBucketsAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	d, err := s.Save(ctx, sample())
	if err != nil {
		t.Fatal(err)
	}
	base := time.Now().UTC().Truncate(5 * time.Minute).Add(-time.Hour)
	point := func(at time.Time, count string) statistics.Sample {
		return statistics.Sample{UpdatedAt: at.Format(time.RFC3339), Items: []statistics.Metric{{Label: "Views", Value: count}}}
	}
	write := func(at time.Time, count string) []statistics.Sample {
		t.Helper()
		got, err := s.RecordStatistics(ctx, "youtube", "channel-es", "video", point(at, count))
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	write(base, "0")
	got := write(base.Add(time.Minute), "5")
	if len(got) != 1 || got[0].Items[0].Value != "5" {
		t.Fatal(got)
	}
	got = write(base, "1") // A late response cannot replace a newer observation.
	if got[0].Items[0].Value != "5" {
		t.Fatal(got)
	}
	got = write(base.Add(5*time.Minute), "3") // Provider corrections may lower totals.
	if len(got) != 2 || got[1].Items[0].Value != "3" {
		t.Fatal(got)
	}
	s.Close()
	s = openTest(t, path)
	got = write(base.Add(10*time.Minute), "18446744073709551615")
	if len(got) != 3 || got[2].Items[0].Value != "18446744073709551615" {
		t.Fatal(got)
	}
	for _, dest := range [][3]string{{"youtube", "channel-en", "video"}, {"youtube", "channel-es", "repost"}, {"facebook", "channel-es", "video"}} {
		separate, err := s.RecordStatistics(ctx, dest[0], dest[1], dest[2], point(base, "7"))
		if err != nil || len(separate) != 1 {
			t.Fatal(separate, err)
		}
	}
	after, err := s.Get(ctx, d.Key)
	if err != nil || after != d {
		t.Fatal("history changed draft", err)
	}
	got = write(time.Now().AddDate(-2, 0, 0), "1")
	if len(got) != 3 {
		t.Fatal("expired sample was retained", got)
	}
	bad := point(base, "-1")
	if _, err = s.RecordStatistics(ctx, "youtube", "channel-es", "video", bad); err == nil {
		t.Fatal("negative counter accepted")
	}
	bad = point(base, "1")
	bad.Items = append(bad.Items, bad.Items[0])
	if _, err = s.RecordStatistics(ctx, "youtube", "channel-es", "video", bad); err == nil {
		t.Fatal("duplicate metric accepted")
	}
}

func TestStatisticsMigrationFromSeven(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d, err := s.Save(context.Background(), sample())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("DROP TABLE cover_assets; DROP TABLE cover_results; ALTER TABLE drafts DROP COLUMN cover; DROP TABLE statistics_samples; ALTER TABLE upload_attempts DROP COLUMN upload_hash; ALTER TABLE upload_attempts DROP COLUMN preparation;UPDATE schema_version SET version=7"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	after, err := s.Get(context.Background(), d.Key)
	if err != nil || after != d {
		t.Fatal("migration lost draft", err)
	}
	var version int
	if err = s.db.QueryRow("SELECT version FROM schema_version").Scan(&version); err != nil || version != 13 {
		t.Fatal(version, err)
	}
}

func TestStatisticsBoundedHistory(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	base := time.Now().UTC().Truncate(5 * time.Minute)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2005; i++ {
		at := base.Add(-time.Duration(i) * 5 * time.Minute).Unix()
		if _, err = tx.Exec("INSERT INTO statistics_samples VALUES(?,?,?,?,?,?)", "youtube", "account", "video", at/300, at, `[{"label":"Views","value":"1"}]`); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	got, err := s.RecordStatistics(ctx, "youtube", "account", "video", statistics.Sample{UpdatedAt: base.Format(time.RFC3339), Items: []statistics.Metric{{Label: "Likes", Value: "0"}}})
	if err != nil || len(got) != 2000 {
		t.Fatal(len(got), err)
	}
	if got[len(got)-1].Items[0].Label != "Likes" {
		t.Fatal("missing metrics were filled")
	}
}
