// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package activitypub

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"code.gitea.io/gitea/modules/activitypub"
	gitea_context "code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"

	ap "github.com/go-ap/activitypub"
	"github.com/go-fed/httpsig"
)

func getPublicKeyFromResponse(b []byte, keyID *url.URL) (p crypto.PublicKey, err error) {
	person := ap.PersonNew(ap.IRI(keyID.String()))
	err = person.UnmarshalJSON(b)
	if err != nil {
		err = fmt.Errorf("ActivityStreams type cannot be converted to one known to have publicKey property: %v", err)
		return
	}
	pubKey := person.PublicKey
	if pubKey.ID.String() != keyID.String() {
		err = fmt.Errorf("cannot find publicKey with id: %s in %s", keyID, string(b))
		return
	}
	pubKeyPem := pubKey.PublicKeyPem
	block, _ := pem.Decode([]byte(pubKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		err = fmt.Errorf("could not decode publicKeyPem to PUBLIC KEY pem block type")
		return
	}
	p, err = x509.ParsePKIXPublicKey(block.Bytes)
	return
}

func verifyHTTPSignatures(ctx *gitea_context.APIContext) (authenticated bool, err error) {
	r := ctx.Req

	// 1. Figure out what key we need to verify
	v, err := httpsig.NewVerifier(r)
	if err != nil {
		return
	}
	ID := v.KeyId()
	idIRI, err := url.Parse(ID)
	if err != nil {
		return
	}
	// 2. Fetch the public key of the other actor
	b, err := activitypub.Fetch(idIRI)
	if err != nil {
		return
	}
	pubKey, err := getPublicKeyFromResponse(b, idIRI)
	if err != nil {
		return
	}
	// 3. Verify the other actor's key
	algo := httpsig.Algorithm(setting.Federation.Algorithms[0])
	authenticated = v.Verify(pubKey, algo) == nil
	return
}

// ReqHTTPSignature function
func ReqHTTPSignature() func(ctx *gitea_context.APIContext) {
	return func(ctx *gitea_context.APIContext) {
		if authenticated, err := verifyHTTPSignatures(ctx); err != nil {
			ctx.Error(http.StatusInternalServerError, "verifyHttpSignatures", err)
		} else if !authenticated {
			ctx.Error(http.StatusForbidden, "reqSignature", "request signature verification failed")
		}
	}
}
