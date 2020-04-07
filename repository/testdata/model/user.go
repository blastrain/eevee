package model

import (
	entity "entity"
	"errors"
	"sort"
)

func (m *UsersInstance) Each(iter func(*UserInstance)) {
	if m == nil {
		return
	}
	for _, value := range m.values {
		iter(value)
	}
}

func (m *UsersInstance) EachIndex(iter func(int, *UserInstance)) {
	if m == nil {
		return
	}
	for idx, value := range m.values {
		iter(idx, value)
	}
}

func (m *UsersInstance) EachWithError(iter func(*UserInstance) error) error {
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

func (m *UsersInstance) EachIndexWithError(iter func(int, *UserInstance) error) error {
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

func (m *UsersInstance) Map(mapFunc func(*UserInstance) *UserInstance) *UsersInstance {
	if m == nil {
		return nil
	}
	mappedValues := []*UserInstance{}
	for _, value := range m.values {
		mappedValue := mapFunc(value)
		if mappedValue != nil {
			mappedValues = append(mappedValues, mappedValue)
		}
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     mappedValues,
	}
}

func (m *UsersInstance) Any(cond func(*UserInstance) bool) bool {
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

func (m *UsersInstance) Some(cond func(*UserInstance) bool) bool {
	return m.Any(cond)
}

func (m *UsersInstance) IsIncluded(cond func(*UserInstance) bool) bool {
	return m.Any(cond)
}

func (m *UsersInstance) All(cond func(*UserInstance) bool) bool {
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

func (m *UsersInstance) Sort(compare func(*UserInstance, *UserInstance) bool) {
	if m == nil {
		return
	}
	sort.Slice(m.values, func(i, j int) bool {
		return compare(m.values[i], m.values[j])
	})
}

func (m *UsersInstance) Find(compare func(*UserInstance, *UserInstance) bool) {
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

func (m *UsersInstance) Filter(filter func(*UserInstance) bool) *UsersInstance {
	if m == nil {
		return nil
	}
	filteredValues := []*UserInstance{}
	for _, value := range m.values {
		if filter(value) {
			filteredValues = append(filteredValues, value)
		}
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     filteredValues,
	}
}

func (m *UsersInstance) IsEmpty() bool {
	if m == nil {
		return true
	}
	if len(m.values) == 0 {
		return true
	}
	return false
}

func (m *UsersInstance) At(idx int) *UserInstance {
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

func (m *UsersInstance) First() *UserInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[0]
	}
	return nil
}

func (m *UsersInstance) Last() *UserInstance {
	if m == nil {
		return nil
	}
	if len(m.values) > 0 {
		return m.values[len(m.values)-1]
	}
	return nil
}

func (m *UsersInstance) Compact() *UsersInstance {
	if m == nil {
		return nil
	}
	compactedValues := []*UserInstance{}
	for _, value := range m.values {
		if value == nil {
			continue
		}
		compactedValues = append(compactedValues, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     compactedValues,
	}
}

func (m *UsersInstance) Add(args ...*UserInstance) *UsersInstance {
	if m == nil {
		return nil
	}
	for _, value := range args {
		m.values = append(m.values, value)
	}
	return m
}

func (m *UsersInstance) Merge(args ...*UsersInstance) *UsersInstance {
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

func (m *UsersInstance) Len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

func (m *UsersInstanceCollection) Merge() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	if len(m) == 1 {
		return m[0]
	}
	values := []*UserInstance{}
	for _, collection := range m {
		for _, value := range collection.values {
			values = append(values, value)
		}
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UsersInstance) UniqueID() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.ID]; exists {
			return false
		}
		filterMap[value.ID] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByID() map[uint64]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.ID]; !exists {
			values[value.ID] = &UsersInstance{}
		}
		values[value.ID].Add(value)
	}
	return values
}

func (m *UsersInstance) IDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.ID)
	}
	return values
}

func (m *UsersInstance) UniqueName() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[string]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.Name]; exists {
			return false
		}
		filterMap[value.Name] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByName() map[string]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[string]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Name]; !exists {
			values[value.Name] = &UsersInstance{}
		}
		values[value.Name].Add(value)
	}
	return values
}

func (m *UsersInstance) Names() []string {
	if m == nil {
		return nil
	}
	values := []string{}
	for _, value := range m.values {
		values = append(values, value.Name)
	}
	return values
}

func (m *UsersInstance) UniqueSex() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[string]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.Sex]; exists {
			return false
		}
		filterMap[value.Sex] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupBySex() map[string]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[string]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Sex]; !exists {
			values[value.Sex] = &UsersInstance{}
		}
		values[value.Sex].Add(value)
	}
	return values
}

func (m *UsersInstance) Sexes() []string {
	if m == nil {
		return nil
	}
	values := []string{}
	for _, value := range m.values {
		values = append(values, value.Sex)
	}
	return values
}

func (m *UsersInstance) UniqueAge() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.Age]; exists {
			return false
		}
		filterMap[value.Age] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByAge() map[int]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.Age]; !exists {
			values[value.Age] = &UsersInstance{}
		}
		values[value.Age].Add(value)
	}
	return values
}

func (m *UsersInstance) Ages() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.Age)
	}
	return values
}

func (m *UsersInstance) UniqueSkillID() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.SkillID]; exists {
			return false
		}
		filterMap[value.SkillID] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupBySkillID() map[uint64]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.SkillID]; !exists {
			values[value.SkillID] = &UsersInstance{}
		}
		values[value.SkillID].Add(value)
	}
	return values
}

func (m *UsersInstance) SkillIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.SkillID)
	}
	return values
}

func (m *UsersInstance) UniqueSkillRank() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[int]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.SkillRank]; exists {
			return false
		}
		filterMap[value.SkillRank] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupBySkillRank() map[int]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[int]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.SkillRank]; !exists {
			values[value.SkillRank] = &UsersInstance{}
		}
		values[value.SkillRank].Add(value)
	}
	return values
}

func (m *UsersInstance) SkillRanks() []int {
	if m == nil {
		return nil
	}
	values := []int{}
	for _, value := range m.values {
		values = append(values, value.SkillRank)
	}
	return values
}

func (m *UsersInstance) UniqueGroupID() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.GroupID]; exists {
			return false
		}
		filterMap[value.GroupID] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByGroupID() map[uint64]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.GroupID]; !exists {
			values[value.GroupID] = &UsersInstance{}
		}
		values[value.GroupID].Add(value)
	}
	return values
}

func (m *UsersInstance) GroupIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.GroupID)
	}
	return values
}

func (m *UsersInstance) UniqueWorldID() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.WorldID]; exists {
			return false
		}
		filterMap[value.WorldID] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByWorldID() map[uint64]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.WorldID]; !exists {
			values[value.WorldID] = &UsersInstance{}
		}
		values[value.WorldID].Add(value)
	}
	return values
}

func (m *UsersInstance) WorldIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.WorldID)
	}
	return values
}

func (m *UsersInstance) UniqueFieldID() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[uint64]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.FieldID]; exists {
			return false
		}
		filterMap[value.FieldID] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByFieldID() map[uint64]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[uint64]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.FieldID]; !exists {
			values[value.FieldID] = &UsersInstance{}
		}
		values[value.FieldID].Add(value)
	}
	return values
}

func (m *UsersInstance) FieldIDs() []uint64 {
	if m == nil {
		return nil
	}
	values := []uint64{}
	for _, value := range m.values {
		values = append(values, value.FieldID)
	}
	return values
}

func (m *UsersInstance) UniqueUserFields() *UsersInstance {
	if len(m) == 0 {
		return nil
	}
	filterMap := map[UserFieldInstance]struct{}{}
	return m.Filter(func(value *UserInstance) bool {
		if _, exists := filterMap[value.UserFields]; exists {
			return false
		}
		filterMap[value.UserFields] = struct{}{}
		return true
	})
}

func (m *UsersInstance) GroupByUserFields() map[UserFieldInstance]*UsersInstance {
	if len(m) == 0 {
		return nil
	}
	values := map[UserFieldInstance]*UsersInstance{}
	for _, value := range m.values {
		if _, exists := values[value.UserFields]; !exists {
			values[value.UserFields] = &UsersInstance{}
		}
		values[value.UserFields].Add(value)
	}
	return values
}

func (m *UsersInstance) UserFieldsCollection() (UserFieldsInstanceCollection, error) {
	if m == nil {
		return nil, nil
	}
	values := UserFieldsInstanceCollection{}
	for _, value := range m.values {
		userFields, err := value.UserFields()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if userFields == nil {
			continue
		}
		values = append(values, userFields)
	}
	return values, nil
}

func (m *UsersInstance) findUserFields(id uint64, finder UserField) (UserFieldInstance, error) {
	if m.userFields != nil {
		return m.userFields.FirstByUserID(), nil
	}
	userFields, err := finder.FindByUserIDs(m.ids)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if userFields == nil {
		return nil, errors.New("cannot find record")
	}
	m.userFields = userFields
	return m.userFields.FirstByUserID(), nil
}

func (m *UsersInstance) FirstByID(a0 uint64) *UserInstance {
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

func (m *UsersInstance) FilterByID(a0 uint64) *UsersInstance {
	if m == nil {
		return nil
	}
	values := []*UserInstance{}
	for _, value := range m.values {
		if value.ID != a0 {
			continue
		}
		values = append(values, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UsersInstance) FirstByName(a0 string) *UserInstance {
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

func (m *UsersInstance) FilterByName(a0 string) *UsersInstance {
	if m == nil {
		return nil
	}
	values := []*UserInstance{}
	for _, value := range m.values {
		if value.Name != a0 {
			continue
		}
		values = append(values, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UsersInstance) FirstBySkillIDAndSkillRank(a0 uint64, a1 int) *UserInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.SkillID != a0 {
			continue
		}
		if value.SkillRank != a1 {
			continue
		}
		return value
	}
	return nil
}

func (m *UsersInstance) FilterBySkillIDAndSkillRank(a0 uint64, a1 int) *UsersInstance {
	if m == nil {
		return nil
	}
	values := []*UserInstance{}
	for _, value := range m.values {
		if value.SkillID != a0 {
			continue
		}
		if value.SkillRank != a1 {
			continue
		}
		values = append(values, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UsersInstance) FirstByGroupID(a0 uint64) *UserInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.GroupID != a0 {
			continue
		}
		return value
	}
	return nil
}

func (m *UsersInstance) FilterByGroupID(a0 uint64) *UsersInstance {
	if m == nil {
		return nil
	}
	values := []*UserInstance{}
	for _, value := range m.values {
		if value.GroupID != a0 {
			continue
		}
		values = append(values, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UsersInstance) FirstByWorldIDAndFieldID(a0 uint64, a1 uint64) *UserInstance {
	if m == nil {
		return nil
	}
	for _, value := range m.values {
		if value.WorldID != a0 {
			continue
		}
		if value.FieldID != a1 {
			continue
		}
		return value
	}
	return nil
}

func (m *UsersInstance) FilterByWorldIDAndFieldID(a0 uint64, a1 uint64) *UsersInstance {
	if m == nil {
		return nil
	}
	values := []*UserInstance{}
	for _, value := range m.values {
		if value.WorldID != a0 {
			continue
		}
		if value.FieldID != a1 {
			continue
		}
		values = append(values, value)
	}
	return &UsersInstance{
		ids:        m.ids,
		userFields: m.userFields,
		values:     values,
	}
}

func (m *UserImpl) createCollection(entities entity.Users) *UsersInstance {
	values := &UsersInstance{
		ids:    entities.IDs(),
		values: make([]*UserInstance, len(entities)),
	}
	for i := 0; i < len(entities); i += 1 {
		values.values[i] = m.create(entities[i], values)
	}
	return values
}

func (m *UserImpl) create(entity *entity.User, values *UsersInstance) *UserInstance {
	var value *UserInstance
	value = &UserInstance{
		User: entity,
		UserFields: func() (UserFieldsInstanceCollection, err) {
			return value.findUserFields(value.ID, m.user)
		},
		isAlreadyCreated: false,
	}
	return value
}
