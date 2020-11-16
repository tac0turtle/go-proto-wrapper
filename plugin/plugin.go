package interfacetype

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity"
	wrapper "github.com/marbar3778/go-proto-wrapper"
)

type plugin struct {
	*generator.Generator
	generator.PluginImports

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

	panic(fmt.Sprintf("%s", msg.Options))
	if msg.Options != nil {
		v := proto.GetBoolExtension(msg.Options, wrapper.E_MsgWrapper, false)
		return v
	}

	return false
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	if !p.gogoImport {
		vanity.TurnOffGogoImport(file.FileDescriptorProto)
	}

	p.PluginImports = generator.NewPluginImports(p.Generator)

	for _, message := range file.Messages() {
		iface := GetWrapper(message.DescriptorProto)
		if !iface {
			panic("not")
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
		p.P(`func (this *`, ccTypeName, `) Unwrap() proto.message {`)
		p.In()
		for _, field := range message.Field {
			fieldname := p.GetOneOfFieldName(message, field)
			if fieldname == "Value" {
				panic("cannot have a onlyone message " + ccTypeName + " with a field named Value")
			}
			p.P(`if x := this.Get`, fieldname, `(); x != nil {`)
			p.In()
			p.P(`return x`)
			p.Out()
			p.P(`}`)
		}
		p.P(`return nil`)
		p.Out()
		p.P(`}`)
		p.P(``)
		p.P(`func (this *`, ccTypeName, `) Wrap(value proto.message) error {`)
		p.In()
		p.P(`if value == nil {`)
		p.In()
		p.P(`this.`, p.GetFieldName(message, message.Field[0]), ` = nil`)
		p.P(`return nil`)
		p.Out()
		p.P("}")
		p.P(`switch vt := value.(type) {`)
		p.In()
		for _, field := range message.Field {
			oneofName := p.GetFieldName(message, field)
			structName := p.OneOfTypeName(message, field)
			goTyp, _ := p.GoType(message, field)
			p.P(`case `, goTyp, `:`)
			p.In()
			p.P(`this.`, oneofName, ` = &`, structName, `{vt}`)
			p.P("return nil")
			p.Out()
		}
		p.P(`}`)
		p.P(`return fmt.Errorf("can't encode value of type %T as message `, ccTypeName, `", value)`)
		p.Out()
		p.P(`}`)
		p.P(``)
	}
}

// func init() {
// 	generator.RegisterPlugin(NewInterfaceType())
// }
