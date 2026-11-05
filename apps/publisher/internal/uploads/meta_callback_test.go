package uploads

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMetaCallbackSetting(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("AppData", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
	check := func(want string) {
		t.Helper()
		got, err := MetaCallbackURL()
		if err != nil || got != want {
			t.Fatalf("callback %q, %v", got, err)
		}
	}
	check("")
	const custom = "https://media.example.org/app/meta/"
	got, err := SetMetaCallbackURL("https://MEDIA.example.org:443/app/meta/")
	if err != nil || got != custom {
		t.Fatalf("normalized URL: %q, %v", got, err)
	}
	for _, bad := range []string{hostedMetaCallback, "https://PUBLISHER.meruw.dev:443/other/", "http://media.example.org/auth/", "https://user@media.example.org/auth/", "https://media.example.org/../auth/", "https://media.example.org/a%2Fb/", "https://media.example.org/auth/?a=b", "https://media.example.org/auth/#x", "https://media.example.org:65536/auth/", "https://media.example.org/auth", "https://127.0.0.1:4000/", "https://local.localhost/"} {
		if _, err = SetMetaCallbackURL(bad); err == nil {
			t.Errorf("accepted invalid URI %q", bad)
		}
		check(custom)
	}
	if err = os.WriteFile(filepath.Join(dir, "CATsoup Publisher", callbackFile), []byte("http://bad.test/"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = MetaCallbackURL(); err == nil {
		t.Fatal("corrupt preference must not silently fall back")
	}
	// The old format could contain the hosted default without acknowledging the notice.
	if err = os.WriteFile(filepath.Join(dir, "CATsoup Publisher", callbackFile), []byte(hostedMetaCallback), 0600); err != nil {
		t.Fatal(err)
	}
	check("")
	if _, err = EnableHostedMetaCallback(); err != nil {
		t.Fatal(err)
	}
	check(hostedMetaCallback)
	if _, err = SetMetaCallbackURL(""); err != nil {
		t.Fatal(err)
	}
	check("")
	if _, err = SetMetaCallbackURL(custom); err != nil {
		t.Fatal(err)
	}
	check(custom)
}
