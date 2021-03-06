// Code generated by eevee. DO NOT EDIT!

package factory

import (
	"rapidashplugin/entity"
	"rapidashplugin/model"
	"time"
)

func DefaultUserField() *model.UserField {
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
	updatedAt, _ := time.Parse("2006-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
	value := &model.UserField{UserField: &entity.UserField{
		CreatedAt: createdAt,
		FieldID:   0,
		ID:        0,
		UpdatedAt: updatedAt,
		UserID:    0,
	}}
	return value
}

func DefaultUserFields() *model.UserFields {
	values := &model.UserFields{}
	{
		createdAt, _ := time.Parse("2006-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
		updatedAt, _ := time.Parse("2006-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
		value := &model.UserField{UserField: &entity.UserField{
			CreatedAt: createdAt,
			FieldID:   0,
			ID:        0,
			UpdatedAt: updatedAt,
			UserID:    0,
		}}
		values.Add(value)
	}
	return values
}
