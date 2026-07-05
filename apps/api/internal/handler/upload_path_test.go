package handler

import "testing"

func TestParseAttachmentPath(t *testing.T) {
	iid, aid, ok := parseAttachmentPath("attachments/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222")
	if !ok {
		t.Fatalf("expected ok for a well-formed attachment path")
	}
	if iid.String() != "11111111-1111-1111-1111-111111111111" || aid.String() != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("parsed ids wrong: issue=%s asset=%s", iid, aid)
	}
	bad := []string{
		"uploads/2026/07/x.png",
		"attachments/not-a-uuid/22222222-2222-2222-2222-222222222222",
		"attachments/11111111-1111-1111-1111-111111111111",
		"attachments/a/b/c",
		"attachments//",
	}
	for _, p := range bad {
		if _, _, ok := parseAttachmentPath(p); ok {
			t.Errorf("parseAttachmentPath(%q) = ok, want not-ok", p)
		}
	}
}

func TestIsServableObjectPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"uploads/2026/07/abc.png", true},
		{"attachments/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222", true},
		{"", false},
		{"attachments/../uploads/secret", false},
		{"uploads/../../etc/passwd", false},
		{"secrets/key", false},
		{"attachmentsfoo/x", false},
		{"uploadsfoo/x", false},
	}
	for _, c := range cases {
		if got := isServableObjectPath(c.path); got != c.want {
			t.Errorf("isServableObjectPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
