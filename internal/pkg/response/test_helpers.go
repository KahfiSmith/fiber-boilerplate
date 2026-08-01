package response

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return string(body)
}
