// Package processors provides the various functions to run processors.
package processors

import (
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

type StringArray struct {
	tengo.ObjectImpl
	Value []string
}

func (o *StringArray) String() string {
	return strings.Join(o.Value, ", ")
}

func (o *StringArray) BinaryOp(op token.Token, rhs tengo.Object) (tengo.Object, error) {
	if rhs, ok := rhs.(*StringArray); ok {
		switch op {
		case token.Add:
			if len(rhs.Value) == 0 {
				return o, nil
			}
			return &StringArray{Value: append(o.Value, rhs.Value...)}, nil
		}
	}

	return nil, tengo.ErrInvalidOperator
}

func (o *StringArray) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *StringArray) Equals(x tengo.Object) bool {
	if x, ok := x.(*StringArray); ok {
		if len(o.Value) != len(x.Value) {
			return false
		}

		for i, v := range o.Value {
			if v != x.Value[i] {
				return false
			}
		}

		return true
	}

	return false
}

func (o *StringArray) Copy() tengo.Object {
	return &StringArray{
		Value: append([]string{}, o.Value...),
	}
}

func (o *StringArray) TypeName() string {
	return "string-array"
}

func (o *StringArray) IndexGet(index tengo.Object) (tengo.Object, error) {
	intIdx, ok := index.(*tengo.Int)
	if ok {
		if intIdx.Value >= 0 && intIdx.Value < int64(len(o.Value)) {
			return &tengo.String{Value: o.Value[intIdx.Value]}, nil
		}

		return nil, tengo.ErrIndexOutOfBounds
	}

	strIdx, ok := index.(*tengo.String)
	if ok {
		for vidx, str := range o.Value {
			if strIdx.Value == str {
				return &tengo.Int{Value: int64(vidx)}, nil
			}
		}

		return tengo.UndefinedValue, nil
	}

	return nil, tengo.ErrInvalidIndexType
}

func (o *StringArray) IndexSet(index, value tengo.Object) error {
	strVal, ok := tengo.ToString(value)
	if !ok {
		return tengo.ErrInvalidIndexValueType
	}

	intIdx, ok := index.(*tengo.Int)
	if ok {
		if intIdx.Value >= 0 && intIdx.Value < int64(len(o.Value)) {
			o.Value[intIdx.Value] = strVal
			return nil
		}

		return tengo.ErrIndexOutOfBounds
	}

	return tengo.ErrInvalidIndexType
}

func (o *StringArray) CanCall() bool {
	return true
}

func (o *StringArray) Call(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	for i, v := range o.Value {
		if v == s1 {
			return &tengo.Int{Value: int64(i)}, nil
		}
	}

	return tengo.UndefinedValue, nil
}

func (o *StringArray) CanIterate() bool {
	return true
}

func (o *StringArray) Iterate() tengo.Iterator {
	return &StringArrayIterator{
		strArr: o,
	}
}

type StringArrayIterator struct {
	tengo.ObjectImpl
	strArr *StringArray
	idx    int
}

func (i *StringArrayIterator) TypeName() string {
	return "string-array-iterator"
}

func (i *StringArrayIterator) Next() bool {
	i.idx++
	return i.idx <= len(i.strArr.Value)
}

func (i *StringArrayIterator) Key() tengo.Object {
	return &tengo.Int{Value: int64(i.idx - 1)}
}

func (i *StringArrayIterator) Value() tengo.Object {
	return &tengo.String{Value: i.strArr.Value[i.idx-1]}

}
