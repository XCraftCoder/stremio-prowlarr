package addon

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTitleCheck(t *testing.T) {
	t.Logf("Diff: %f", checkTitleSimilarity("House", "Winter House"))
	t.Logf("Diff: %f", checkTitleSimilarity("House", "House_-_"))
	t.Logf("Diff: %f", checkTitleSimilarity("Mad Max: Fury Road", "Mad Max Fury Road"))
}

func Test_Encoding(t *testing.T) {
	userData := &UserData{
		RDAPIKey: "RD API TOKEN",
	}

	jsonEncoded, err := json.Marshal(userData)
	require.NoError(t, err)
	base64Encoded := base64.RawURLEncoding.EncodeToString(jsonEncoded)
	t.Logf("Encoded: %s", base64Encoded)
}
