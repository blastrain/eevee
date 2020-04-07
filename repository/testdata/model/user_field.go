package model

import (
	entity "entity"
	"errors"
	"sort"
)

func (m *UserFieldsInstance) Each(iter func(*UserFieldInstance)) {
	if m == nil {
		return
	}
	for _, value := range m.values {
		iter(value)
	}
}

func (m *UserFieldsInstance) EachIndex(iter func(int, *UserFieldInstance)) {
	if m == nil {
		return
	}
	for idx, value := range m.values {
		iter(idx, value)
	}
}

func (m *UserFieldsInstance) EachWithError(iter func(*UserFieldInstance) error) error {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if err := iter(value); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (m *UserFieldsInstance) EachIndexWithError(iter func(int, *UserFieldInstance) error) error {
	if m == nil {
		return nil
	}
	for idx, value := range m.values {
		if err := iter(idx, value); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (m *UserFieldsInstance) Map(mapFunc func(*UserFieldInstance) *UserFieldInstance) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	mappedValues := []*UserFieldInstance{}
	for _, value := range m.values {
		mappedValue := mapFunc(value)
		if mappedValue != nil {
			mappedValues = append(mappedValues, mappedValue)
		}
	}
	return &UserFieldsInstance{values: mappedValues}
}

func (m *UserFieldsInstance) Any(cond func(*UserFieldInstance) bool) bool {
	if m == nil {
		return false
	}
	for _, value := range m.values {
		if cond(value) {
			return true
		}
	}
	return false
}

func (m *UserFieldsInstance) Some(cond func(*UserFieldInstance) bool) bool {
	return m.Any(cond)
}

func (m *UserFieldsInstance) IsIncluded(cond func(*UserFieldInstance) bool) bool {
	return m.Any(cond)
}

func (m *UserFieldsInstance) All(cond func(*UserFieldInstance) bool) bool {
	if m == nil {
		return false
	}
	for _, value := range m.values {
		if !cond(value) {
			return false
		}
	}
	return true
}

func (m *UserFieldsInstance) Sort(compare func(*UserFieldInstance, *UserFieldInstance) bool) {
	if m == nil {
		return
	}
	sort.Slice(m.values, func(i, j int) bool {
		return compare(m.values[i], m.values[j])
	})
}

func (m *UserFieldsInstance) Find(compare func(*UserFieldInstance, *UserFieldInstance) bool) {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if cond(value) {
			return value
		}
	}
	return nil
}

func (m *UserFieldsInstance) Filter(filter func(*UserFieldInstance) bool) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	filteredValues := []*UserFieldInstance{}
	for _, value := range m.values {
		if filter(value) {
			filteredValues = append(filteredValues, value)
		}
	}
	return &UserFieldsInstance{values: filteredValues}
}

func (m *UserFieldsInstance) IsEmpty() bool {
	if m == nil {
		return true
	}
	if len(m.values) == 0 {
		return true
	}
	return false
}

func (m *UserFieldsInstance) At(idx int) *UserFieldInstance {
	if m == nil {
		return nil
	}
	if idx < 0 {
		return nil
	}
	if len(m.values) > idx {
		return m.values[idx]
	}
	return nil
}

func (m *UserFieldsInstance) First() *UserFieldInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[0]
	}
	return nil
}

func (m *UserFieldsInstance) Last() *UserFieldInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[len(m.values)-1]
	}
	return nil
}

func (m *UserFieldsInstance) Compact() *UserFieldsInstance {
	if m == nil {
		return nil
	}
	compactedValues := []*UserFieldInstance{}
	for _, value := range m.values {
		if value == nil {
			continue
		}
		compactedValues = append(compactedValues, value)
	}
	return &UserFieldsInstance{values: compactedValues}
}

func (m *UserFieldsInstance) Add(args ...*UserFieldInstance) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	for _, value := range args {
		m.values = append(m.values, value)
	}
	return m
}

func (m *UserFieldsInstance) Merge(args ...*UserFieldsInstance) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	for _, arg := range args {
		for _, value := range arg.values {
			m.values = append(m.values, value)
		}
	}
	return m
}

func (m *UserFieldsInstance) Len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

func (m *UserFieldsInstanceCollection) Merge() *UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	values := []*UserFieldInstance{}
	for _, collection := range m {
		for _, value := range collection.values {
			values = append(values, value)
		}
	}
	return &UserFieldsInstance{values: values}
}

func (m *UserFieldsInstance) UniqueID() *UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserFieldInstance) bool {
		if _, exists := filterMap[value.ID]; exists {
			return false
		}
		filterMap[value.ID] = struct{}{}
		return true
	})
}

func (m *UserFieldsInstance) GroupByID() map[uint64]*UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UserFieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.ID]; !exists {
			values[value.ID] = &UserFieldsInstance{}
		}
		values[value.ID].Add(value)
	}
	return values
}

func (m *UserFieldsInstance) IDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.ID)
	}
	return values
}

func (m *UserFieldsInstance) UniqueUserID() *UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserFieldInstance) bool {
		if _, exists := filterMap[value.UserID]; exists {
			return false
		}
		filterMap[value.UserID] = struct{}{}
		return true
	})
}

func (m *UserFieldsInstance) GroupByUserID() map[uint64]*UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UserFieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.UserID]; !exists {
			values[value.UserID] = &UserFieldsInstance{}
		}
		values[value.UserID].Add(value)
	}
	return values
}

func (m *UserFieldsInstance) UserIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.UserID)
	}
	return values
}

func (m *UserFieldsInstance) UniqueFieldID() *UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserFieldInstance) bool {
		if _, exists := filterMap[value.FieldID]; exists {
			return false
		}
		filterMap[value.FieldID] = struct{}{}
		return true
	})
}

func (m *UserFieldsInstance) GroupByFieldID() map[uint64]*UserFieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UserFieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.FieldID]; !exists {
			values[value.FieldID] = &UserFieldsInstance{}
		}
		values[value.FieldID].Add(value)
	}
	return values
}

func (m *UserFieldsInstance) FieldIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.FieldID)
	}
	return values
}

func (m *UserFieldsInstance) FirstByID(a0 uint64) *UserFieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.ID != a0 {
			continue
		}
		return value
	}
	return nil
}

func (m *UserFieldsInstance) FilterByID(a0 uint64) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	values := []*UserFieldInstance{}
	for _, value := range m.values {
		if value.ID != a0 {
			continue
		}
		values = append(values, value)
	}
	return &UserFieldsInstance{values: values}
}

func (m *UserFieldsInstance) FirstByUserIDAndFieldID(a0 uint64, a1 uint64) *UserFieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.UserID != a0 {
			continue
		}
		if value.FieldID != a1 {
			continue
		}
		return value
	}
	return nil
}

func (m *UserFieldsInstance) FilterByUserIDAndFieldID(a0 uint64, a1 uint64) *UserFieldsInstance {
	if m == nil {
		return nil
	}
	values := []*UserFieldInstance{}
	for _, value := range m.values {
		if value.UserID != a0 {
			continue
		}
		if value.FieldID != a1 {
			continue
		}
		values = append(values, value)
	}
	return &UserFieldsInstance{values: values}
}

func (m *UserFieldImpl) createCollection(entities entity.UserFields) *UserFieldsInstance {
	values := &UserFieldsInstance{values: make([]*UserFieldInstance, len(entities))}
	for i := 0; i < len(entities); i += 1 {
		values.values[i] = m.create(entities[i], values)
	}
	return values
}

func (m *UserFieldImpl) create(entity *entity.UserField, values *UserFieldsInstance) *UserFieldInstance {
	var value *UserFieldInstance
	value = &UserFieldInstance{
		UserField:        entity,
		isAlreadyCreated: false,
	}
	return value
}
