package images

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var minimalGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff,
	0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

type mockQuerier struct {
    *db.QuerierStub
    createFn func(context.Context, db.CreateUploadedImageForUploaderParams) (int64, error)
}

func (m *mockQuerier) CreateUploadedImageForUploader(ctx context.Context, arg db.CreateUploadedImageForUploaderParams) (int64, error) {
    if m.createFn != nil {
        return m.createFn(ctx, arg)
    }
    return 0, nil
}

func TestUploadImageTask_Action_SecurityFix(t *testing.T) {
	// Setup
	cfg := &config.RuntimeConfig{
		ImageMaxBytes:  1024 * 1024,
		ImageUploadDir: "/tmp/uploads", // Dummy dir
	}

	// Track the uploaded path to verify ID length
	var capturedPath string
	var capturedUploaderID int32

	stub := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("images", "upload", "post"),
	)

    mock := &mockQuerier{
        QuerierStub: stub,
        createFn: func(ctx context.Context, arg db.CreateUploadedImageForUploaderParams) (int64, error) {
            if arg.Path.Valid {
                capturedPath = arg.Path.String
            }
            capturedUploaderID = arg.UploaderID
            return 1, nil
        },
    }

	ctx := context.Background()
	cd := common.NewCoreData(ctx, mock, cfg)
	cd.UserID = 123

	// Create a multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.gif")
	require.NoError(t, err)
	_, err = part.Write(minimalGIF)
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Inject CoreData into context
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Create the task instance
	task := UploadImageTask{}

	res := task.Action(w, req)

	// Check response
	if errVal, ok := res.(error); ok {
		t.Fatalf("Action returned error: %v", errVal)
	}

	txt, ok := res.(handlers.TextByteWriter)
	if !ok {
		 t.Fatalf("Expected TextByteWriter, got %T", res)
	}

	respStr := string(txt)
	assert.Contains(t, respStr, "image:")

	require.NotEmpty(t, capturedPath, "CreateUploadedImageForUploader should be called")

	// /uploads/sub1/sub2/ID.gif
	parts := strings.Split(capturedPath, "/")
	filename := parts[len(parts)-1]
	// Remove extension
	id := strings.TrimSuffix(filename, ".gif")

	t.Logf("Generated ID: %s", id)

	// SHA256 length is 64
	assert.Equal(t, 64, len(id), "Expected SHA256 hash length (64) after fix")

	// Also verify UploaderID
	assert.Equal(t, int32(123), capturedUploaderID)
}
