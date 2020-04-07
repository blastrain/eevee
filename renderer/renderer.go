package renderer

import (
	. "go.knocknote.io/eevee/code"
	"go.knocknote.io/eevee/types"
)

type RendererHelper interface {
	Receiver() *Statement
	Field(string) *Statement
	MethodCall(string, ...Code) *Statement
	Package(string) string
	GetClass() *types.Class
	GetImportList() types.ImportList
	IsModelPackage() bool
	CreateMethodDeclare() *types.MethodDeclare
	CreateCollectionMethodDeclare() *types.MethodDeclare
}

type Renderer interface {
	Render(RendererHelper) *types.Method
	RenderWithOption(RendererHelper) *types.Method
	RenderCollection(RendererHelper) *types.Method
	RenderCollectionWithOption(RendererHelper) *types.Method
	Marshaler(RendererHelper) *types.Method
	MarshalerContext(RendererHelper) *types.Method
	MarshalerCollection(RendererHelper) *types.Method
	MarshalerCollectionContext(RendererHelper) *types.Method
	Unmarshaler(RendererHelper) *types.Method
	UnmarshalerCollection(RendererHelper) *types.Method
}
