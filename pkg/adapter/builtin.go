// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"

	"github.com/goschtalt/goschtalt"
)

type float interface {
	~float32 | ~float64
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// only returns true if the list is empty or the string appears in the list
func only(s string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, v := range list {
		if s == v {
			return true
		}
	}

	return false
}

func numOrStringToString(v reflect.Value, accepts ...string) string {
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
	}

	if v.Kind() == reflect.String {
		return v.String()
	}

	return numToString(v, accepts...)
}

func numToString(v reflect.Value, accepts ...string) string {
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
	}

	if v.Kind() == reflect.Bool && only("bool", accepts) {
		return strconv.FormatBool(v.Bool())
	}
	if v.CanUint() && only("uint", accepts) {
		return strconv.FormatUint(v.Uint(), 10)
	}
	if v.CanInt() && only("int", accepts) {
		return strconv.FormatInt(v.Int(), 10)
	}
	if v.CanFloat() && only("float", accepts) {
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	// Ignore complex64 & complex128 since mapstruct and hashstruct don't
	// support them.

	return ""
}

func numToStringErr(v reflect.Value, only ...string) (string, error) {
	s := numToString(v, only...)
	if s == "" {
		return "", goschtalt.ErrNotApplicable
	}
	return s, nil
}

func sToB(s string, ptr int) (any, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
			err = goschtalt.ErrNotApplicable
		}
		return nil, err
	}

	switch ptr {
	case 0:
		return b, nil
	case 1:
		return toPtr(b), nil
	case 2:
		return toPtr(toPtr(b)), nil
	case 3:
		return toPtr(toPtr(toPtr(b))), nil
	default:
	}
	return nil, goschtalt.ErrUnsupported
}

func sToF[T float](s string, bits, ptr int) (any, error) {
	f, err := strconv.ParseFloat(s, bits)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
			err = goschtalt.ErrNotApplicable
		}
		return nil, err
	}

	rv := T(f)
	switch ptr {
	case 0:
		return rv, nil
	case 1:
		return toPtr(rv), nil
	case 2:
		return toPtr(toPtr(rv)), nil
	case 3:
		return toPtr(toPtr(toPtr(rv))), nil
	default:
	}
	return nil, goschtalt.ErrUnsupported
}

func sToI[T signed](s string, bits, ptr int) (any, error) {
	i, err := strconv.ParseInt(s, 0, bits)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
			err = goschtalt.ErrNotApplicable
		}
		return nil, err
	}

	rv := T(i)
	switch ptr {
	case 0:
		return rv, nil
	case 1:
		return toPtr(rv), nil
	case 2:
		return toPtr(toPtr(rv)), nil
	case 3:
		return toPtr(toPtr(toPtr(rv))), nil
	default:
	}
	return nil, goschtalt.ErrUnsupported
}

func sToU[T unsigned](s string, bits, ptr int) (any, error) {
	u, err := strconv.ParseUint(s, 0, bits)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrSyntax { //nolint:errorlint
			err = goschtalt.ErrNotApplicable
		}
		return nil, err
	}

	rv := T(u)
	switch ptr {
	case 0:
		return rv, nil
	case 1:
		return toPtr(rv), nil
	case 2:
		return toPtr(toPtr(rv)), nil
	case 3:
		return toPtr(toPtr(toPtr(rv))), nil
	default:
	}
	return nil, goschtalt.ErrUnsupported
}

func sToS(s string, ptr int) (any, error) {
	switch ptr {
	case 0:
		return s, nil
	case 1:
		return toPtr(s), nil
	case 2:
		return toPtr(toPtr(s)), nil
	case 3:
		return toPtr(toPtr(toPtr(s))), nil
	default:
	}
	return nil, goschtalt.ErrUnsupported
}

func stringToNum(s string, want reflect.Value) (any, error) {
	if s == "" {
		return nil, goschtalt.ErrNotApplicable
	}

	var ptr int
	for want.Kind() == reflect.Ptr {
		want = want.Elem()
		ptr++
	}

	switch want.Kind() {
	// Boolean
	case reflect.Bool:
		return sToB(s, ptr)

	// Float
	case reflect.Float32:
		return sToF[float32](s, want.Type().Bits(), ptr)
	case reflect.Float64:
		return sToF[float64](s, want.Type().Bits(), ptr)

	// Signed
	case reflect.Int:
		return sToI[int](s, want.Type().Bits(), ptr)
	case reflect.Int8:
		return sToI[int8](s, want.Type().Bits(), ptr)
	case reflect.Int16:
		return sToI[int16](s, want.Type().Bits(), ptr)
	case reflect.Int32:
		return sToI[int32](s, want.Type().Bits(), ptr)
	case reflect.Int64:
		return sToI[int64](s, want.Type().Bits(), ptr)

	// Unsigned
	case reflect.Uint:
		return sToU[uint](s, want.Type().Bits(), ptr)
	case reflect.Uint8:
		return sToU[uint8](s, want.Type().Bits(), ptr)
	case reflect.Uint16:
		return sToU[uint16](s, want.Type().Bits(), ptr)
	case reflect.Uint32:
		return sToU[uint32](s, want.Type().Bits(), ptr)
	case reflect.Uint64:
		return sToU[uint64](s, want.Type().Bits(), ptr)
	case reflect.Uintptr:
		return sToU[uintptr](s, want.Type().Bits(), ptr)

	// String
	case reflect.String:
		return sToS(s, ptr)

	default:
	}

	return nil, goschtalt.ErrNotApplicable
}

type marshalBuiltin struct {
	typ string
}

func (marshalBuiltin) From(from, to reflect.Value) (any, error) {
	return stringToNum(numOrStringToString(from), to)
}

func (m marshalBuiltin) To(from reflect.Value) (any, error) {
	return numToStringErr(from, m.typ)
}
