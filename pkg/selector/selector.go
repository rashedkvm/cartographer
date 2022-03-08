// Copyright 2021 VMware
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

package selector

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

type Selectable interface {
	GetLabels() map[string]string
}

type SelectingObject interface {
	GetSelector() v1alpha1.Selector
}

type SelectorMatchError interface {
	error
	GetSelectingObjectIndex() int
}

type selectorMatchError struct {
	Err                  error
	SelectingObjectIndex int
}

func (e selectorMatchError) Error() string {
	return e.Err.Error()
}

func (e selectorMatchError) GetSelectingObjectIndex() int {
	return e.SelectingObjectIndex
}

func Matches(req v1alpha1.FieldSelectorRequirement, context interface{}) (bool, error) {
	evaluator := eval.EvaluatorBuilder()
	actualValue, err := evaluator.EvaluateJsonPath(req.Key, context)
	if err != nil {
		return false, err
	}

	switch req.Operator {
	case v1alpha1.FieldSelectorOpIn:
		for _, v := range req.Values {
			if actualValue == v {
				return true, nil
			}
		}
		return false, nil
	case v1alpha1.FieldSelectorOpNotIn:
		for _, v := range req.Values {
			if actualValue == v {
				return false, nil
			}
		}
		return true, nil
	case v1alpha1.FieldSelectorOpExists:
		return actualValue != nil, nil
	case v1alpha1.FieldSelectorOpDoesNotExist:
		return actualValue == nil, nil
	default:
		return false, fmt.Errorf("invalid operator %s for field selector", req.Operator)
	}
}

type EnumerableSelectingObjects interface {
	EachSelectingObject(handler func(index int, selectingObject SelectingObject) SelectorMatchError) SelectorMatchError
}

func BestSelectorMatchIndices(selectable Selectable, selectingObjects EnumerableSelectingObjects) ([]int, SelectorMatchError) {

	//	if len(selectingObjects) == 0 { //FIXME: is this behavior still preserved?
	//		return nil, nil, nil
	//	}

	var matchingSelectorIndices = map[int][]int{}
	var highWaterMark = 0

	err := selectingObjects.EachSelectingObject(func(idx int, selectingObject SelectingObject) SelectorMatchError {
		selectors := selectingObject.GetSelector()

		matchScore := 0
		labelSelector := &metav1.LabelSelector{
			MatchLabels:      selectors.MatchLabels,
			MatchExpressions: selectors.MatchExpressions,
		}

		// -- Labels
		sel, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return selectorMatchError{
				Err:                  fmt.Errorf("selector matchLabels or matchExpressions are not valid: %w", err),
				SelectingObjectIndex: idx,
			}
		}
		if !sel.Matches(labels.Set(selectable.GetLabels())) {
			return nil // Bail early!
		}

		matchScore += len(labelSelector.MatchLabels)
		matchScore += len(labelSelector.MatchExpressions)

		// -- Fields
		allFieldsMatched, err := matchesAllFields(selectable, selectors.MatchFields)
		if err != nil {
			// Todo: test in unit test
			return selectorMatchError{
				Err:                  fmt.Errorf("failed to evaluate selector matchFields: %w", err),
				SelectingObjectIndex: idx,
			}
		}
		if !allFieldsMatched {
			return nil // Bail early!
		}
		matchScore += len(selectors.MatchFields)

		// -- decision time
		if matchScore > 0 {
			if matchingSelectorIndices[matchScore] == nil { //FIXME: needed?
				matchingSelectorIndices[matchScore] = []int{}
			}
			if matchScore > highWaterMark {
				highWaterMark = matchScore
			}
			matchingSelectorIndices[matchScore] = append(matchingSelectorIndices[matchScore], idx)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchingSelectorIndices[highWaterMark], nil
}

func matchesAllFields(source interface{}, requirements []v1alpha1.FieldSelectorRequirement) (bool, error) {
	for _, requirement := range requirements {
		match, err := Matches(requirement, source)
		//TODO: what happens if its JsonPathDoesNotExistError?
		if err != nil {
			return false, fmt.Errorf("unable to match field requirement with key [%s] operator [%s] values [%v]: %w", requirement.Key, requirement.Operator, requirement.Values, err)
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}
