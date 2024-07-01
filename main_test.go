package main

import "testing"

func TestGetMaroto(t *testing.T) {
	username := "xhermitx"

	m := GetMaroto(username)

	document, err := m.Generate()
	if err != nil {
		t.Fatal("Could not generate Document")
	}

	if err := document.Save("assets/xhermitx.pdf"); err != nil {
		t.Fatal(err)
	}
}
