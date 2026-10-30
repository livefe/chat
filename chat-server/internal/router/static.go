package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"app/pkg/storage"

	"github.com/gin-gonic/gin"
)

var staticMIMEByExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".3gp":  "video/3gpp",
	".webm": "video/webm",
	".m4a":  "audio/mp4",
	".mp3":  "audio/mpeg",
	".ogg":  "audio/ogg",
	".wav":  "audio/wav",
	".apk":  "application/vnd.android.package-archive",
}

func registerStaticRoutes(engine *gin.Engine) {
	if engine == nil {
		return
	}

	abs, err := filepath.Abs(storage.LocalRootDir)
	if err != nil {
		return
	}
	_ = os.MkdirAll(abs, 0o755)

	prefix := storage.LocalURLPrefix
	engine.Use(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, prefix+"/") && c.Request.URL.Path != prefix {
			c.Next()
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Range, Content-Type")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
		c.Header("Accept-Ranges", "bytes")
		c.Header("X-Content-Type-Options", "nosniff")
		if ext := strings.ToLower(filepath.Ext(c.Request.URL.Path)); ext != "" {
			if mime, ok := staticMIMEByExt[ext]; ok {
				c.Header("Content-Type", mime)
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	engine.Static(prefix, abs)
}
