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

package apply

import "istio.io/istio/pkg/test/framework/resource/config/cleanup"

// Options provide options for applying configuration
type Options struct {
	// Cleanup strategy
	Cleanup cleanup.Strategy

	// Wait for configuration to be propagated.
	Wait bool
}
