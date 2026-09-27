package main

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello World", "hello-world"},
		{"  Foo   Bar  ", "foo-bar"},
		{" 9-c.d", "9-c-d"},
		{"纯中文path", "path"},
		{"A++9ASCII2", "a-9ascii2"},
		{"", ""},
		{"   ", ""},
		{"---", ""},
		{"A", "a"},
		{"a+b+c", "a-b-c"},
		{"foo(bar)", "foobar"},
		{"Foo's Bar", "foos-bar"},
		{"a/b", "a-b"},
		{"My.File", "my-file"},
		{"under_score", "under-score"},
		{"x[y]z", "xyz"},
		{"{brackets}", "brackets"},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIDAllocatorNext(t *testing.T) {
	a := &idAllocator{}
	// 首次返回裸 slug（无后缀）
	if got := a.Next([]string{"My App"}); got != "my-app" {
		t.Fatalf("first Next = %q, want %q", got, "my-app")
	}
	// 同名后续从序号 2 起
	if got := a.Next([]string{"My App"}); got != "my-app-2" {
		t.Fatalf("second Next = %q, want %q", got, "my-app-2")
	}
	if got := a.Next([]string{"My App"}); got != "my-app-3" {
		t.Fatalf("third Next = %q, want %q", got, "my-app-3")
	}
	// 不同名独立计数
	if got := a.Next([]string{"Other"}); got != "other" {
		t.Fatalf("other Next = %q, want %q", got, "other")
	}
}

func TestIDAllocatorCandidatesLastNonEmpty(t *testing.T) {
	a := &idAllocator{}
	// 遍历候选，取最后一个非空 slug
	if got := a.Next([]string{"", "   ", "Foo", "Bar"}); got != "bar" {
		t.Fatalf("candidate Next = %q, want %q", got, "bar")
	}
	if got := a.Next([]string{"Foo", "Bar"}); got != "bar-2" {
		t.Fatalf("candidate second = %q, want %q", got, "bar-2")
	}
}

func TestIDAllocatorSlugDedup(t *testing.T) {
	a := &idAllocator{}
	if got := a.Next([]string{"App One"}); got != "app-one" {
		t.Fatalf("first alloc %q", got)
	}
	if got := a.Next([]string{"app-one"}); got != "app-one-2" {
		t.Fatalf("same slug must share counter, got %q", got)
	}
	if got := a.Next([]string{" App   One "}); got != "app-one-3" {
		t.Fatalf("trimmed same slug must share counter, got %q", got)
	}
	if len(a.counts) != 1 {
		t.Fatalf("expected 1 slug entry, got %d", len(a.counts))
	}
}

func TestIDAllocatorEmptyFallback(t *testing.T) {
	a := &idAllocator{}
	// 全空候选 → 兜底 key "item"
	if got := a.Next([]string{"   ", ""}); got != "item" {
		t.Fatalf("empty name Next = %q, want %q", got, "item")
	}
	if got := a.Next([]string{"", ""}); got != "item-2" {
		t.Fatalf("empty name second Next = %q, want %q", got, "item-2")
	}
}
