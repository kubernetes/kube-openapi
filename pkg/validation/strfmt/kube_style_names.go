// Copyright 2024 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"regexp"
)

// KubeStyleNames is the formats registry for Kubernetes style names.
var KubeStyleNames = NewSeededFormats(nil, nil)

func init() {
	// register formats in the KubeStyleNames registry:
	//   - k8s-dns1123subdomain
	//   - k8s-dns1123label
	//   - k8s-dns1035label
	dns1123subdomain := DNS1123Subdomain("")
	KubeStyleNames.Add("k8s-dns1123subdomain", &dns1123subdomain, IsDNS1123Subdomain)

	dns1123label := DNS1123Label("")
	KubeStyleNames.Add("k8s-dns1123label", &dns1123label, IsDNS1123Label)

	dns1035label := DNS1035Label("")
	KubeStyleNames.Add("k8s-dns1035label", &dns1035label, IsDNS1035Label)
}

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

// DNS1123LabelMaxLength is a label's max length in DNS (RFC 1123)
const DNS1123LabelMaxLength int = 63

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// IsDNS1123Label tests for a string that almost conforms to the definition of a label in
// DNS (RFC 1123), except that uppercase letters are not allowed.
// xref: https://github.com/kubernetes/kubernetes/issues/71140
func IsDNS1123Label(value string) bool {
	return len(value) <= DNS1123LabelMaxLength &&
		dns1123LabelRegexp.MatchString(value)
}

const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"

// DNS1123SubdomainMaxLength is a subdomain's max length in DNS (RFC 1123)
const DNS1123SubdomainMaxLength int = 253

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

// IsDNS1123Subdomain tests for a string that almost conforms to the definition of a
// subdomain in DNS (RFC 1123), except that uppercase letters are not allowed.
// and there is no max length of limit of 63 for each of the dot separated DNS Labels
// that make up the subdomain.
// xref: https://github.com/kubernetes/kubernetes/issues/71140
// xref: https://github.com/kubernetes/kubernetes/issues/79351
func IsDNS1123Subdomain(value string) bool {
	return len(value) <= DNS1123SubdomainMaxLength &&
		dns1123SubdomainRegexp.MatchString(value)
}

const dns1035LabelFmt string = "[a-z]([-a-z0-9]*[a-z0-9])?"

// DNS1035LabelMaxLength is a label's max length in DNS (RFC 1035)
const DNS1035LabelMaxLength int = 63

var dns1035LabelRegexp = regexp.MustCompile("^" + dns1035LabelFmt + "$")

// IsDNS1035Label tests for a string that almost conforms to the definition of a label in
// DNS (RFC 1035), except that uppercase letters are not allowed.
func IsDNS1035Label(value string) bool {
	return len(value) <= DNS1035LabelMaxLength &&
		dns1035LabelRegexp.MatchString(value)
}
