// Code generated by eevee. DO NOT EDIT!

package model

import (
	"daoplugin/entity"
	"daoplugin/model"
)

func DefaultField() *model.Field {
	value := model.NewField(&entity.Field{
		Difficulty: 1,
		ID:         1,
		Level:      1,
		LocationX:  1,
		LocationY:  1,
		Name:       "",
		ObjectNum:  1,
	}, nil)
	return value
}

func DefaultsField() *model.Fields {
	values := model.NewFields(entity.Fields{})
	values.Add(DefaultField())
	return values
}