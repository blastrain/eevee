// Code generated by eevee. DO NOT EDIT!

package repository

import (
	"/entity"
	"/model"
	"context"
	"log"
	"reflect"

	"golang.org/x/xerrors"
)

type WorldMock struct {
	expect *WorldExpect
}

func (r *WorldMock) EXPECT() *WorldExpect {
	return r.expect
}

func NewWorldMock() *WorldMock {
	return &WorldMock{expect: NewWorldExpect()}
}

type WorldToModelExpect struct {
	expect        *WorldExpect
	isOutOfOrder  bool
	isAnyTimes    bool
	requiredTimes int
	calledTimes   int
	actions       []func(value *entity.World)
	value         *entity.World
	r0            *model.World
}

func (r *WorldToModelExpect) Return(r0 *model.World) *WorldToModelExpect {
	r.r0 = r0
	return r
}

func (r *WorldToModelExpect) Do(action func(value *entity.World)) *WorldToModelExpect {
	r.actions = append(r.actions, action)
	return r
}

func (r *WorldToModelExpect) OutOfOrder() *WorldToModelExpect {
	r.isOutOfOrder = true
	return r
}

func (r *WorldToModelExpect) AnyTimes() *WorldToModelExpect {
	r.isAnyTimes = true
	return r
}

func (r *WorldToModelExpect) Times(n int) *WorldToModelExpect {
	r.requiredTimes = n
	return r
}

func (r *WorldMock) ToModel(value *entity.World) (r0 *model.World) {
	if len(r.expect.toModel) == 0 {
		log.Printf("cannot find mock method for World.ToModel")
		return
	}
	for _, exp := range r.expect.toModel {
		if !reflect.DeepEqual(exp.value, value) {
			continue
		}
		for _, action := range exp.actions {
			action(value)
		}
		if exp.isAnyTimes {
			r0 = exp.r0
			return
		}
		if exp.requiredTimes > 1 && exp.calledTimes > exp.requiredTimes {
			log.Printf("invalid call times. requiredTimes: [%d] calledTimes: [%d]", exp.requiredTimes, exp.calledTimes)
			return
		}
		exp.calledTimes++
		r0 = exp.r0
		return
	}
	log.Printf("invalid argument World value:[%+v]", value)
	return
}

func (r *WorldExpect) ToModel(value *entity.World) *WorldToModelExpect {
	exp := &WorldToModelExpect{
		actions: []func(value *entity.World){},
		expect:  r,
		value:   value,
	}
	r.toModel = append(r.toModel, exp)
	return exp
}

type WorldToModelsExpect struct {
	expect        *WorldExpect
	isOutOfOrder  bool
	isAnyTimes    bool
	requiredTimes int
	calledTimes   int
	actions       []func(values entity.Worlds)
	values        entity.Worlds
	r0            *model.Worlds
}

func (r *WorldToModelsExpect) Return(r0 *model.Worlds) *WorldToModelsExpect {
	r.r0 = r0
	return r
}

func (r *WorldToModelsExpect) Do(action func(values entity.Worlds)) *WorldToModelsExpect {
	r.actions = append(r.actions, action)
	return r
}

func (r *WorldToModelsExpect) OutOfOrder() *WorldToModelsExpect {
	r.isOutOfOrder = true
	return r
}

func (r *WorldToModelsExpect) AnyTimes() *WorldToModelsExpect {
	r.isAnyTimes = true
	return r
}

func (r *WorldToModelsExpect) Times(n int) *WorldToModelsExpect {
	r.requiredTimes = n
	return r
}

func (r *WorldMock) ToModels(values entity.Worlds) (r0 *model.Worlds) {
	if len(r.expect.toModels) == 0 {
		log.Printf("cannot find mock method for World.ToModels")
		return
	}
	for _, exp := range r.expect.toModels {
		if !reflect.DeepEqual(exp.values, values) {
			continue
		}
		for _, action := range exp.actions {
			action(values)
		}
		if exp.isAnyTimes {
			r0 = exp.r0
			return
		}
		if exp.requiredTimes > 1 && exp.calledTimes > exp.requiredTimes {
			log.Printf("invalid call times. requiredTimes: [%d] calledTimes: [%d]", exp.requiredTimes, exp.calledTimes)
			return
		}
		exp.calledTimes++
		r0 = exp.r0
		return
	}
	log.Printf("invalid argument World values:[%+v]", values)
	return
}

func (r *WorldExpect) ToModels(values entity.Worlds) *WorldToModelsExpect {
	exp := &WorldToModelsExpect{
		actions: []func(values entity.Worlds){},
		expect:  r,
		values:  values,
	}
	r.toModels = append(r.toModels, exp)
	return exp
}

type WorldCreateExpect struct {
	expect        *WorldExpect
	isOutOfOrder  bool
	isAnyTimes    bool
	requiredTimes int
	calledTimes   int
	actions       []func(ctx context.Context, value *entity.World)
	ctx           context.Context
	value         *entity.World
	r0            *model.World
	r1            error
}

func (r *WorldCreateExpect) Return(r0 *model.World, r1 error) *WorldCreateExpect {
	r.r0 = r0
	r.r1 = r1
	return r
}

func (r *WorldCreateExpect) Do(action func(ctx context.Context, value *entity.World)) *WorldCreateExpect {
	r.actions = append(r.actions, action)
	return r
}

func (r *WorldCreateExpect) OutOfOrder() *WorldCreateExpect {
	r.isOutOfOrder = true
	return r
}

func (r *WorldCreateExpect) AnyTimes() *WorldCreateExpect {
	r.isAnyTimes = true
	return r
}

func (r *WorldCreateExpect) Times(n int) *WorldCreateExpect {
	r.requiredTimes = n
	return r
}

func (r *WorldMock) Create(ctx context.Context, value *entity.World) (r0 *model.World, r1 error) {
	if len(r.expect.create) == 0 {
		r1 = xerrors.New("cannot find mock method for World.Create")
		return
	}
	for _, exp := range r.expect.create {
		if !reflect.DeepEqual(exp.ctx, ctx) {
			continue
		}
		if !reflect.DeepEqual(exp.value, value) {
			continue
		}
		for _, action := range exp.actions {
			action(ctx, value)
		}
		if exp.isAnyTimes {
			r0 = exp.r0
			r1 = exp.r1
			return
		}
		if exp.requiredTimes > 1 && exp.calledTimes > exp.requiredTimes {
			r1 = xerrors.Errorf("invalid call times. requiredTimes: [%d] calledTimes: [%d]", exp.requiredTimes, exp.calledTimes)
			return
		}
		exp.calledTimes++
		r0 = exp.r0
		r1 = exp.r1
		return
	}
	r1 = xerrors.Errorf("invalid argument World ctx:[%+v] value:[%+v]", ctx, value)
	return
}

func (r *WorldExpect) Create(ctx context.Context, value *entity.World) *WorldCreateExpect {
	exp := &WorldCreateExpect{
		actions: []func(ctx context.Context, value *entity.World){},
		ctx:     ctx,
		expect:  r,
		value:   value,
	}
	r.create = append(r.create, exp)
	return exp
}

type WorldCreatesExpect struct {
	expect        *WorldExpect
	isOutOfOrder  bool
	isAnyTimes    bool
	requiredTimes int
	calledTimes   int
	actions       []func(ctx context.Context, entities entity.Worlds)
	ctx           context.Context
	entities      entity.Worlds
	r0            *model.Worlds
	r1            error
}

func (r *WorldCreatesExpect) Return(r0 *model.Worlds, r1 error) *WorldCreatesExpect {
	r.r0 = r0
	r.r1 = r1
	return r
}

func (r *WorldCreatesExpect) Do(action func(ctx context.Context, entities entity.Worlds)) *WorldCreatesExpect {
	r.actions = append(r.actions, action)
	return r
}

func (r *WorldCreatesExpect) OutOfOrder() *WorldCreatesExpect {
	r.isOutOfOrder = true
	return r
}

func (r *WorldCreatesExpect) AnyTimes() *WorldCreatesExpect {
	r.isAnyTimes = true
	return r
}

func (r *WorldCreatesExpect) Times(n int) *WorldCreatesExpect {
	r.requiredTimes = n
	return r
}

func (r *WorldMock) Creates(ctx context.Context, entities entity.Worlds) (r0 *model.Worlds, r1 error) {
	if len(r.expect.creates) == 0 {
		r1 = xerrors.New("cannot find mock method for World.Creates")
		return
	}
	for _, exp := range r.expect.creates {
		if !reflect.DeepEqual(exp.ctx, ctx) {
			continue
		}
		if !reflect.DeepEqual(exp.entities, entities) {
			continue
		}
		for _, action := range exp.actions {
			action(ctx, entities)
		}
		if exp.isAnyTimes {
			r0 = exp.r0
			r1 = exp.r1
			return
		}
		if exp.requiredTimes > 1 && exp.calledTimes > exp.requiredTimes {
			r1 = xerrors.Errorf("invalid call times. requiredTimes: [%d] calledTimes: [%d]", exp.requiredTimes, exp.calledTimes)
			return
		}
		exp.calledTimes++
		r0 = exp.r0
		r1 = exp.r1
		return
	}
	r1 = xerrors.Errorf("invalid argument World ctx:[%+v] entities:[%+v]", ctx, entities)
	return
}

func (r *WorldExpect) Creates(ctx context.Context, entities entity.Worlds) *WorldCreatesExpect {
	exp := &WorldCreatesExpect{
		actions:  []func(ctx context.Context, entities entity.Worlds){},
		ctx:      ctx,
		entities: entities,
		expect:   r,
	}
	r.creates = append(r.creates, exp)
	return exp
}

type WorldExpect struct {
	toModel  []*WorldToModelExpect
	toModels []*WorldToModelsExpect
	create   []*WorldCreateExpect
	creates  []*WorldCreatesExpect
}

func NewWorldExpect() *WorldExpect {
	return &WorldExpect{
		create:   []*WorldCreateExpect{},
		creates:  []*WorldCreatesExpect{},
		toModel:  []*WorldToModelExpect{},
		toModels: []*WorldToModelsExpect{},
	}
}