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

package repository

import (
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/selector"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type SelectingObject interface {
	GetSelectors() v1alpha1.Selectors
	GetObjectKind() schema.ObjectKind
	GetName() string
}

type Selectable interface {
	GetLabels() map[string]string
}

// BestSelectorMatch attempts at finding the selectors that best match their selectors
// against the selectors.
func BestSelectorMatch(selectable Selectable, blueprints []SelectingObject) ([]SelectingObject, error) {

	if len(blueprints) == 0 {
		return nil, nil
	}

	var matchingSelectors = map[int][]SelectingObject{}
	var highWaterMark = 0

	for _, target := range blueprints {
		selectors := target.GetSelectors()

		size := 0
		labelSelector := &metav1.LabelSelector{
			MatchLabels:      selectors.Selector,
			MatchExpressions: selectors.SelectorMatchExpressions,
		}

		// -- Labels
		sel, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, fmt.Errorf(
				"selectorMatchExpressions or selectors of [%s/%s] are not valid: %w",
				target.GetObjectKind().GroupVersionKind().Kind,
				target.GetName(),
				err,
			)
		}
		if !sel.Matches(labels.Set(selectable.GetLabels())) {
			continue // Bail early!
		}

		size += len(labelSelector.MatchLabels)
		size += len(labelSelector.MatchExpressions)

		// -- Fields
		allFieldsMatched, err := matchesAllFields(selectable, selectors.SelectorMatchFields)
		if err != nil {
			// Todo: test in unit test
			return nil, fmt.Errorf(
				"failed to evaluate all matched fields of [%s/%s]: %w",
				target.GetObjectKind().GroupVersionKind().Kind,
				target.GetName(),
				err,
			)
		}
		if !allFieldsMatched {
			continue // Bail early!
		}
		size += len(selectors.SelectorMatchFields)

		// -- decision time
		if size > 0 {
			if matchingSelectors[size] == nil {
				matchingSelectors[size] = []SelectingObject{}
			}
			if size > highWaterMark {
				highWaterMark = size
			}
			matchingSelectors[size] = append(matchingSelectors[size], target)
		}
	}

	return matchingSelectors[highWaterMark], nil
}



type SelectingObject2 interface {
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

type TemplateOptionList []v1alpha1.TemplateOption

func (l TemplateOptionList) EachSelectingObject(handler func(idx int, selectingObject SelectingObject2) SelectorMatchError) SelectorMatchError {
	for idx, item := range l {
		if err := handler(idx, item); err != nil {
			return err
		}
	}
	return nil
}

type EnumerableSelectingObjects interface {
	EachSelectingObject(handler func(index int, selectingObject SelectingObject2) SelectorMatchError) SelectorMatchError
}

func BestSelectorMatchIndices(selectable Selectable, selectingObjects EnumerableSelectingObjects) ([]int, SelectorMatchError) {

	//	if len(selectingObjects) == 0 { //FIXME: is this behavior still preserved?
	//		return nil, nil, nil
	//	}

	var matchingSelectorIndices = map[int][]int{}
	var highWaterMark = 0

	err := selectingObjects.EachSelectingObject(func(idx int, selectingObject SelectingObject2) SelectorMatchError {
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
		match, err := selector.Matches(requirement, source)
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
