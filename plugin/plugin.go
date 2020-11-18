package interfacetype

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity"

	wrapper "github.com/marbar3778/go-proto-wrapper"
)

type plugin struct {
	*generator.Generator
	generator.PluginImports

	fmtPkg generator.Single

	gogoImport bool
}

func NewPlugin(gogoImport bool) *plugin {
	return &plugin{gogoImport: gogoImport}
}

func (p *plugin) Name() string {
	return "wrapper"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
}

func GetWrapper(msg *descriptor.DescriptorProto) bool {
	if msg == nil {
		return false
	}
	if msg.Options != nil {
		return proto.GetBoolExtension(msg, wrapper.E_MsgWrapper, true)
	}
	return false
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	if !p.gogoImport {
		vanity.TurnOffGogoImport(file.FileDescriptorProto)
	}

	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.fmtPkg = p.NewImport("fmt")

	for _, message := range file.Messages() {
		if !GetWrapper(message.DescriptorProto) {
			continue
		}
		if len(message.OneofDecl) != 1 {
			panic("wrapper only supports messages with exactly one oneof declaration")
		}
		for _, field := range message.Field {
			if idx := field.OneofIndex; idx == nil || *idx != 0 {
				panic("all fields in wrapper message must belong to the oneof")
			}
		}

		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		p.P(`func (this *`, ccTypeName, `) Unwrap() (proto.Message, error) {`)
		p.In()
		p.P(`switch msg := this.Sum.(type) {`)
		p.In()
		for _, field := range message.Field {
			fieldname := p.GetOneOfFieldName(message, field)
			structName := p.OneOfTypeName(message, field)
			// goTyp, _ := p.GoType(message, field)
			p.P(`case *`, structName, `:`)
			p.In()
			p.P(`return this.Get`, fieldname, `(), nil`)
			p.Out()
		}
		p.P("default:")
		p.In()
		p.P(`return nil, fmt.Errorf("unknown message: %T", msg)`)
		p.Out()
		p.P(`}`)
		p.Out()
		p.P(`}`)
		p.P(``)

		p.P(`func (this *`, ccTypeName, `) Wrap(msg proto.Message) error {`)
		p.In()
		p.P(`if msg == nil {`)
		p.In()
		p.P(`this.`, p.GetFieldName(message, message.Field[0]), ` = nil`)
		p.P(`return nil`)
		p.Out()
		p.P("}")
		p.P(`switch vt := msg.(type) {`)
		p.In()
		for _, field := range message.Field {
			oneofName := p.GetFieldName(message, field)
			structName := p.OneOfTypeName(message, field)
			goTyp, _ := p.GoType(message, field)
			p.P(`case `, goTyp, `:`)
			p.In()
			p.P(`this.`, oneofName, ` = &`, structName, `{vt}`)
			p.Out()
		}
		p.P("default:")
		p.In()
		p.P(`return fmt.Errorf("unknown message: %T", msg)`)
		p.Out()
		p.P(`}`)
		p.P("return nil")
		p.Out()
		p.P(`}`)
		p.P(``)
	}
}

// func init() {
// 	generator.RegisterPlugin(NewInterfaceType())
// }
