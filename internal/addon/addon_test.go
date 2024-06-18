package addon

import (
	"testing"
)

func TestTitleCheck(t *testing.T) {
	t.Logf("Diff: %f", checkTitleSimilarity("House", "Winter House"))
	t.Logf("Diff: %f", checkTitleSimilarity("House", "House_-_"))
	t.Logf("Diff: %f", checkTitleSimilarity("Mad Max: Fury Road", "Mad Max Fury Road"))
}
