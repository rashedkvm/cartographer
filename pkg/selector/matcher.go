package selector

import (
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

type EnumerableSelectingObjects interface {
	EachSelectingObject(handler func(index int, selectingObject SelectingObject) SelectorMatchError) SelectorMatchError
}

func BestSelectorMatchIndices(selectable Selectable, selectingObjects EnumerableSelectingObjects) ([]int, SelectorMatchError) {
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
		if matchScore > 0 && matchScore >= highWaterMark {
			highWaterMark = matchScore
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
		if err != nil {
			if _, ok := err.(eval.JsonPathDoesNotExistError); !ok {
				return false, fmt.Errorf("unable to match field requirement with key [%s] operator [%s] values [%v]: %w", requirement.Key, requirement.Operator, requirement.Values, err)
			}

		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

