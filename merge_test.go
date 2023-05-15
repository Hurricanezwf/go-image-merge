package goimagemerge_test

import (
	"image/jpeg"
	"os"
	"testing"

	goimagemerge "github.com/ozankasikci/go-image-merge"
)

func TestMergeWithRemoteImage(t *testing.T) {
	m := goimagemerge.NewWithRemoteImages([]string{
		"https://cos-xica-prod.tiamat.world/user/t11OW_xpCbbtYoGM8Abao/createdimage/MLmfx5SoVBb4K6DIQaEbb.png?x-image-process=style/scale",
		"https://cos-xica-prod.tiamat.world/user/t11OW_xpCbbtYoGM8Abao/createdimage/LnHL9qUH702m0Jn5DquTn.png?x-image-process=style/scale",
		"https://cos-xica-prod.tiamat.world/user/t11OW_xpCbbtYoGM8Abao/createdimage/6wND7zVeOicsTAIiat8gt.png?x-image-process=style/scale",
		"https://cos-xica-prod.tiamat.world/user/t11OW_xpCbbtYoGM8Abao/createdimage/GmnGCk8XgzZiHjazEg7tX.png?x-image-process=style/scale",
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
