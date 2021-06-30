/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/gengo/examples/set-gen/sets"
	"k8s.io/gengo/types"
	"k8s.io/klog/v2"
)

const (
	// TagLifecycle is the comment tag prefix for specifying
	// the API lifecycle info. Example tag format:
	// +lifecycle:component=<component-name>,minVersion=<version>,status=<prerelease-status>,featureGate=<featuregate-name>
	TagLifecycle = "lifecycle:component"

	KeyComponent   = "component"
	KeyMinVersion  = "minVersion"
	KeyStatus      = "status"
	KeyFeatureGate = "featureGate"

	kubernetesComponentName = "kubernetes"

	// as defined in https://github.com/kubernetes/kubernetes/blob/master/pkg/features/kube_features.go
	kubernetesFeatureGateType = "featuregate.Feature"
)

var (
	AllowedTagKeys = sets.NewString(KeyComponent, KeyMinVersion, KeyStatus, KeyFeatureGate)

	// the status names must match the prereleases defined in
	// k8s.io/component-base/featuregate
	AllowedStatusNames = sets.NewString("alpha", "beta", "deprecated")

	// regex for a valid k8s version
	minVersionRegex = regexp.MustCompile(`^v[1-9][0-9]*\.(0|[1-9][0-9]*)$`)
)

// APILifecyleTag implements APIRule interface.
type APILifecyleTag struct {
	// FeatureGateFileNames is the list of files defining feature gates.
	FeatureGateMap map[string]FeatureGateInfo
}

// FeatureGateInfo stores info about a particular feature gate.
type FeatureGateInfo struct {
	// status is the highest status for the feature gate.
	status string
	// minVersion is the minimum k8s version for the status.
	minVersion string
}

// Name returns the name of APIRule
func (l *APILifecyleTag) Name() string {
	return "incorrect_lifecycle_tag"
}

// Validate evaluates API rule on type t and returns a list of field names in
// the type that violate the rule. Empty field name [""] implies the entire
// type violates the rule.
func (l *APILifecyleTag) Validate(t *types.Type) ([]string, error) {
	fields := make([]string, 0)

	switch t.Kind {
	case types.Struct:
		for _, m := range t.Members {
			commentTag := types.ExtractCommentTags("+", m.CommentLines)[TagLifecycle]
			// Existing and new GA fields are listed as exceptions and added to violations.
			// This ensures that the types.go files are not spammed with comment tags and
			// absence of a comment tag is treated as an error instead of assuming the fields to be GA.
			if len(commentTag) == 0 {
				klog.V(3).Infof("[%s] %s: no comment tag found for \"%s\" tag: must be a GA field, adding to violations", t.Name, m.Name, TagLifecycle)
				fields = append(fields, m.Name)
				continue
			}

			// only consider the first occurance of tagLifecycleComponent
			if len(commentTag) > 1 {
				klog.V(3).Infof("[%s] %s: multiple comments found for \"%s\" tag, considering only the first occurance of %v", t.Name, m.Name, TagLifecycle, commentTag)
			}

			// commentValues is of the form ["kubernetes", "minVersion=1.22", "status=alpha", "featureGate=Foo"]
			commentValues := strings.Split(commentTag[0], ",")

			// the lifecycle tag must have complete info
			if len(commentValues) != 4 {
				klog.Warningf("[%s] %s: \"%s\" tag must have all keys in %v, found: %v", t.Name, m.Name, TagLifecycle, AllowedTagKeys.List(), commentTag[0])
				fields = append(fields, m.Name)
				continue
			}

			// component name must be kubernetesComponentName
			if commentValues[0] != kubernetesComponentName {
				klog.Warningf("[%s] %s: \"%s\" tag must have component name as \"%s\", found: %s", t.Name, m.Name, TagLifecycle, kubernetesComponentName, commentValues[0])
				fields = append(fields, m.Name)
				continue
			}

			var minVersion, status, featureGate string
			for _, commentValue := range commentValues[1:] {
				// must be of the form foo=bar
				commentValueParts := strings.Split(commentValue, "=")
				if len(commentValueParts) != 2 {
					klog.Warningf("[%s] %s: \"%s\" tag must have keys in the format \"foo=bar\", found invalid key: %v", t.Name, m.Name, TagLifecycle, commentValue)
					fields = append(fields, m.Name)
					break
				}
				key := commentValueParts[0]
				value := commentValueParts[1]

				// must be a valid tag key
				if !AllowedTagKeys.Has(key) {
					klog.Warningf("[%s] %s: \"%s\" tag has invalid key \"%s\", must be one of %v", t.Name, m.Name, TagLifecycle, key, AllowedTagKeys.List())
					fields = append(fields, m.Name)
					break
				}

				switch {
				case key == KeyMinVersion:
					if !minVersionRegex.MatchString(value) {
						klog.Warningf("[%s] %s: \"%s\" tag has invalid %s value: \"%s\", must satisfy the regular expression %s", t.Name, m.Name, TagLifecycle, KeyMinVersion, value, minVersionRegex)
						fields = append(fields, m.Name)
						break
					}
					minVersion = value
				case key == KeyStatus:
					if !AllowedStatusNames.Has(value) {
						klog.Warningf("[%s] %s: \"%s\" tag has invalid %s value: \"%s\", must be one of %v", t.Name, m.Name, TagLifecycle, KeyStatus, value, AllowedStatusNames.List())
						fields = append(fields, m.Name)
						break
					}
					status = value
				case key == KeyFeatureGate:
					featureGate = value
				}
			}

			if len(l.FeatureGateMap) != 0 && len(featureGate) != 0 && len(status) != 0 && len(minVersion) != 0 {
				if err := l.validateFeatureGates(featureGate, status, minVersion); err != nil {
					klog.Warningf(fmt.Sprintf("[%s] %s: \"%s\" tag: ", t.Name, m.Name, TagLifecycle) + err.Error())
					fields = append(fields, m.Name)
				}
			}
		}
	}
	return fields, nil
}

// validateFeatureGates returns true if the following conditions are true:
// 1. The value of the featureGate key is a valid feature gate as defined in the feature gate file.
// 2. The value of the status key is equal to the highest status of the feature gate as defined in the feature gate file.
// 3. The value of the minVersion key is equal to the minimum k8s version for the status defined in point 2.
func (l *APILifecyleTag) validateFeatureGates(featureGateValue, statusValue, minVersionValue string) error {
	if len(l.FeatureGateMap) != 0 {
		info, exists := l.FeatureGateMap[featureGateValue]
		if !exists {
			return fmt.Errorf("invalid %s value: \"%s\" is not a valid feature gate", KeyFeatureGate, featureGateValue)
		}
		if statusValue != info.status {
			return fmt.Errorf("invalid %s value: \"%s\" is not the highest status of the feature gate \"%s\", must be \"%s\"", KeyStatus, statusValue, featureGateValue, info.status)
		}
		if minVersionValue != info.minVersion {
			return fmt.Errorf("invalid %s value: \"%s\" is not the highest status of the feature gate \"%s\", must be \"%s\"", KeyMinVersion, minVersionValue, featureGateValue, info.minVersion)
		}
	}
	return nil
}

// ParseFeatureGateFiles parses the feature gate files and returns a map of
// featureGate names to corresponding FeatureGateInfo.
func ParseFeatureGateFiles(featureGateFileNames []string) (map[string]FeatureGateInfo, error) {
	out := map[string]FeatureGateInfo{}
	fset := token.NewFileSet()
	var fileAst *ast.File
	var err error

	for _, fileName := range featureGateFileNames {
		fileAst, err = parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			return out, err
		}

		for _, dd := range fileAst.Decls {
			gd, ok := dd.(*ast.GenDecl)
			if !ok {
				continue
			}
			// find constant declrations
			if gd.Tok != token.CONST {
				continue
			}
			for _, sp := range gd.Specs {
				// filter by kubernetesFeatureGateType
				valSp, ok := sp.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if gotypes.ExprString(valSp.Type) != kubernetesFeatureGateType {
					continue
				}

				// derive the featureGate name
				bslit, ok := valSp.Values[0].(*ast.BasicLit)
				if !ok {
					continue
				}
				featureGateName, err := strconv.Unquote(bslit.Value)
				if err != nil {
					return out, err
				}

				// assumes that comment lines are in the order: alpha, beta, deprecated, stable
				// also assumes that the format is of the type "alpha: v1.22"
				var status, minVersion string
				comments := strings.Split(valSp.Doc.Text(), "\n")
				for _, line := range comments {
					if lineParts := strings.Split(line, ":"); len(lineParts) != 2 {
						continue
					}

					switch {
					case strings.HasPrefix(line, "alpha"):
						status = "alpha"
						minVersion = strings.TrimSpace(strings.Split(line, ":")[1])
					case strings.HasPrefix(line, "beta"):
						status = "beta"
						minVersion = strings.TrimSpace(strings.Split(line, ":")[1])
					case strings.HasPrefix(line, "deprecated"):
						status = "deprecated"
						minVersion = strings.TrimSpace(strings.Split(line, ":")[1])
					case strings.HasPrefix(line, "stable"):
						status = ""
						minVersion = ""
					}
				}

				// don't store stable featuregates
				if len(status) != 0 && len(minVersion) != 0 {
					out[featureGateName] = FeatureGateInfo{status: status, minVersion: minVersion}
				}
			}
		}
	}
	return out, nil
}
