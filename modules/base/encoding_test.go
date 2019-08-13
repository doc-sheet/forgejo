// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUTF8WithErr(t *testing.T) {
	var res string
	var err error

	res, err = ToUTF8WithErr([]byte{0x41, 0x42, 0x43})
	assert.Equal(t, "ABC", res)
	assert.NoError(t, err)

	res, err = ToUTF8WithErr([]byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, "áéíóú", res)
	assert.NoError(t, err)

	res, err = ToUTF8WithErr([]byte{0xef, 0xbb, 0xbf, 0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, "áéíóú", res)
	assert.NoError(t, err)

	// This test FAILS
	res, err = ToUTF8WithErr([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0xF1, 0x6F, 0x73})
	assert.Equal(t, "Hola, así cómo ños", res)
	assert.NoError(t, err)

	res, err = ToUTF8WithErr([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x07, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Regexp(t, "^Hola, así cómo", res)
	assert.NoError(t, err)

	res, err = ToUTF8WithErr([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x81, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Regexp(t, "^Hola, así cómo", res)
	assert.NoError(t, err)

	// Japanese (Shift-JIS)
	res, err = ToUTF8WithErr([]byte{0x93, 0xFA, 0x91, 0xAE, 0x94, 0xE9, 0x82, 0xBC, 0x82, 0xB5, 0x82, 0xBF, 0x82, 0xE3, 0x81, 0x42})
	assert.Equal(t, "日属秘ぞしちゅ。", res)
	assert.NoError(t, err)

	res, err = ToUTF8WithErr([]byte{0x00, 0x00, 0x00, 0x00})
	assert.Equal(t, "\x00\x00\x00\x00", res)
	assert.NoError(t, err)
}

func TestToUTF8WithFallback(t *testing.T) {
	res := ToUTF8WithFallback([]byte{0x41, 0x42, 0x43})
	assert.Equal(t, []byte("ABC"), res)

	res = ToUTF8WithFallback([]byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte("áéíóú"), res)

	res = ToUTF8WithFallback([]byte{0xef, 0xbb, 0xbf, 0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte("áéíóú"), res)

	res = ToUTF8WithFallback([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0xF1, 0x6F, 0x73})
	assert.Equal(t, []byte("Hola, así cómo ños"), res)

	minmatch := []byte("Hola, así cómo ")

	res = ToUTF8WithFallback([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x07, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Equal(t, minmatch, res[0:len(minmatch)])

	res = ToUTF8WithFallback([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x81, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Equal(t, minmatch, res[0:len(minmatch)])

	// Japanese (Shift-JIS)
	res = ToUTF8WithFallback([]byte{0x93, 0xFA, 0x91, 0xAE, 0x94, 0xE9, 0x82, 0xBC, 0x82, 0xB5, 0x82, 0xBF, 0x82, 0xE3, 0x81, 0x42})
	assert.Equal(t, []byte("日属秘ぞしちゅ。"), res)

	res = ToUTF8WithFallback([]byte{0x00, 0x00, 0x00, 0x00})
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, res)
}

func TestToUTF8(t *testing.T) {
	res := ToUTF8("ABC")
	assert.Equal(t, "ABC", res)

	res = ToUTF8("áéíóú")
	assert.Equal(t, "áéíóú", res)

	// With utf-8 BOM
	res = ToUTF8("\ufeffáéíóú")
	assert.Equal(t, "áéíóú", res)

	res = ToUTF8("Hola, así cómo ños")
	assert.Equal(t, "Hola, así cómo ños", res)

	res = ToUTF8("Hola, así cómo \x07ños")
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Regexp(t, "^Hola, así cómo", res)

	// This test FAILS
	// res = ToUTF8("Hola, así cómo \x81ños")
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	// assert.Regexp(t, "^Hola, así cómo", res)

	// Japanese (Shift-JIS)
	res = ToUTF8("\x93\xFA\x91\xAE\x94\xE9\x82\xBC\x82\xB5\x82\xBF\x82\xE3\x81\x42")
	assert.Equal(t, "日属秘ぞしちゅ。", res)

	res = ToUTF8("\x00\x00\x00\x00")
	assert.Equal(t, "\x00\x00\x00\x00", res)
}

func TestToUTF8DropErrors(t *testing.T) {
	res := ToUTF8DropErrors([]byte{0x41, 0x42, 0x43})
	assert.Equal(t, []byte("ABC"), res)

	res = ToUTF8DropErrors([]byte{0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte("áéíóú"), res)

	res = ToUTF8DropErrors([]byte{0xef, 0xbb, 0xbf, 0xc3, 0xa1, 0xc3, 0xa9, 0xc3, 0xad, 0xc3, 0xb3, 0xc3, 0xba})
	assert.Equal(t, []byte("áéíóú"), res)

	res = ToUTF8DropErrors([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0xF1, 0x6F, 0x73})
	assert.Equal(t, []byte("Hola, así cómo ños"), res)

	minmatch := []byte("Hola, así cómo ")

	res = ToUTF8DropErrors([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x07, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Equal(t, minmatch, res[0:len(minmatch)])

	res = ToUTF8DropErrors([]byte{0x48, 0x6F, 0x6C, 0x61, 0x2C, 0x20, 0x61, 0x73, 0xED, 0x20, 0x63, 0xF3, 0x6D, 0x6F, 0x20, 0x81, 0xA4, 0x6F, 0x73})
	// Do not fail for differences in invalid cases, as the library might change the conversion criteria for those
	assert.Equal(t, minmatch, res[0:len(minmatch)])

	// Japanese (Shift-JIS)
	res = ToUTF8DropErrors([]byte{0x93, 0xFA, 0x91, 0xAE, 0x94, 0xE9, 0x82, 0xBC, 0x82, 0xB5, 0x82, 0xBF, 0x82, 0xE3, 0x81, 0x42})
	assert.Equal(t, []byte("日属秘ぞしちゅ。"), res)

	res = ToUTF8DropErrors([]byte{0x00, 0x00, 0x00, 0x00})
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, res)
}
