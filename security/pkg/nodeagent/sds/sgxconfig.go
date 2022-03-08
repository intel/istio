// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sds

import (
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	CsrConfig = `
openssl_conf = openssl_def

[openssl_def]
engines = engine_section

[engine_section]
pkcs11 = pkcs11_section

[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
[req_distinguished_name]
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
URI.1 = {{ .SAN }}
[pkcs11_section]
engine_id = pkcs11
dynamic_path = /usr/lib/x86_64-linux-gnu/engines-1.1/libpkcs11.so
MODULE_PATH = /usr/local/lib/libp11sgx.so
PIN =`

	SgxLibrary       = "/usr/local/lib/libp11sgx.so"
	KeyLabel         = "default"
	RsaKeySize       = "2048"
	EcdsaKeyParam    = "P-256"
	Stage1           = "init"
	Stage2           = "cert"
	KeyType          = "rsa"
	SGXAnnotationKey = "sgx"
	ValidatorName    = "envoy.tls.cert_validator.extension"
	DefaultQuoteKey  = "1.3.6.1.4.1.54392.5.1283"
	DefaultPubKey    = "1.3.6.1.4.1.54392.5.1284"
)

var (
	UsrPin     = rand.String(10)
	SoPin      = rand.String(10)
	TokenLabel = rand.String(10)
)
