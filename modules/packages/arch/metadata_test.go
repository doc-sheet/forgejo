// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package arch

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"code.gitea.io/gitea/modules/packages"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"
)

func TestParsePackage(t *testing.T) {
	// Minimal PKGINFO contents and test FS
	const PKGINFO = `pkgname = a
pkgbase = b
pkgver = 1-2
arch = x86_64
`
	fs := fstest.MapFS{
		"pkginfo": &fstest.MapFile{
			Data:    []byte(PKGINFO),
			Mode:    os.ModePerm,
			ModTime: time.Now(),
		},
		"mtree": &fstest.MapFile{
			Data:    []byte("data"),
			Mode:    os.ModePerm,
			ModTime: time.Now(),
		},
	}

	// Test .PKGINFO file
	pinf, err := fs.Stat("pkginfo")
	require.NoError(t, err)

	pfile, err := fs.Open("pkginfo")
	require.NoError(t, err)

	parcname, err := archiver.NameInArchive(pinf, ".PKGINFO", ".PKGINFO")
	require.NoError(t, err)

	// Test .MTREE file
	minf, err := fs.Stat("mtree")
	require.NoError(t, err)

	mfile, err := fs.Open("mtree")
	require.NoError(t, err)

	marcname, err := archiver.NameInArchive(minf, ".MTREE", ".MTREE")
	require.NoError(t, err)

	t.Run("normal archive", func(t *testing.T) {
		var buf bytes.Buffer

		archive := archiver.NewTarZstd()
		archive.Create(&buf)

		err = archive.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   pinf,
				CustomName: parcname,
			},
			ReadCloser: pfile,
		})
		require.NoError(t, errors.Join(pfile.Close(), err))

		err = archive.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   minf,
				CustomName: marcname,
			},
			ReadCloser: mfile,
		})
		require.NoError(t, errors.Join(mfile.Close(), archive.Close(), err))

		reader, err := packages.CreateHashedBufferFromReader(&buf)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		_, err = ParsePackage(reader)

		require.NoError(t, err)
	})

	t.Run("missing .PKGINFO", func(t *testing.T) {
		var buf bytes.Buffer

		archive := archiver.NewTarZstd()
		archive.Create(&buf)
		require.NoError(t, archive.Close())

		reader, err := packages.CreateHashedBufferFromReader(&buf)
		require.NoError(t, err)

		defer reader.Close()
		_, err = ParsePackage(reader)

		require.Error(t, err)
		require.Contains(t, err.Error(), ".PKGINFO file not found")
	})

	t.Run("missing .MTREE", func(t *testing.T) {
		var buf bytes.Buffer

		pfile, err := fs.Open("pkginfo")
		require.NoError(t, err)

		archive := archiver.NewTarZstd()
		archive.Create(&buf)

		err = archive.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   pinf,
				CustomName: parcname,
			},
			ReadCloser: pfile,
		})
		require.NoError(t, errors.Join(pfile.Close(), archive.Close(), err))
		reader, err := packages.CreateHashedBufferFromReader(&buf)
		require.NoError(t, err)

		defer reader.Close()
		_, err = ParsePackage(reader)

		require.Error(t, err)
		require.Contains(t, err.Error(), ".MTREE file not found")
	})
}

func TestParsePackageInfo(t *testing.T) {
	const PKGINFO = `# Generated by makepkg 6.0.2
# using fakeroot version 1.31
pkgname = a
pkgbase = b
pkgver = 1-2
pkgdesc = comment
url = https://example.com/
group = group
builddate = 3
packager = Name Surname <login@example.com>
size = 5
arch = x86_64
license = BSD
provides = pvd
depend = smth
optdepend = hex
checkdepend = ola
makedepend = cmake
backup = usr/bin/paket1
`
	p, err := ParsePackageInfo("zst", strings.NewReader(PKGINFO))
	require.NoError(t, err)
	require.Equal(t, Package{
		CompressType: "zst",
		Name:         "a",
		Version:      "1-2",
		VersionMetadata: VersionMetadata{
			Base:         "b",
			Description:  "comment",
			ProjectURL:   "https://example.com/",
			Groups:       []string{"group"},
			Provides:     []string{"pvd"},
			License:      []string{"BSD"},
			Depends:      []string{"smth"},
			OptDepends:   []string{"hex"},
			MakeDepends:  []string{"cmake"},
			CheckDepends: []string{"ola"},
			Backup:       []string{"usr/bin/paket1"},
		},
		FileMetadata: FileMetadata{
			InstalledSize: 5,
			BuildDate:     3,
			Packager:      "Name Surname <login@example.com>",
			Arch:          "x86_64",
		},
	}, *p)
}

func TestValidatePackageSpec(t *testing.T) {
	newpkg := func() Package {
		return Package{
			Name:    "abc",
			Version: "1-1",
			VersionMetadata: VersionMetadata{
				Base:         "ghx",
				Description:  "whoami",
				ProjectURL:   "https://example.com/",
				Groups:       []string{"gnome"},
				Provides:     []string{"abc", "def"},
				License:      []string{"GPL"},
				Depends:      []string{"go", "gpg=1", "curl>=3", "git<=7"},
				OptDepends:   []string{"git", "libgcc=1.0", "gzip>1.0", "gz>=1.0", "lz<1.0", "gzip<=1.0", "zstd>1.0:foo bar<test>"},
				MakeDepends:  []string{"chrom"},
				CheckDepends: []string{"bariy"},
				Backup:       []string{"etc/pacman.d/filo"},
			},
			FileMetadata: FileMetadata{
				CompressedSize: 1,
				InstalledSize:  2,
				SHA256:         "def",
				BuildDate:      3,
				Packager:       "smon",
				Arch:           "x86_64",
			},
		}
	}

	t.Run("valid package", func(t *testing.T) {
		p := newpkg()

		err := ValidatePackageSpec(&p)

		require.NoError(t, err)
	})

	t.Run("invalid package name", func(t *testing.T) {
		p := newpkg()
		p.Name = "!$%@^!*&()"

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid package name")
	})

	t.Run("invalid package base", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.Base = "!$%@^!*&()"

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid package base")
	})

	t.Run("invalid package version", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.Base = "una-luna?"

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid package base")
	})

	t.Run("invalid package version", func(t *testing.T) {
		p := newpkg()
		p.Version = "una-luna"

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid package version")
	})

	t.Run("missing architecture", func(t *testing.T) {
		p := newpkg()
		p.FileMetadata.Arch = ""

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "architecture should be specified")
	})

	t.Run("invalid URL", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.ProjectURL = "http%%$#"

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid project URL")
	})

	t.Run("invalid check dependency", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.CheckDepends = []string{"Err^_^"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid check dependency")
	})

	t.Run("invalid dependency", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.Depends = []string{"^^abc"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid dependency")
	})

	t.Run("invalid make dependency", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.MakeDepends = []string{"^m^"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid make dependency")
	})

	t.Run("invalid provides", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.Provides = []string{"^m^"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid provides")
	})

	t.Run("invalid optional dependency", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.OptDepends = []string{"^m^:MM"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid optional dependency")
	})

	t.Run("invalid optional dependency", func(t *testing.T) {
		p := newpkg()
		p.VersionMetadata.Backup = []string{"/ola/cola"}

		err := ValidatePackageSpec(&p)

		require.Error(t, err)
		require.Contains(t, err.Error(), "backup file contains leading forward slash")
	})
}

func TestDescString(t *testing.T) {
	const pkgdesc = `%FILENAME%
zstd-1.5.5-1-x86_64.pkg.tar.zst

%NAME%
zstd

%BASE%
zstd

%VERSION%
1.5.5-1

%DESC%
Zstandard - Fast real-time compression algorithm

%GROUPS%
dummy1
dummy2

%CSIZE%
401

%ISIZE%
1500453

%MD5SUM%
5016660ef3d9aa148a7b72a08d3df1b2

%SHA256SUM%
9fa4ede47e35f5971e4f26ecadcbfb66ab79f1d638317ac80334a3362dedbabd

%URL%
https://facebook.github.io/zstd/

%LICENSE%
BSD
GPL2

%ARCH%
x86_64

%BUILDDATE%
1681646714

%PACKAGER%
Jelle van der Waa <jelle@archlinux.org>

%PROVIDES%
libzstd.so=1-64

%DEPENDS%
glibc
gcc-libs
zlib
xz
lz4

%OPTDEPENDS%
dummy3
dummy4

%MAKEDEPENDS%
cmake
gtest
ninja

%CHECKDEPENDS%
dummy5
dummy6

`

	md := &Package{
		CompressType: "zst",
		Name:         "zstd",
		Version:      "1.5.5-1",
		VersionMetadata: VersionMetadata{
			Base:         "zstd",
			Description:  "Zstandard - Fast real-time compression algorithm",
			ProjectURL:   "https://facebook.github.io/zstd/",
			Groups:       []string{"dummy1", "dummy2"},
			Provides:     []string{"libzstd.so=1-64"},
			License:      []string{"BSD", "GPL2"},
			Depends:      []string{"glibc", "gcc-libs", "zlib", "xz", "lz4"},
			OptDepends:   []string{"dummy3", "dummy4"},
			MakeDepends:  []string{"cmake", "gtest", "ninja"},
			CheckDepends: []string{"dummy5", "dummy6"},
		},
		FileMetadata: FileMetadata{
			CompressedSize: 401,
			InstalledSize:  1500453,
			MD5:            "5016660ef3d9aa148a7b72a08d3df1b2",
			SHA256:         "9fa4ede47e35f5971e4f26ecadcbfb66ab79f1d638317ac80334a3362dedbabd",
			BuildDate:      1681646714,
			Packager:       "Jelle van der Waa <jelle@archlinux.org>",
			Arch:           "x86_64",
		},
	}
	require.Equal(t, pkgdesc, md.Desc())
}
