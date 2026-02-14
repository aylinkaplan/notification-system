package ptr

import "testing"

func TestString(t *testing.T) {
	s := "hello"
	p := String(s)
	if p == nil {
		t.Fatal("String returned nil")
	}
	if *p != s {
		t.Errorf("String(%q) = %q", s, *p)
	}
}

func TestInt(t *testing.T) {
	i := 42
	p := Int(i)
	if p == nil {
		t.Fatal("Int returned nil")
	}
	if *p != i {
		t.Errorf("Int(%d) = %d", i, *p)
	}
}
