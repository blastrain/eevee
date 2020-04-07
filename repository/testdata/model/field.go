package model

import (
	entity "entity"
	"errors"
	"sort"
)

func (m *FieldsInstance) Each(iter func(*FieldInstance)) {
	if m == nil {
		return
	}
	for _, value := range m.values {
		iter(value)
	}
}

func (m *FieldsInstance) EachIndex(iter func(int, *FieldInstance)) {
	if m == nil {
		return
	}
	for idx, value := range m.values {
		iter(idx, value)
	}
}

func (m *FieldsInstance) EachWithError(iter func(*FieldInstance) error) error {
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

func (m *FieldsInstance) EachIndexWithError(iter func(int, *FieldInstance) error) error {
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

func (m *FieldsInstance) Map(mapFunc func(*FieldInstance) *FieldInstance) *FieldsInstance {
	if m == nil {
		return nil
	}
	mappedValues := []*FieldInstance{}
	for _, value := range m.values {
		mappedValue := mapFunc(value)
		if mappedValue != nil {
			mappedValues = append(mappedValues, mappedValue)
		}
	}
	return &FieldsInstance{values: mappedValues}
}

func (m *FieldsInstance) Any(cond func(*FieldInstance) bool) bool {
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

func (m *FieldsInstance) Some(cond func(*FieldInstance) bool) bool {
	return m.Any(cond)
}

func (m *FieldsInstance) IsIncluded(cond func(*FieldInstance) bool) bool {
	return m.Any(cond)
}

func (m *FieldsInstance) All(cond func(*FieldInstance) bool) bool {
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

func (m *FieldsInstance) Sort(compare func(*FieldInstance, *FieldInstance) bool) {
	if m == nil {
		return
	}
	sort.Slice(m.values, func(i, j int) bool {
		return compare(m.values[i], m.values[j])
	})
}

func (m *FieldsInstance) Find(compare func(*FieldInstance, *FieldInstance) bool) {
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

func (m *FieldsInstance) Filter(filter func(*FieldInstance) bool) *FieldsInstance {
	if m == nil {
		return nil
	}
	filteredValues := []*FieldInstance{}
	for _, value := range m.values {
		if filter(value) {
			filteredValues = append(filteredValues, value)
		}
	}
	return &FieldsInstance{values: filteredValues}
}

func (m *FieldsInstance) IsEmpty() bool {
	if m == nil {
		return true
	}
	if len(m.values) == 0 {
		return true
	}
	return false
}

func (m *FieldsInstance) At(idx int) *FieldInstance {
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

func (m *FieldsInstance) First() *FieldInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[0]
	}
	return nil
}

func (m *FieldsInstance) Last() *FieldInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[len(m.values)-1]
	}
	return nil
}

func (m *FieldsInstance) Compact() *FieldsInstance {
	if m == nil {
		return nil
	}
	compactedValues := []*FieldInstance{}
	for _, value := range m.values {
		if value == nil {
			continue
		}
		compactedValues = append(compactedValues, value)
	}
	return &FieldsInstance{values: compactedValues}
}

func (m *FieldsInstance) Add(args ...*FieldInstance) *FieldsInstance {
	if m == nil {
		return nil
	}
	for _, value := range args {
		m.values = append(m.values, value)
	}
	return m
}

func (m *FieldsInstance) Merge(args ...*FieldsInstance) *FieldsInstance {
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

func (m *FieldsInstance) Len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

func (m *FieldsInstanceCollection) Merge() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	values := []*FieldInstance{}
	for _, collection := range m {
		for _, value := range collection.values {
			values = append(values, value)
		}
	}
	return &FieldsInstance{values: values}
}

func (m *FieldsInstance) UniqueID() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.ID]; exists {
			return false
		}
		filterMap[value.ID] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByID() map[uint64]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.ID]; !exists {
			values[value.ID] = &FieldsInstance{}
		}
		values[value.ID].Add(value)
	}
	return values
}

func (m *FieldsInstance) IDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.ID)
	}
	return values
}

func (m *FieldsInstance) UniqueName() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[string]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.Name]; exists {
			return false
		}
		filterMap[value.Name] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByName() map[string]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[string]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Name]; !exists {
			values[value.Name] = &FieldsInstance{}
		}
		values[value.Name].Add(value)
	}
	return values
}

func (m *FieldsInstance) Names() []string {
	if m == nil {
		return nil
	}
	values := []string{}
	for _, value := range m.values {
		values = append(values, value.Name)
	}
	return values
}

func (m *FieldsInstance) UniqueLocationX() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.LocationX]; exists {
			return false
		}
		filterMap[value.LocationX] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByLocationX() map[int]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.LocationX]; !exists {
			values[value.LocationX] = &FieldsInstance{}
		}
		values[value.LocationX].Add(value)
	}
	return values
}

func (m *FieldsInstance) LocationXes() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.LocationX)
	}
	return values
}

func (m *FieldsInstance) UniqueLocationY() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.LocationY]; exists {
			return false
		}
		filterMap[value.LocationY] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByLocationY() map[int]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.LocationY]; !exists {
			values[value.LocationY] = &FieldsInstance{}
		}
		values[value.LocationY].Add(value)
	}
	return values
}

func (m *FieldsInstance) LocationIes() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.LocationY)
	}
	return values
}

func (m *FieldsInstance) UniqueObjectNum() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.ObjectNum]; exists {
			return false
		}
		filterMap[value.ObjectNum] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByObjectNum() map[int]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.ObjectNum]; !exists {
			values[value.ObjectNum] = &FieldsInstance{}
		}
		values[value.ObjectNum].Add(value)
	}
	return values
}

func (m *FieldsInstance) ObjectNums() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.ObjectNum)
	}
	return values
}

func (m *FieldsInstance) UniqueLevel() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.Level]; exists {
			return false
		}
		filterMap[value.Level] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByLevel() map[int]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Level]; !exists {
			values[value.Level] = &FieldsInstance{}
		}
		values[value.Level].Add(value)
	}
	return values
}

func (m *FieldsInstance) Levels() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.Level)
	}
	return values
}

func (m *FieldsInstance) UniqueDifficulty() *FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *FieldInstance) bool {
		if _, exists := filterMap[value.Difficulty]; exists {
			return false
		}
		filterMap[value.Difficulty] = struct{}{}
		return true
	})
}

func (m *FieldsInstance) GroupByDifficulty() map[int]*FieldsInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*FieldsInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Difficulty]; !exists {
			values[value.Difficulty] = &FieldsInstance{}
		}
		values[value.Difficulty].Add(value)
	}
	return values
}

func (m *FieldsInstance) Difficulties() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.Difficulty)
	}
	return values
}

func (m *FieldsInstance) FirstByID(a0 uint64) *FieldInstance {
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

func (m *FieldsInstance) FilterByID(a0 uint64) *FieldsInstance {
	if m == nil {
		return nil
	}
	values := []*FieldInstance{}
	for _, value := range m.values {
		if value.ID != a0 {
			continue
		}
		values = append(values, value)
	}
	return &FieldsInstance{values: values}
}

func (m *FieldsInstance) FirstByName(a0 string) *FieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.Name != a0 {
			continue
		}
		return value
	}
	return nil
}

func (m *FieldsInstance) FilterByName(a0 string) *FieldsInstance {
	if m == nil {
		return nil
	}
	values := []*FieldInstance{}
	for _, value := range m.values {
		if value.Name != a0 {
			continue
		}
		values = append(values, value)
	}
	return &FieldsInstance{values: values}
}

func (m *FieldsInstance) FirstByLocationXAndLocationY(a0 int, a1 int) *FieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.LocationX != a0 {
			continue
		}
		if value.LocationY != a1 {
			continue
		}
		return value
	}
	return nil
}

func (m *FieldsInstance) FilterByLocationXAndLocationY(a0 int, a1 int) *FieldsInstance {
	if m == nil {
		return nil
	}
	values := []*FieldInstance{}
	for _, value := range m.values {
		if value.LocationX != a0 {
			continue
		}
		if value.LocationY != a1 {
			continue
		}
		values = append(values, value)
	}
	return &FieldsInstance{values: values}
}

func (m *FieldsInstance) FirstByObjectNum(a0 int) *FieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.ObjectNum != a0 {
			continue
		}
		return value
	}
	return nil
}

func (m *FieldsInstance) FilterByObjectNum(a0 int) *FieldsInstance {
	if m == nil {
		return nil
	}
	values := []*FieldInstance{}
	for _, value := range m.values {
		if value.ObjectNum != a0 {
			continue
		}
		values = append(values, value)
	}
	return &FieldsInstance{values: values}
}

func (m *FieldsInstance) FirstByDifficultyAndLevel(a0 int, a1 int) *FieldInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.Difficulty != a0 {
			continue
		}
		if value.Level != a1 {
			continue
		}
		return value
	}
	return nil
}

func (m *FieldsInstance) FilterByDifficultyAndLevel(a0 int, a1 int) *FieldsInstance {
	if m == nil {
		return nil
	}
	values := []*FieldInstance{}
	for _, value := range m.values {
		if value.Difficulty != a0 {
			continue
		}
		if value.Level != a1 {
			continue
		}
		values = append(values, value)
	}
	return &FieldsInstance{values: values}
}

func (m *FieldImpl) createCollection(entities entity.Fields) *FieldsInstance {
	values := &FieldsInstance{values: make([]*FieldInstance, len(entities))}
	for i := 0; i < len(entities); i += 1 {
		values.values[i] = m.create(entities[i], values)
	}
	return values
}

func (m *FieldImpl) create(entity *entity.Field, values *FieldsInstance) *FieldInstance {
	var value *FieldInstance
	value = &FieldInstance{
		Field:            entity,
		isAlreadyCreated: false,
	}
	return value
}
