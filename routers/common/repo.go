// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/charset"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/httpcache"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/typesniffer"
	"code.gitea.io/gitea/modules/util"
)

// ServeBlob serve git.Blob which represents a normal(non-lfs) file stored in repositories
// todo: implement io.Seeker for git.Blob.blobReader to support Range-Request
func ServeBlob(ctx *context.Context, blob *git.Blob, lastModified time.Time) error {
	if httpcache.HandleGenericETagTimeCache(ctx.Req, ctx.Resp, `"`+blob.ID.String()+`"`, lastModified) {
		return nil
	}

	dataRc, err := blob.DataAsync()
	if err != nil {
		return err
	}
	defer func() {
		if err = dataRc.Close(); err != nil {
			log.Error("ServeBlob: Close: %v", err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := util.ReadAtMost(dataRc, buf)
	if err != nil {
		return err
	}
	if n >= 0 {
		buf = buf[:n]
	}

	size := blob.Size()
	if size >= 0 {
		ctx.Resp.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	} else {
		log.Error("ServeData called to serve data: %s with size < 0: %d", ctx.Repo.TreePath, size)
	}

	if err := setCommonHeaders(ctx, ctx.Repo.TreePath, buf); err != nil {
		return err
	}

	_, err = ctx.Resp.Write(buf)
	if err != nil {
		return err
	}
	_, err = io.Copy(ctx.Resp, dataRc)
	return err
}

func setCommonHeaders(ctx *context.Context, name string, data interface{}) error {
	// Google Chrome dislike commas in filenames, so let's change it to a space
	name = strings.ReplaceAll(name, ",", " ")

	ctx.Resp.Header().Set("Cache-Control", "public, max-age=300")

	// reset the offset to the start of served file
	if seeker, ok := data.(io.ReadSeeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}

	st, err := typesniffer.DetectContentTypeExtFirst(name, data)
	if nil != err {
		return err
	}

	mappedMimeType := ""
	if setting.MimeTypeMap.Enabled {
		fileExtension := strings.ToLower(filepath.Ext(name))
		mappedMimeType = setting.MimeTypeMap.Map[fileExtension]
	}

	if st.IsText() || ctx.FormBool("render") {
		var cs string
		var err error
		if reader, ok := data.(io.ReadSeeker); ok {
			cs, err = charset.DetectEncodingFromReader(reader)
			_, _ = reader.Seek(0, io.SeekStart)
		} else {
			cs, err = charset.DetectEncoding(data.([]byte))
		}
		if err != nil {
			log.Error("Detect raw file %s charset failed: %v, using by default utf-8", name, err)
			cs = "utf-8"
		}

		if mappedMimeType == "" {
			mappedMimeType = "text/plain"
		}
		ctx.Resp.Header().Set("Content-Type", mappedMimeType+"; charset="+strings.ToLower(cs))

	} else {
		ctx.Resp.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		if mappedMimeType != "" {
			ctx.Resp.Header().Set("Content-Type", mappedMimeType)
		}
		if (st.IsImage() || st.IsPDF()) && (setting.UI.SVG.Enabled || !st.IsSvgImage()) {
			ctx.Resp.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, name))
			if st.IsSvgImage() || st.IsPDF() {
				ctx.Resp.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
				ctx.Resp.Header().Set("X-Content-Type-Options", "nosniff")
				if st.IsSvgImage() {
					ctx.Resp.Header().Set("Content-Type", typesniffer.SvgMimeType)
				} else {
					ctx.Resp.Header().Set("Content-Type", typesniffer.ApplicationOctetStream)
				}
			}
		} else {
			ctx.Resp.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
		}
	}

	return nil
}

// ServeLargeFile Serve files stored with Git LFS and attachments uploaded on the Releases page
func ServeLargeFile(ctx *context.Context, name string, time time.Time, reader io.ReadSeeker) error {
	if err := setCommonHeaders(ctx, name, reader); err != nil {
		return err
	}
	http.ServeContent(ctx.Resp, ctx.Req, name, time, reader)
	return nil
}
