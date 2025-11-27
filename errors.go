/*
 * Copyright (c) 2023 ivfzhou
 * modify-variables-temporarily is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package modify_variables_temporarily

import (
	"errors"
	"fmt"
)

var (
	ErrTargetCannotBeNil             = errors.New("[MVT]: target cannot be nil")
	ErrTargetCannotBeNilType         = errors.New("[MVT]: target cannot be a variable with type but no value")
	ErrTargetIsNotPointer            = errors.New("[MVT]: target is not a pointer type")
	ErrTargetIsNotFunc               = errors.New("[MVT]: target is not a function type")
	ErrTargetIsNotPointerOrInterface = errors.New("[MVT]: target is neither a pointer type nor an interface type")
	ErrTargetIsNotSlice              = errors.New("[MVT]: target is not a slice type")
	ErrTargetIsNotSliceOrArray       = errors.New("[MVT]: target is neither a slice type nor an array type")
	ErrTargetIsNotMap                = errors.New("[MVT]: target is not a map type")
	ErrTargetIsNotStruct             = errors.New("[MVT]: target is not a struct type")
	ErrIncompatibleTypeAssignment    = errors.New("[MVT]: incompatible type assignment")
	ErrStructFieldNameCannotBeEmpty  = errors.New("[MVT]: field name can not be empty")
	ErrStructFieldNotFound           = errors.New("[MVT]: struct field not found")
	ErrTargetCannotBeSet             = errors.New("[MVT]: target can not be set")
	ErrCannotToNext                  = errors.New("[MVT]: cannot to next")
	ErrInvalidMapKeyType             = errors.New("[MVT]: invalid map key type")
	ErrNoActions                     = errors.New("[MVT]: no actions")
	ErrIndexOutOfBound               = errors.New("[MVT]: index out of bound")
)

func newStructFieldNotFoundError(structName string, index int) error {
	return fmt.Errorf("%w. struct %s does not have a field with index %d",
		ErrStructFieldNotFound, structName, index)
}

func newStructFieldNotFoundByNameError(structName string, name string) error {
	return fmt.Errorf("%w. struct %s does not have a field with named %s", ErrStructFieldNotFound, structName, name)
}

func newIncompatibleTypeAssignmentError(typeName, toTypeName string) error {
	return fmt.Errorf("%w. %s cannot be assigned to %s", ErrIncompatibleTypeAssignment, typeName, toTypeName)
}

func newInvalidMapKeyError(keyTypeName, toKeyTypeName string) error {
	return fmt.Errorf("%w. key %s cannot use in %s", ErrInvalidMapKeyType, keyTypeName, toKeyTypeName)
}

func newIndexOutOfBoundError(index int, typeName string, length int) error {
	return fmt.Errorf("%w. index %d out of %s length %d", ErrIndexOutOfBound, index, typeName, length)
}

func newTypeInvalid(err error, typ string) error {
	return fmt.Errorf("%w, type chain is %s", err, typ)
}
