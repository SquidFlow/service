package writer

import (
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"

	"github.com/squidflow/service/pkg/fs"
	"github.com/squidflow/service/pkg/store"
)

func TestDeleteFromProject(t *testing.T) {
	tests := map[string]struct {
		wantErr  string
		beforeFn func() fs.FS
		assertFn func(*testing.T, fs.FS)
	}{
		"Should remove entire app folder, if it contains only one overlay": {
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				return fs.Create(memfs)
			},
			assertFn: func(t *testing.T, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)))
			},
		},
		"Should delete just the overlay, if there are more": {
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project"), 0666)
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				return fs.Create(memfs)
			},
			assertFn: func(t *testing.T, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir)))
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project")))
			},
		},
		"Should remove directory apps": {
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", "project"), 0666)
				return fs.Create(memfs)
			},
			assertFn: func(t *testing.T, repofs fs.FS) {
				assert.False(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
		"Should not delete anything, if kust app is not in project": {
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", store.Default.OverlaysDir, "project2"), 0666)
				return fs.Create(memfs)
			},
			assertFn: func(t *testing.T, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
		"Should not delete anything, if dir app is not in project": {
			beforeFn: func() fs.FS {
				memfs := memfs.New()
				_ = memfs.MkdirAll(filepath.Join(store.Default.AppsDir, "app", "project2"), 0666)
				return fs.Create(memfs)
			},
			assertFn: func(t *testing.T, repofs fs.FS) {
				assert.True(t, repofs.ExistsOrDie(filepath.Join(store.Default.AppsDir, "app")))
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repofs := tt.beforeFn()
			if err := DeleteFromProject(repofs, "app", "project"); err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if tt.assertFn != nil {
				tt.assertFn(t, repofs)
			}
		})
	}
}
