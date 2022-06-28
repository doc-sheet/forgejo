// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package i18n

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"

	"gopkg.in/ini.v1"
)

var (
	ErrLocaleAlreadyExist = errors.New("lang already exists")

	DefaultLocales = NewLocaleStore(true)
)

type locale struct {
	store    *LocaleStore
	langName string
	textMap  map[int]string // the map key (idx) is generated by store's textIdxMap

	sourceFileName      string
	sourceFileInfo      os.FileInfo
	lastReloadCheckTime time.Time
}

type LocaleStore struct {
	reloadMu *sync.RWMutex // for non-prod(dev), use a mutex for live-reload. for prod, no mutex, no live-reload.

	langNames []string
	langDescs []string

	localeMap  map[string]*locale
	textIdxMap map[string]int

	defaultLang string
}

func NewLocaleStore(isProd bool) *LocaleStore {
	ls := &LocaleStore{localeMap: make(map[string]*locale), textIdxMap: make(map[string]int)}
	if !isProd {
		ls.reloadMu = &sync.RWMutex{}
	}
	return ls
}

// AddLocaleByIni adds locale by ini into the store
// if source is a string, then the file is loaded. in dev mode, the file can be live-reloaded
// if source is a []byte, then the content is used
func (ls *LocaleStore) AddLocaleByIni(langName, langDesc string, source interface{}) error {
	if _, ok := ls.localeMap[langName]; ok {
		return ErrLocaleAlreadyExist
	}

	lc := &locale{store: ls, langName: langName}
	if fileName, ok := source.(string); ok {
		lc.sourceFileName = fileName
		lc.sourceFileInfo, _ = os.Stat(fileName) // live-reload only works for regular files. the error can be ignored
	}

	ls.langNames = append(ls.langNames, langName)
	ls.langDescs = append(ls.langDescs, langDesc)
	ls.localeMap[lc.langName] = lc

	return ls.reloadLocaleByIni(langName, source)
}

func (ls *LocaleStore) reloadLocaleByIni(langName string, source interface{}) error {
	iniFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment:         true,
		UnescapeValueCommentSymbols: true,
	}, source)
	if err != nil {
		return fmt.Errorf("unable to load ini: %w", err)
	}
	iniFile.BlockMode = false

	lc := ls.localeMap[langName]
	lc.textMap = make(map[int]string)
	for _, section := range iniFile.Sections() {
		for _, key := range section.Keys() {
			var trKey string
			if section.Name() == "" || section.Name() == "DEFAULT" {
				trKey = key.Name()
			} else {
				trKey = section.Name() + "." + key.Name()
			}
			textIdx, ok := ls.textIdxMap[trKey]
			if !ok {
				textIdx = len(ls.textIdxMap)
				ls.textIdxMap[trKey] = textIdx
			}
			lc.textMap[textIdx] = key.Value()
		}
	}
	iniFile = nil
	return nil
}

func (ls *LocaleStore) HasLang(langName string) bool {
	_, ok := ls.localeMap[langName]
	return ok
}

func (ls *LocaleStore) ListLangNameDesc() (names, desc []string) {
	return ls.langNames, ls.langDescs
}

// SetDefaultLang sets default language as a fallback
func (ls *LocaleStore) SetDefaultLang(lang string) {
	ls.defaultLang = lang
}

// Tr translates content to target language. fall back to default language.
func (ls *LocaleStore) Tr(lang, trKey string, trArgs ...interface{}) string {
	l, ok := ls.localeMap[lang]
	if !ok {
		l, ok = ls.localeMap[ls.defaultLang]
	}
	if ok {
		return l.Tr(trKey, trArgs...)
	}
	return trKey
}

// Tr translates content to locale language. fall back to default language.
func (l *locale) Tr(trKey string, trArgs ...interface{}) string {
	if l.store.reloadMu != nil {
		l.store.reloadMu.RLock()
		defer l.store.reloadMu.RUnlock()
	}
	msg, _ := l.tryTr(trKey, trArgs...)
	return msg
}

func (l *locale) tryTr(trKey string, trArgs ...interface{}) (msg string, found bool) {
	if l.store.reloadMu != nil {
		now := time.Now()
		if now.Sub(l.lastReloadCheckTime) >= time.Second && l.sourceFileInfo != nil && l.sourceFileName != "" {
			l.store.reloadMu.RUnlock() // if the locale file should be reloaded, then we release the read-lock
			l.store.reloadMu.Lock()    // and acquire the write-lock
			l.lastReloadCheckTime = now
			if sourceFileInfo, err := os.Stat(l.sourceFileName); err == nil && !sourceFileInfo.ModTime().Equal(l.sourceFileInfo.ModTime()) {
				if err = l.store.reloadLocaleByIni(l.langName, l.sourceFileName); err == nil {
					l.sourceFileInfo = sourceFileInfo
				} else {
					log.Error("unable to live-reload the locale file %q, err: %v", l.sourceFileName, err)
				}
			}
			l.store.reloadMu.Unlock() // release the write-lock
			l.store.reloadMu.RLock()  // and re-acquire the read-lock, which was managed by outer Tr function
		}
	}
	trMsg := trKey
	textIdx, ok := l.store.textIdxMap[trKey]
	if ok {
		if msg, found = l.textMap[textIdx]; found {
			trMsg = msg // use current translation
		} else if l.langName != l.store.defaultLang {
			if def, ok := l.store.localeMap[l.store.defaultLang]; ok {
				return def.tryTr(trKey, trArgs...)
			}
		} else if !setting.IsProd {
			log.Error("missing i18n translation key: %q", trKey)
		}
	}

	if len(trArgs) > 0 {
		fmtArgs := make([]interface{}, 0, len(trArgs))
		for _, arg := range trArgs {
			val := reflect.ValueOf(arg)
			if val.Kind() == reflect.Slice {
				// before, it can accept Tr(lang, key, a, [b, c], d, [e, f]) as Sprintf(msg, a, b, c, d, e, f), it's an unstable behavior
				// now, we restrict the strange behavior and only support:
				// 1. Tr(lang, key, [slice-items]) as Sprintf(msg, items...)
				// 2. Tr(lang, key, args...) as Sprintf(msg, args...)
				if len(trArgs) == 1 {
					for i := 0; i < val.Len(); i++ {
						fmtArgs = append(fmtArgs, val.Index(i).Interface())
					}
				} else {
					log.Error("the args for i18n shouldn't contain uncertain slices, key=%q, args=%v", trKey, trArgs)
					break
				}
			} else {
				fmtArgs = append(fmtArgs, arg)
			}
		}
		return fmt.Sprintf(trMsg, fmtArgs...), found
	}
	return trMsg, found
}

// ResetDefaultLocales resets the current default locales
// NOTE: this is not synchronized
func ResetDefaultLocales(isProd bool) {
	DefaultLocales = NewLocaleStore(isProd)
}

// Tr use default locales to translate content to target language.
func Tr(lang, trKey string, trArgs ...interface{}) string {
	return DefaultLocales.Tr(lang, trKey, trArgs...)
}
