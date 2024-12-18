package source

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	billyUtils "github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
)

func TestCopyToMemFS(t *testing.T) {
	tests := map[string]struct {
		setupFS  func() fs.FS
		srcPath  string
		destPath string
		wantErr  bool
		validate func(t *testing.T, memFS filesys.FileSystem)
	}{
		"copy single file": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = billyUtils.WriteFile(memFS, "test.yaml", []byte("test content"), 0666)
				return fs.Create(memFS)
			},
			srcPath:  "/",
			destPath: "/",
			wantErr:  false,
			validate: func(t *testing.T, memFS filesys.FileSystem) {
				content, err := memFS.ReadFile("test.yaml")
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			},
		},
		"copy directory structure": {
			setupFS: func() fs.FS {
				memFS := memfs.New()
				_ = memFS.MkdirAll("dir/subdir", 0666)
				_ = billyUtils.WriteFile(memFS, "dir/test1.yaml", []byte("content1"), 0666)
				_ = billyUtils.WriteFile(memFS, "dir/subdir/test2.yaml", []byte("content2"), 0666)
				return fs.Create(memFS)
			},
			srcPath:  "dir",
			destPath: "/target",
			wantErr:  false,
			validate: func(t *testing.T, memFS filesys.FileSystem) {
				content1, err := memFS.ReadFile("/target/test1.yaml")
				assert.NoError(t, err)
				assert.Equal(t, "content1", string(content1))

				content2, err := memFS.ReadFile("/target/subdir/test2.yaml")
				assert.NoError(t, err)
				assert.Equal(t, "content2", string(content2))
			},
		},
		"source directory not found": {
			setupFS: func() fs.FS {
				return fs.Create(memfs.New())
			},
			srcPath:  "/nonexistent",
			destPath: "/",
			wantErr:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.setupFS()
			memFS := filesys.MakeFsInMemory()

			err := copyToMemFS(repofs, tt.srcPath, tt.destPath, memFS)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, memFS)
			}
		})
	}
}
