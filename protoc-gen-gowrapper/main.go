package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	wrapper_plugin "github.com/marbar3778/go-proto-wrapper/plugin"
)

func main() {
	gen := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		gen.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, gen.Request); err != nil {
		gen.Error(err, "parsing input proto")
	}

	if len(gen.Request.FileToGenerate) == 0 {
		gen.Fail("no files to generate")
	}

	useGogoImport := false
	// Match parsing algorithm from Generator.CommandLineParameters
	for _, parameter := range strings.Split(gen.Request.GetParameter(), ",") {
		kvp := strings.SplitN(parameter, "=", 2)
		// We only care about key-value pairs where the key is "gogoimport"
		if len(kvp) != 2 || kvp[0] != "gogoimport" {
			continue
		}
		useGogoImport, err = strconv.ParseBool(kvp[1])
		if err != nil {
			gen.Error(err, "parsing gogoimport option")
		}
	}

	var targets []string
	for _, target := range gen.Request.FileToGenerate {
		// var re = regexp.MustCompile(`\b(\w*msg_wrapper\w*)\b`)
		b, err := ioutil.ReadFile(target)
		if err != nil {
			panic(err)
		}
		s := string(b)

		//check whether s contains substring text
		if strings.Contains(s, "msg_wrapper") {
			targets = append(targets, target)
		}
	}

	gen.Request.FileToGenerate = targets

	gen.CommandLineParameters(gen.Request.GetParameter())

	gen.WrapTypes()
	gen.SetPackageNames()
	gen.BuildTypeNameMap()
	gen.GeneratePlugin(wrapper_plugin.NewPlugin(useGogoImport))

	for i := 0; i < len(gen.Response.File); i++ {
		gen.Response.File[i].Name = proto.String(strings.Replace(*gen.Response.File[i].Name, ".pb.go", ".wrapper.pb.go", -1))
	}

	// Send back the results.
	data, err = proto.Marshal(gen.Response)
	if err != nil {
		gen.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		gen.Error(err, "failed to write output proto")
	}
}
