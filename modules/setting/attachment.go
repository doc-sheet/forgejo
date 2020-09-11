// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"path/filepath"
	"strings"

	"code.gitea.io/gitea/modules/log"
)

var (
	// Attachment settings
	Attachment = struct {
		Storage
		Path         string
		AllowedTypes string
		MaxSize      int64
		MaxFiles     int
		Enabled      bool
	}{
		Storage: Storage{
			StoreType:   "local",
			ServeDirect: false,
		},
		AllowedTypes: "image/jpeg,image/png,application/zip,application/gzip",
		MaxSize:      4,
		MaxFiles:     5,
		Enabled:      true,
	}
)

func newAttachmentService() {
	sec := Cfg.Section("attachment")
	Attachment.StoreType = sec.Key("STORE_TYPE").MustString("")
	if Attachment.StoreType == "" {
		Attachment.StoreType = "default"
	}

	if Attachment.StoreType != "local" && Attachment.StoreType != "minio" {
		storage, ok := storages[Attachment.StoreType]
		if !ok {
			log.Fatal("Failed to get attachment storage type: %s", Attachment.StoreType)
		}
		Attachment.StoreType = storage.StoreType
		Attachment.ServeDirect = storage.ServeDirect
		Attachment.Minio = storage.Minio
	}

	// Override
	Attachment.ServeDirect = sec.Key("SERVE_DIRECT").MustBool(Attachment.ServeDirect)

	switch Attachment.StoreType {
	case "local":
		Attachment.Path = sec.Key("PATH").MustString(filepath.Join(AppDataPath, "attachments"))
		if !filepath.IsAbs(Attachment.Path) {
			Attachment.Path = filepath.Join(AppWorkPath, Attachment.Path)
		}
	case "minio":
		Attachment.Minio.Endpoint = sec.Key("MINIO_ENDPOINT").MustString(Attachment.Minio.Endpoint)
		Attachment.Minio.AccessKeyID = sec.Key("MINIO_ACCESS_KEY_ID").MustString(Attachment.Minio.AccessKeyID)
		Attachment.Minio.SecretAccessKey = sec.Key("MINIO_SECRET_ACCESS_KEY").MustString(Attachment.Minio.SecretAccessKey)
		Attachment.Minio.Bucket = sec.Key("MINIO_BUCKET").MustString(Attachment.Minio.Bucket)
		Attachment.Minio.Location = sec.Key("MINIO_LOCATION").MustString(Attachment.Minio.Location)
		Attachment.Minio.UseSSL = sec.Key("MINIO_USE_SSL").MustBool(Attachment.Minio.UseSSL)
		Attachment.Minio.BasePath = sec.Key("MINIO_BASE_PATH").MustString("attachments/")
	}

	Attachment.AllowedTypes = strings.Replace(sec.Key("ALLOWED_TYPES").MustString("image/jpeg,image/png,application/zip,application/gzip"), "|", ",", -1)
	Attachment.MaxSize = sec.Key("MAX_SIZE").MustInt64(4)
	Attachment.MaxFiles = sec.Key("MAX_FILES").MustInt(5)
	Attachment.Enabled = sec.Key("ENABLED").MustBool(true)
}
