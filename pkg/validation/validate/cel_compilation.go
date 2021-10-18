/*
Copyright 2016 The Kubernetes Authors.

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

package validate

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	celmodel "k8s.io/kube-openapi/third_party/forked/celopenapi/model"
)

// CelRules defines the format of the x-kubernetes-validator schema extension.
type CelRules []CelRule

// CelRule defines the format of each rule in CelRules.
type CelRule struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func compileCel(schema *spec.Schema, CelRules CelRules) ([]cel.Program, []error) {
	allErrors := []error{}
	env, _ := cel.NewEnv()
	reg := celmodel.NewRegistry(env)
	rt, err := celmodel.NewRuleTypes("__root__", schema, reg)
	if err != nil {
		allErrors = append(allErrors, err)
		return nil, allErrors
	}
	opts, err := rt.EnvOptions(env.TypeProvider())
	if err != nil {
		allErrors = append(allErrors, err)
		return nil, allErrors
	}

	var propDecls []*expr.Decl
	if root, ok := rt.FindDeclType("__root__"); ok {
		if root.IsObject() {
			for k, f := range root.Fields {
				propDecls = append(propDecls, decls.NewVar(k, f.Type.ExprType()))
			}
		}
		// TODO: handle types other than object
	}
	opts = append(opts, cel.Declarations(propDecls...))
	env, err = env.Extend(opts...)
	if err != nil {
		allErrors = append(allErrors, err)
		return nil, allErrors
	}

	programs := make([]cel.Program, len(CelRules))
	for i, rule := range CelRules {
		ast, issues := env.Compile(rule.Rule)
		if issues != nil {
			// TODO: return detailed error message or rule.message?
			allErrors = append(allErrors, issues.Err())
		} else {
			prog, err := env.Program(ast)
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("program instantiation failed: %v", err))
			} else {
				programs[i] = prog
			}
		}
	}
	return programs, allErrors
}
