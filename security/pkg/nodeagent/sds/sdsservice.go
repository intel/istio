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

// Package sds implements secret discovery service in NodeAgent.
package sds

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	cryptomb "github.com/envoyproxy/go-control-plane/contrib/envoy/extensions/private_key_providers/cryptomb/v3alpha"
	qat "github.com/envoyproxy/go-control-plane/contrib/envoy/extensions/private_key_providers/qat/v3alpha"
	sgx "github.com/envoyproxy/go-control-plane/contrib/envoy/extensions/private_key_providers/sgx/v3alpha"
	sgxtls "github.com/envoyproxy/go-control-plane/contrib/envoy/extensions/transport_sockets/tls/cert_validator/extension/v3alpha"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	sds "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/model"
	authn_model "istio.io/istio/pilot/pkg/security/model"
	"istio.io/istio/pilot/pkg/util/protoconv"
	"istio.io/istio/pilot/pkg/xds"
	v3 "istio.io/istio/pilot/pkg/xds/v3"
	"istio.io/istio/pkg/config/schema/kind"
	"istio.io/istio/pkg/security"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/pkg/test/util/tmpl"
	"istio.io/istio/pkg/util/sets"
	pkiutil "istio.io/istio/security/pkg/pki/util"
	"istio.io/pkg/log"
)

var sdsServiceLog = log.RegisterScope("sds", "SDS service debugging", 0)

type sdsservice struct {
	st                                security.SecretManager
	XdsServer                         *xds.DiscoveryServer
	stop                              chan struct{}
	rootCaPath                        string
	pkpConf                           *mesh.PrivateKeyProvider
	SgxEnabled                        bool
	SgxCertExtensionValidationEnabled bool
	CertificateReady                  bool
	sgxMutex                          sync.RWMutex
	useECDSA                          bool
	sanExtension                      string
	// callback function to invoke when there is a new CSR object.
	CSRCallback func(csr []byte) (bool, bool, error)
}

// Assert we implement the generator interface
var _ model.XdsResourceGenerator = &sdsservice{}

func NewXdsServer(stop chan struct{}, gen model.XdsResourceGenerator) *xds.DiscoveryServer {
	s := xds.NewXDS(stop)
	s.DiscoveryServer.Generators = map[string]model.XdsResourceGenerator{
		v3.SecretType: gen,
	}
	s.DiscoveryServer.ProxyNeedsPush = func(proxy *model.Proxy, req *model.PushRequest) bool {
		// Empty changes means "all"
		if len(req.ConfigsUpdated) == 0 {
			return true
		}
		var resources []string
		proxy.RLock()
		if proxy.WatchedResources[v3.SecretType] != nil {
			resources = proxy.WatchedResources[v3.SecretType].ResourceNames
		}
		proxy.RUnlock()

		if resources == nil {
			return false
		}

		names := sets.New(resources...)
		found := false
		for name := range model.ConfigsOfKind(req.ConfigsUpdated, kind.Secret) {
			if names.Contains(name.Name) {
				found = true
				break
			}
		}
		return found
	}
	s.DiscoveryServer.Start(stop)
	return s.DiscoveryServer
}

// newSDSService creates Secret Discovery Service which implements envoy SDS API.
func newSDSService(st security.SecretManager, options *security.Options, pkpConf *mesh.PrivateKeyProvider) *sdsservice {
	ret := &sdsservice{
		st:                                st,
		stop:                              make(chan struct{}),
		pkpConf:                           pkpConf,
		CertificateReady:                  false,
		SgxEnabled:                        options.SgxEnabled,
		SgxCertExtensionValidationEnabled: options.SgxCertExtensionValidationEnabled,
	}

	csrHostName := &spiffe.Identity{
		TrustDomain:    options.TrustDomain,
		Namespace:      options.WorkloadNamespace,
		ServiceAccount: options.ServiceAccount,
	}
	san := csrHostName.String()
	ret.sanExtension = san

	if options.ECCSigAlg == "ECDSA" {
		ret.useECDSA = true
	}
	ret.XdsServer = NewXdsServer(ret.stop, ret)

	ret.rootCaPath = options.CARootPath

	if options.FileMountedCerts {
		return ret
	}

	// Pre-generate workload certificates to improve startup latency and ensure that for OUTPUT_CERTS
	// case we always write a certificate. A workload can technically run without any mTLS/CA
	// configured, in which case this will fail; if it becomes noisy we should disable the entire SDS
	// server in these cases.
	go func() {
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 0
		for {
			_, err := st.GenerateSecret(security.WorkloadKeyCertResourceName)
			if err == nil {
				break
			}
			sdsServiceLog.Warnf("failed to warm certificate: %v", err)
			select {
			case <-ret.stop:
				return
			case <-time.After(b.NextBackOff()):
			}
		}
		for {
			_, err := st.GenerateSecret(security.RootCertReqResourceName)
			if err == nil {
				break
			}
			sdsServiceLog.Warnf("failed to warm root certificate: %v", err)
			select {
			case <-ret.stop:
				return
			case <-time.After(b.NextBackOff()):
			}
		}
	}()

	return ret
}

func (s *sdsservice) generate(resourceNames []string) (model.Resources, error) {
	s.sgxMutex.Lock()
	defer s.sgxMutex.Unlock()
	resources := model.Resources{}
	var res *anypb.Any
	// here if Envoy request stage "init"
	// just response the SDS config and waiting for external CSR
	// if Envoy request stage "cert" by gRPC request
	// call GenerateSecret() to generate cert and call toEnvoySecret() response with cert
	for _, resourceName := range resourceNames {
		fmt.Println("ResourceName: " + resourceName)
		secret, err := s.st.GenerateSecret(resourceName)
		if err != nil {
			// Typically, in Istiod, we do not return an error for a failure to generate a resource
			// However, here it makes sense, because we are generally streaming a single resource,
			// so sending an error will not cause a single failure to prevent the entire multiplex stream
			// of resources, and failures here are generally due to temporary networking issues to the CA
			// rather than a result of configuration issues, which trigger updates in Istiod when resolved.
			// Instead, we rely on the client to retry (with backoff) on failures.
			return nil, fmt.Errorf("failed to generate secret for %v: %v", resourceName, err)
		}
		if s.SgxCertExtensionValidationEnabled {
			fmt.Println("SgxCertExtensionValidationEnabled == true. Envoy will verify the SGX extension in the peer certificate")
		} else {
			fmt.Println("SgxCertExtensionValidationEnabled == false")
		}
		res = protoconv.MessageToAny(toEnvoySecret(secret, s.rootCaPath, s.SgxEnabled, s.CertificateReady, s.useECDSA,
			s.SgxCertExtensionValidationEnabled, s.sanExtension, s.pkpConf))
		resources = append(resources, &discovery.Resource{
			Name:     resourceName,
			Resource: res,
		})
	}
	return resources, nil
}

// Generate implements the XDS Generator interface. This allows the XDS server to dispatch requests
// for SecretTypeV3 to our server to generate the Envoy response.
func (s *sdsservice) Generate(proxy *model.Proxy, w *model.WatchedResource, updates *model.PushRequest) (model.Resources, model.XdsLogDetails, error) {
	// updates.Full indicates we should do a complete push of all updated resources
	// In practice, all pushes should be incremental (ie, if the `default` cert changes we won't push
	// all file certs).
	if updates.Full {
		resp, err := s.generate(w.ResourceNames)
		return resp, pushLog(w.ResourceNames), err
	}
	names := []string{}
	watched := sets.New(w.ResourceNames...)
	for i := range updates.ConfigsUpdated {
		if i.Kind == kind.Secret && watched.Contains(i.Name) {
			names = append(names, i.Name)
		}
	}
	resp, err := s.generate(names)
	return resp, pushLog(names), err
}

// register adds the SDS handle to the grpc server
func (s *sdsservice) register(rpcs *grpc.Server) {
	sds.RegisterSecretDiscoveryServiceServer(rpcs, s)
}

// StreamSecrets serves SDS discovery requests and SDS push requests
func (s *sdsservice) StreamSecrets(stream sds.SecretDiscoveryService_StreamSecretsServer) error {
	return s.XdsServer.Stream(stream)
}

func (s *sdsservice) DeltaSecrets(stream sds.SecretDiscoveryService_DeltaSecretsServer) error {
	return status.Error(codes.Unimplemented, "DeltaSecrets not implemented")
}

func (s *sdsservice) FetchSecrets(ctx context.Context, discReq *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "FetchSecrets not implemented")
}

func (s *sdsservice) SendCsrAndQuote(stream sds.SecretDiscoveryService_SendCsrAndQuoteServer) error {
	// Handle upstream SDS recv
	go func() {
		for {
			ctx, err := stream.Recv()
			if err != nil {
				sdsServiceLog.Infof(codes.NotFound, "Can't get CsrAndQuoteRequest")
				return
			}
			sdsServiceLog.Info("Received CsrAndQuoteRequest")

			if !s.SgxEnabled {
				sdsServiceLog.Info("As SGX is not enabled, Request in SendCsrAndQuote() is going to be ignored.")
				return
			}
			/*
				1. Get CSR and encode it to base64, as a byte[]
				2. validate this CSR
				3. Generate Workload Certificate according to the CSR
				3. retrun Cert to Envoy by SDS response.
			*/

			csr := ctx.GetCsr()
			needPush := false
			certReady := false

			s.sgxMutex.Lock()

			if len(csr) != 0 {
				needPush, certReady, _ = s.CSRCallback([]byte(ctx.Csr))
				if needPush && certReady {
					s.CertificateReady = true
				}
			}
			s.sgxMutex.Unlock()

			if s.CertificateReady {
				// Update Callback
				s.XdsServer.Push(&model.PushRequest{
					Full: false,
					ConfigsUpdated: map[model.ConfigKey]struct{}{
						{Kind: kind.Secret, Name: authn_model.SDSDefaultResourceName}: {},
					},
					Reason: []model.TriggerReason{model.SecretTrigger},
				})
				sdsServiceLog.Info("CertificateReady")
				return
			}
		}
	}()
	return nil
}

func (s *sdsservice) Close() {
	close(s.stop)
	s.XdsServer.Shutdown()
}

// toEnvoySecret converts a security.SecretItem to an Envoy tls.Secret
func toEnvoySecret(s *security.SecretItem, caRootPath string, sgxEnabled bool, certificateReady bool, useECDSA bool,
	sgxCertExtensionValidationEnabled bool, sanExtension string, pkpConf *mesh.PrivateKeyProvider) *tls.Secret {
	secret := &tls.Secret{
		Name: s.ResourceName,
	}
	var cfg security.SdsCertificateConfig
	ok := false
	if s.ResourceName == security.FileRootSystemCACert {
		cfg, ok = security.SdsCertificateConfigFromResourceNameForOSCACert(caRootPath)
	} else {
		cfg, ok = security.SdsCertificateConfigFromResourceName(s.ResourceName)
	}
	if s.ResourceName == security.RootCertReqResourceName || (ok && cfg.IsRootCertificate()) {
		if sgxCertExtensionValidationEnabled {
			ValidatorConfig := &sgxtls.ExtensionCertValidatorConfig{
				Extensions: []*sgxtls.ExtensionCertValidatorConfig_Extension{
					{
						Key:   DefaultQuoteKey,
						Value: pkiutil.ExtensionMessage,
					},
				},
			}
			msg, _ := anypb.New(ValidatorConfig)

			secret.Type = &tls.Secret_ValidationContext{
				ValidationContext: &tls.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: s.RootCert,
						},
					},
					CustomValidatorConfig: &core.TypedExtensionConfig{
						Name:        ValidatorName,
						TypedConfig: msg,
					},
				},
			}
		} else {
			secret.Type = &tls.Secret_ValidationContext{
				ValidationContext: &tls.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: s.RootCert,
						},
					},
				},
			}
		}
	} else {
		switch pkpConf.GetProvider().(type) {
		case *mesh.PrivateKeyProvider_Cryptomb:
			crypto := pkpConf.GetCryptomb()
			msg := protoconv.MessageToAny(&cryptomb.CryptoMbPrivateKeyMethodConfig{
				PollDelay: durationpb.New(time.Duration(crypto.GetPollDelay().Nanos)),
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{
						InlineBytes: s.PrivateKey,
					},
				},
			})
			secret.Type = &tls.Secret_TlsCertificate{
				TlsCertificate: &tls.TlsCertificate{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: s.CertificateChain,
						},
					},
					PrivateKeyProvider: &tls.PrivateKeyProvider{
						ProviderName: "cryptomb",
						ConfigType: &tls.PrivateKeyProvider_TypedConfig{
							TypedConfig: msg,
						},
					},
				},
			}
		case *mesh.PrivateKeyProvider_Qat:
			qatConf := pkpConf.GetQat()
			msg := protoconv.MessageToAny(&qat.QatPrivateKeyMethodConfig{
				PollDelay: durationpb.New(time.Duration(qatConf.GetPollDelay().Nanos)),
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{
						InlineBytes: s.PrivateKey,
					},
				},
			})
			secret.Type = &tls.Secret_TlsCertificate{
				TlsCertificate: &tls.TlsCertificate{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: s.CertificateChain,
						},
					},
					PrivateKeyProvider: &tls.PrivateKeyProvider{
						ProviderName: "qat",
						ConfigType: &tls.PrivateKeyProvider_TypedConfig{
							TypedConfig: msg,
						},
					},
				},
			}

		default:
			if sgxEnabled {
				s.PrivateKey = nil
				stage := Stage1
				if certificateReady {
					stage = Stage2
				}

				sdsSecretConfig := authn_model.ConstructSdsSecretConfig(authn_model.SDSDefaultResourceName)

				CSRConfig, err := NewCSRConfig(sanExtension)
				if err != nil {
					return nil
				}
				conf := &sgx.SgxPrivateKeyMethodConfig{
					SgxLibrary:  SgxLibrary,
					KeyLabel:    KeyLabel,
					UsrPin:      UsrPin,
					SoPin:       SoPin,
					TokenLabel:  TokenLabel,
					RsaKeySize:  RsaKeySize,
					Stage:       stage,
					KeyType:     KeyType,
					CsrConfig:   CSRConfig + UsrPin + "\n",
					SdsConfig:   sdsSecretConfig.SdsConfig,
					QuoteKey:    DefaultQuoteKey,
					QuotepubKey: DefaultPubKey,
				}

				if useECDSA {
					conf.KeyType = "ecdsa"
					conf.EcdsaKeyParam = "P-256"
				}

				msg, _ := anypb.New(conf)

				secret.Type = &tls.Secret_TlsCertificate{
					TlsCertificate: &tls.TlsCertificate{
						CertificateChain: &core.DataSource{
							Specifier: &core.DataSource_InlineBytes{
								InlineBytes: s.CertificateChain,
							},
						},
						PrivateKeyProvider: &tls.PrivateKeyProvider{
							ProviderName: "sgx",
							ConfigType: &tls.PrivateKeyProvider_TypedConfig{
								TypedConfig: msg,
							},
						},
						PrivateKey: nil,
					},
				}
			} else {
				secret.Type = &tls.Secret_TlsCertificate{
					TlsCertificate: &tls.TlsCertificate{
						CertificateChain: &core.DataSource{
							Specifier: &core.DataSource_InlineBytes{
								InlineBytes: s.CertificateChain,
							},
						},
						PrivateKey: &core.DataSource{
							Specifier: &core.DataSource_InlineBytes{
								InlineBytes: s.PrivateKey,
							},
						},
					},
				}
			}
		}
	}
	return secret
}

func NewCSRConfig(san string) (string, error) {
	return tmpl.Evaluate(CsrConfig, map[string]interface{}{
		"SAN": san,
	})
}

func pushLog(names []string) model.XdsLogDetails {
	if len(names) == 1 {
		// For common case of single resource, show which resource it was
		return model.XdsLogDetails{AdditionalInfo: "resource:" + names[0]}
	}
	return model.DefaultXdsLogDetails
}
