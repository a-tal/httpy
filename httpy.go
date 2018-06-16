// Package httpy provides an interface for embedding python http workers
package httpy

/*
#cgo pkg-config: python-3.6
#include "Python.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

var (
	reqPointer *C.PyObject
)

// Init must be called before Request can be used.
//
// Args:
//     initModule: optional python module to load the initFunction from
//     initFunction: optional python function to call for initialization
//     reqModule: required python module where the reqFunction is
//     reqFunction: required python function to call per HTTPyRequest
//
// Returns:
//     routes map of paths -> http methods returned from the initFunction
//     error interface if anything went wrong during init (golang pls...)
func Init(initModule, initFunction, reqModule, reqFunction string) (
	routes map[string][]string,
	err error,
) {
	if reqModule == "" || reqFunction == "" {
		return nil, errors.New("both reqModule and reqFunction are required")
	}

	if err = initPython(); err != nil {
		return nil, err // such an antipattern it's not even funny
	}

	var initMod *C.PyObject
	var reqMod *C.PyObject

	if initModule != "" && initFunction != "" {
		initMod, routes, err = doPyInit(initModule, initFunction)
		if err != nil {
			return nil, err
		}
		if reqModule == initModule {
			reqMod = initMod
		}
	}

	if reqMod == nil {
		reqMod = getPyModule(reqModule)
	}

	reqPointer = getPyAttr(reqMod, reqFunction)

	return routes, nil
}

// Request is a per HTTP request interface to the initialized python worker.
//
// Args:
//     method: string http method called
//     path: route which was matched (up to your implementation if patterned)
//     body: request body as string
//     params: path parameters as a map of string: strings
//     query: query parameters as a map of string: strings
//     headers: header values as a map of string: strings
//
// Returns:
//     status: integer http status from python
//     respBody: string response body from python
//     respHeaders: map of string: strings response headers from python
//     err: error interface if anything went wrong (antipattern :/)
func Request(
	method, path, body string,
	params, query, headers map[string][]string,
) (
	status int,
	respBody []byte,
	respHeaders map[string][]string,
	err error,
) {

	if reqPointer == nil {
		return 500, nil, nil, errors.New("Init must be called before Request")
	}

	// this is lazy and slow. room for improvement here for sure
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	state := C.PyGILState_Ensure()
	defer C.PyGILState_Release(state)

	args := pythonTuple([]*C.PyObject{
		pythonString(method),
		pythonString(path),
		mapToPyDict(params),
		mapToPyDict(query),
		mapToPyDict(headers),
		pythonString(body),
	})

	out, err := callPy(reqPointer, args, nil)

	if err != nil {
		status = 500
		err = fmt.Errorf("python exception: %+v", err)
	} else {
		status = int(C.PyLong_AsLong(C.PyTuple_GetItem(out, 0)))
		respBody = []byte(goString(C.PyTuple_GetItem(out, 1)))
		respHeaders = pyDictToMap(C.PyTuple_GetItem(out, 2))
	}

	C.Py_DecRef(out)

	return status, respBody, respHeaders, err
}

/*  END OF EXPORTED FUNCTIONS */

func initPython() error {
	// start python
	if C.Py_IsInitialized() == 0 {
		C.Py_Initialize()
	}
	if C.Py_IsInitialized() == 0 {
		return errors.New("failed to initialize python")
	}

	// make sure the GIL is correctly initialized
	if C.PyEval_ThreadsInitialized() == 0 {
		C.PyEval_InitThreads()
	}
	if C.PyEval_ThreadsInitialized() == 0 {
		return errors.New("failed to initialize python threads")
	}
	C.PyEval_ReleaseThread(C.PyGILState_GetThisThreadState())

	return nil
}

func doPyInit(m, f string) (
	mod *C.PyObject,
	routes map[string][]string,
	err error,
) {
	mod, pyInit, err := getPyFunc(m, f)
	if err != nil {
		return nil, nil, err
	}

	state := C.PyGILState_Ensure()
	defer C.PyGILState_Release(state)
	out, err := callPy(pyInit, nil, nil)

	if err != nil {
		C.PyErr_PrintEx(0)
		return nil, nil, fmt.Errorf("init failure in %s.%s", m, f)
	} else if out != nil {
		routes = pyDictToMap(out)
	}

	C.Py_DecRef(out)

	return mod, routes, nil
}

func getPyModule(name string) *C.PyObject {
	state := C.PyGILState_Ensure()
	defer C.PyGILState_Release(state)
	return C.PyImport_ImportModule(C.CString(name))
}

func getPyAttr(m *C.PyObject, f string) *C.PyObject {
	if m != nil {
		state := C.PyGILState_Ensure()
		defer C.PyGILState_Release(state)
		return C.PyObject_GetAttrString(m, C.CString(f))
	}
	return nil
}

func getPyFunc(m, f string) (module, function *C.PyObject, err error) {
	module = getPyModule(m)
	if module == nil {
		return nil, nil, fmt.Errorf("could not import: %s", m)
	}

	function = getPyAttr(module, f)
	if function == nil {
		return module, nil, fmt.Errorf("could not find %s in %s", f, m)
	}

	return module, function, nil
}

func callPy(f, a, kw *C.PyObject) (res *C.PyObject, err error) {
	if kw == nil {
		res = C.PyObject_CallObject(f, a)
	} else {
		res = C.PyObject_Call(f, a, kw)
		C.Py_DecRef(kw)
	}
	C.Py_DecRef(a)

	pyErr := C.PyErr_Occurred()
	if pyErr != nil {
		// would be nice to capture this and return as error string
		C.PyErr_Print()
		err = errors.New("python exception")
	}

	return res, err
}

// convert a map of string: strings to a python dictionary
func mapToPyDict(m map[string][]string) *C.PyObject {
	pyDict := map[*C.PyObject]*C.PyObject{}
	for key, values := range m {
		pyDict[pythonString(key)] = pythonListOfStrings(values)
	}
	return pythonDict(pyDict)
}

// convert a python dictionary of string: strings to golang
func pyDictToMap(o *C.PyObject) map[string][]string {
	goMap := map[string][]string{}
	keys := C.PyDict_Keys(o)
	numKeys := int(C.PyList_Size(keys))
	for i := 0; i < numKeys; i++ {
		key := C.PyList_GetItem(keys, C.long(i))
		values := C.PyDict_GetItem(o, key)
		if values != nil {
			numValues := int(C.PyList_Size(values))
			goKey := goString(key)
			goMap[goKey] = make([]string, numValues)
			for j := 0; j < numValues; j++ {
				goMap[goKey][j] = goString(C.PyList_GetItem(values, C.long(j)))
			}
		}
	}
	C.Py_DecRef(keys)
	C.Py_DecRef(o)
	return goMap
}

func pythonDict(kv map[*C.PyObject]*C.PyObject) *C.PyObject {
	d := C.PyDict_New()
	for key, value := range kv {
		if res := C.PyDict_SetItem(d, key, value); res != 0 {
			return d
		}
	}
	return d
}

func pythonTuple(a []*C.PyObject) *C.PyObject {
	t := C.PyTuple_New(C.long(len(a)))
	for i, obj := range a {
		if res := C.PyTuple_SetItem(t, C.long(i), obj); res != 0 {
			return t
		}
	}
	return t
}

func pythonListOfStrings(a []string) *C.PyObject {
	l := C.PyList_New(C.long(len(a)))
	for i, obj := range a {
		if res := C.PyList_SetItem(l, C.long(i), pythonString(obj)); res != 0 {
			return l
		}
	}
	return l
}

func pythonString(s string) *C.PyObject {
	cStr := C.CString(s)
	pyStr := C.PyUnicode_FromString(cStr)
	defer C.free(unsafe.Pointer(cStr))
	return pyStr
}

func goString(o *C.PyObject) string {
	return C.GoString(C.PyUnicode_AsUTF8(o))
}
