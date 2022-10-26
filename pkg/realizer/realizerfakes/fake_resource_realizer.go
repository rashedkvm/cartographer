// Code generated by counterfeiter. DO NOT EDIT.
package realizerfakes

import (
	"context"
	"sync"

	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type FakeResourceRealizer struct {
	DoStub        func(context.Context, realizer.OwnerResource, string, realizer.Outputs, meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, error)
	doMutex       sync.RWMutex
	doArgsForCall []struct {
		arg1 context.Context
		arg2 realizer.OwnerResource
		arg3 string
		arg4 realizer.Outputs
		arg5 meta.RESTMapper
	}
	doReturns struct {
		result1 templates.Reader
		result2 *unstructured.Unstructured
		result3 *templates.Output
		result4 bool
		result5 error
	}
	doReturnsOnCall map[int]struct {
		result1 templates.Reader
		result2 *unstructured.Unstructured
		result3 *templates.Output
		result4 bool
		result5 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeResourceRealizer) Do(arg1 context.Context, arg2 realizer.OwnerResource, arg3 string, arg4 realizer.Outputs, arg5 meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, error) {
	fake.doMutex.Lock()
	ret, specificReturn := fake.doReturnsOnCall[len(fake.doArgsForCall)]
	fake.doArgsForCall = append(fake.doArgsForCall, struct {
		arg1 context.Context
		arg2 realizer.OwnerResource
		arg3 string
		arg4 realizer.Outputs
		arg5 meta.RESTMapper
	}{arg1, arg2, arg3, arg4, arg5})
	stub := fake.DoStub
	fakeReturns := fake.doReturns
	fake.recordInvocation("Do", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.doMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3, ret.result4, ret.result5
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3, fakeReturns.result4, fakeReturns.result5
}

func (fake *FakeResourceRealizer) DoCallCount() int {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	return len(fake.doArgsForCall)
}

func (fake *FakeResourceRealizer) DoCalls(stub func(context.Context, realizer.OwnerResource, string, realizer.Outputs, meta.RESTMapper) (templates.Reader, *unstructured.Unstructured, *templates.Output, bool, error)) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = stub
}

func (fake *FakeResourceRealizer) DoArgsForCall(i int) (context.Context, realizer.OwnerResource, string, realizer.Outputs, meta.RESTMapper) {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	argsForCall := fake.doArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeResourceRealizer) DoReturns(result1 templates.Reader, result2 *unstructured.Unstructured, result3 *templates.Output, result4 bool, result5 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	fake.doReturns = struct {
		result1 templates.Reader
		result2 *unstructured.Unstructured
		result3 *templates.Output
		result4 bool
		result5 error
	}{result1, result2, result3, result4, result5}
}

func (fake *FakeResourceRealizer) DoReturnsOnCall(i int, result1 templates.Reader, result2 *unstructured.Unstructured, result3 *templates.Output, result4 bool, result5 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	if fake.doReturnsOnCall == nil {
		fake.doReturnsOnCall = make(map[int]struct {
			result1 templates.Reader
			result2 *unstructured.Unstructured
			result3 *templates.Output
			result4 bool
			result5 error
		})
	}
	fake.doReturnsOnCall[i] = struct {
		result1 templates.Reader
		result2 *unstructured.Unstructured
		result3 *templates.Output
		result4 bool
		result5 error
	}{result1, result2, result3, result4, result5}
}

func (fake *FakeResourceRealizer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeResourceRealizer) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ realizer.ResourceRealizer = new(FakeResourceRealizer)
