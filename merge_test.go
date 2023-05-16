package goimagemerge_test

import (
	"image/jpeg"
	"os"
	"testing"

	goimagemerge "github.com/Hurricanezwf/go-image-merge"
)

func TestMergeWithRemoteImage(t *testing.T) {
	m := goimagemerge.NewWithRemoteImages([]string{
		// TODO: give me four image urls.
	}, 2, 2)

	image, err := m.Merge()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	f, err := os.OpenFile("/tmp/goimagemerge.jpg", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err = jpeg.Encode(f, image, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
