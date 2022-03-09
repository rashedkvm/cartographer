package selector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/cartographer/pkg/selector"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

var _ = Describe("BestSelectorMatchIndices", func() {

	type testcase struct {
		selectable               selector.Selectable
		selectingObjects         selectingObjectList
		exptectedSelectorIndices []int
	}

	DescribeTable(
		"non error-cases",
		func(tc testcase) {
			actual, err := selector.BestSelectorMatchIndices(
				tc.selectable, tc.selectingObjects,
			)

			if tc.exptectedSelectorIndices == nil {
				Expect(actual).To(BeNil())
				return
			}

			Expect(err).NotTo(HaveOccurred())

			Expect(actual).To(Equal(tc.exptectedSelectorIndices))
		},

		Entry("no selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{}},
			exptectedSelectorIndices: nil,
		}),

		Entry("complete mismatched selectors & selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				},
			},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"my": "label",
					},
					nil,
					fields{
						v1alpha1.FieldSelectorRequirement{
							Key:      "Spec.libertyGibbet",
							Operator: "Exists",
							Values:   nil,
						},
					},
				),
			},
			exptectedSelectorIndices: nil,
		}),

		Entry("partial match; selectors with less labels than selectors", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"type": "web",
					},
					nil,
					nil,
				),
			},
			exptectedSelectorIndices: []int{0},
		}),

		Entry("partial match; selectors with less labels than target", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
				}},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					nil,
					nil,
				),
			},
			exptectedSelectorIndices: nil,
		}),

		Entry("absolute match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
				}},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"type": "web",
						"test": "webvalue",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type": "web",
						"test": "tekton",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type": "mobile",
						"test": "mobilevalue",
					},
					nil,
					nil,
				),
			},
			exptectedSelectorIndices: []int{1},
		}),

		Entry("exact partial match", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type":  "web",
					"test":  "tekton",
					"scan":  "security",
					"input": "image",
				}},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"input": "image",
					},
					nil,
					nil,
				),
			},
			exptectedSelectorIndices: []int{1, 2},
		}),

		Entry("exact match with no extras", testcase{
			selectable: selectable{
				labels: labels2.Set{
					"type": "web",
					"test": "tekton",
					"scan": "security",
				}},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					labels2.Set{
						"type": "atype",
						"test": "tekton",
						"scan": "ascan",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type": "web",
						"test": "tekton",
						"scan": "security",
					},
					nil,
					nil,
				),
				newSelectingObject(
					labels2.Set{
						"type":  "web",
						"test":  "tekton",
						"scan":  "security",
						"input": "image",
					},
					nil,
					nil,
				),
			},
			exptectedSelectorIndices: []int{1},
		}),

		Entry("match selectors with many fields in selectors", testcase{
			selectable: selectable{
				Spec: Spec{
					Color: "red",
					Age:   4,
				},
			},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					nil,
					nil,
					fields{
						{
							Key:      "Spec.Color",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				),
			},
			exptectedSelectorIndices: []int{0},
		}),

		Entry("match selectors when json path is not found, don't error", testcase{
			selectable: selectable{
				Spec: Spec{
					Color:  "red",
					Age:    4,
					Bucket: []Thing{{Name: "morko"}},
				},
			},
			selectingObjects: selectingObjectList{
				newSelectingObject(
					nil,
					nil,
					fields{
						{
							Key:      "spec.bucket[?(@.name==\"marco\")].name",
							Operator: "NotIn",
							Values:   []string{"green", "blue"},
						},
					},
				),
			},
			exptectedSelectorIndices: nil,
		}),
		//FIXME: should this case matter? it's broken
		//FEntry("match selectors when json path is not found, don't error", testcase{
		//	selectable: selectable{
		//		Spec: Spec{
		//			Color:  "red",
		//			Age:    4,
		//			Map: map[string]string{"name": "morko"},
		//		},
		//	},
		//	selectors: selectingObjectList{
		//		newSelectingObject(
		//			nil,
		//			nil,
		//			fields{
		//				{
		//					Key:      "spec.map.marco",
		//					Operator: "NotIn",
		//					Values:   []string{"green", "blue"},
		//				},
		//			},
		//		),
		//	},
		//	exptectedSelectorIndices: []int{},
		//}),
	)

	Describe("malformed selectors", func() {
		Context("label selector invalid", func() {
			var sel selectingObjectList
			BeforeEach(func() {
				sel = selectingObjectList{
					newSelectingObjectWithID(
						"valid-selector",
						"Special",
						labels2.Set{
							"fred": "derf",
						},
						nil,
						nil,
					),
					newSelectingObjectWithID(
						"invalid-selector",
						"Special",
						labels2.Set{
							"fred-": "derf-",
						},
						nil,
						nil,
					),
				}
			})

			It("returns an error", func() {
				_, err := selector.BestSelectorMatchIndices(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selector matchLabels or matchExpressions are not valid")))
				Expect(err).To(MatchError(ContainSubstring("key: Invalid value")))
				Expect(err.GetSelectingObjectIndex()).To(Equal(1))
			})
		})

		Context("expression selector invalid", func() {
			var sel selectingObjectList
			BeforeEach(func() {
				sel = selectingObjectList{
					newSelectingObjectWithID(
						"valid-selector",
						"Special",
						nil,
						[]metav1.LabelSelectorRequirement{
							{
								Key:      "valid",
								Operator: "Exists",
								Values:   nil,
							},
						},
						nil,
					),
					newSelectingObjectWithID(
						"invalid-selector",
						"Special",
						nil,
						[]metav1.LabelSelectorRequirement{
							{
								Key:      "fred",
								Operator: "Matchingest",
								Values:   nil,
							},
						},
						nil,
					),
				}
			})

			It("returns an error", func() {
				_, err := selector.BestSelectorMatchIndices(selectable{}, sel)
				Expect(err).To(MatchError(ContainSubstring("selector matchLabels or matchExpressions are not valid")))
				// TODO: 'pod' - Hmmmmm - perhaps we shouldn't be using v1 code?
				Expect(err).To(MatchError(ContainSubstring("\"Matchingest\" is not a valid pod selector operator")))
				Expect(err.GetSelectingObjectIndex()).To(Equal(1))
			})
		})
	})
})

type fields []v1alpha1.FieldSelectorRequirement

type Spec struct {
	Color  string            `json:"color"`
	Age    int               `json:"age"`
	Bucket []Thing           `json:"bucket"`
	Map    map[string]string `json:"map"`
}

type Thing struct {
	Name string `json:"name"`
}

type selectable struct {
	labels map[string]string
	Spec   `json:"spec"`
}

func (o selectable) GetLabels() map[string]string {
	return o.labels
}

type selectingObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	v1alpha1.Selector
}

func newSelectingObject(labels labels2.Set, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *selectingObject {
	return newSelectingObjectWithID("testSelectingObject", "Test", labels, expressions, fields)
}

func newSelectingObjectWithID(name, kind string, labels labels2.Set, expressions []metav1.LabelSelectorRequirement, fields []v1alpha1.FieldSelectorRequirement) *selectingObject {
	return &selectingObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: "testv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Selector: v1alpha1.Selector{
			LabelSelector: metav1.LabelSelector{
				MatchLabels:      labels,
				MatchExpressions: expressions,
			},
			MatchFields: fields,
		},
	}
}

func (b *selectingObject) GetSelector() v1alpha1.Selector {
	return b.Selector
}

type selectingObjectList []*selectingObject

func (l selectingObjectList) EachSelectingObject(handler func(idx int, selectingObject selector.SelectingObject) selector.SelectorMatchError) selector.SelectorMatchError {
	for idx, item := range l {
		if err := handler(idx, item); err != nil {
			return err
		}
	}
	return nil
}
