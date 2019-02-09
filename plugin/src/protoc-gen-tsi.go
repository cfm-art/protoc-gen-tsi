package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"./com"

	// protocのパッケージも公開されているようなので
	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// genTypeName is 最初の.自身のパッケージを取り除く
func genTypeName(file *descriptor.FileDescriptorProto, typeName string) string {
	// .<パッケージ>.<型>
	// .<型>
	var reg *regexp.Regexp
	if file.Package != nil {
		reg = regexp.MustCompile(`^\.(` + strings.Replace(*file.Package, "/", ".", -1) + `\.)?(.+)`)
	} else {
		reg = regexp.MustCompile(`^\.(.+)`)
	}
	v := reg.ReplaceAllString(typeName, "$2")
	return strings.Replace(v, ".", "_", strings.Count(v, ".")-1) // 最後の1個の.を残して後は_へ
}

// getLabelText is 各フィールドの型にOptional/Arrayを付与
func getLabelText(file *descriptor.FileDescriptorProto, fieldName string, typeName string, fieldLabel descriptor.FieldDescriptorProto_Label) string {
	if fieldLabel == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
		// OPTIONAL is ?
		return fieldName + "?" + " : " + genTypeName(file, typeName)
	} else if fieldLabel == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
		// !
		return fieldName + " : " + genTypeName(file, typeName)
	} else if fieldLabel == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// REPEATED is []
		return fieldName + " : " + genTypeName(file, typeName) + "[]"
	} else {
		return fieldName + "?" + " : " + genTypeName(file, typeName)
	}
}

// getTypeText is 各フィールドの型名の解決
func getTypeText(typeName *string, fieldType descriptor.FieldDescriptorProto_Type) string {
	if fieldType == descriptor.FieldDescriptorProto_TYPE_DOUBLE {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_FLOAT {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_INT64 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_UINT64 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_INT32 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_FIXED64 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_FIXED32 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_BOOL {
		return "boolean"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_STRING {
		return "string"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_GROUP {
		// 謎
		return "any"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		return *typeName
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_BYTES {
		return "Uint8Array"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_UINT32 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_ENUM {
		return *typeName
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_SFIXED32 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_SFIXED64 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_SINT32 {
		return "number"
	} else if fieldType == descriptor.FieldDescriptorProto_TYPE_SINT64 {
		return "number"
	} else {
		return "any"
	}
}

// makeMessageField is 各フィールドを解決
func makeMessageField(file *descriptor.FileDescriptorProto, field *descriptor.FieldDescriptorProto) string {
	fieldName := *field.Name
	typeName := getTypeText(field.TypeName, *field.Type) // FieldDescriptorProto_Type
	return getLabelText(file, fieldName, typeName, *field.Label)
}

/*
export interface <ClassName>
{
____<Field>;
}
*/

// makeMessageType is 構造体を解決
func makeMessageType(option com.Option, file *descriptor.FileDescriptorProto, message *descriptor.DescriptorProto) string {
	result := "export interface " + *message.Name + "\n{\n"
	for _, field := range message.Field {
		fieldText := makeMessageField(file, field)
		result += "\t" + fieldText + ";\n"
	}
	result += "}\n"
	return result
}

/*
export enum <ClassName>
{
____<Name> = <Value>,;
}
*/

// makeEnumType is 列挙型を解決
func makeEnumType(option com.Option, file *descriptor.FileDescriptorProto, en descriptor.EnumDescriptorProto) string {
	result := "export enum " + *en.Name + "\n{\n"
	for _, value := range en.Value {
		result += "\t" + *value.Name + " = " + strconv.FormatInt(int64(*value.Number), 10) + ",\n"
	}
	result += "}\n"
	return result
}

/*
export interface <ClassName>
{
____async <Method>(input: <Input>): Promise<<Output>>;
}

export class <ClassName>Client
____implements <ClassName>
{
____constructor()
____<Method>
}
*/

// makeServiceImpl is 実際に通信する部分をfetchで仮作成
func makeServiceClientAsFetch(file *descriptor.FileDescriptorProto, sv descriptor.ServiceDescriptorProto) string {
	// class
	result := "export class " + *sv.Name + "Client\n\timplements " + *sv.Name + "\n{\n"
	// default constructor
	result += `
	constructor(
		private basePath: string
				= '',                                          // default
		private makeUrl: (basePath: string, packageName: string, className:string, methodName: string) => string
				= (b, _, c, m) => ` + "`${b}${c}/${m}`" + `,             // =default
		private makeHeaders: (baseHeaders: {}) => {}
				= (h) => h,                                    // =default
		private makeQuery: (baseQuery: {}) => {}
				= (q) => q                                     // =default
	) {}

	`

	for _, method := range sv.Method {
		name := *method.Name
		input := genTypeName(file, *method.InputType)
		output := genTypeName(file, *method.OutputType)
		result += `
	async ` + name + ` (input: ` + input + `) : Promise<` + output + `> {
		const url = this.makeUrl(this.basePath, PackageName, '` + *sv.Name + `', '` + name + `');
		const headers = this.makeHeaders({'Content-Type': 'application/json'});
		const query = this.makeQuery({
			method: 'POST',
			headers,
			body: JSON.stringify(input)
		});
		const response = await fetch(url, query);
		return (await response.json()) as ` + output + `;
	}
	`
	}
	result += "\n}\n\n"
	return result
}

// makeServiceImpl is 実際に通信する部分をXMLHttpRequest で仮作成
func makeServiceClientAsAjax(file *descriptor.FileDescriptorProto, sv descriptor.ServiceDescriptorProto) string {
	log.Fatalf("not implemented 'makeServiceClientAsAjax'")
	return ""
}

// makeService is エンドポイントを解決
func makeService(option com.Option, file *descriptor.FileDescriptorProto, sv descriptor.ServiceDescriptorProto) string {
	// interface
	result := "export interface " + *sv.Name + "\n{\n"
	for _, method := range sv.Method {
		name := *method.Name
		input := genTypeName(file, *method.InputType)
		output := genTypeName(file, *method.OutputType)
		result += "\t" + name + "(input: " + input + ") : Promise<" + output + ">;\n"
	}
	result += "}\n\n"
	if option.GenClient {
		if option.ClientType == "fetch" {
			result += makeServiceClientAsFetch(file, sv)
		} else if option.ClientType == "ajax" {
			result += makeServiceClientAsAjax(file, sv)
		}
	}
	return result
}

//  makeEnumTypes is 列挙型を全て解決
func makeEnumTypes(option com.Option, f *descriptor.FileDescriptorProto) string {
	content := ""
	for _, e := range f.EnumType {
		content += makeEnumType(option, f, *e)
	}
	return content
}

// makeMessageTypes is 構造体を全て解決
func makeMessageTypes(option com.Option, f *descriptor.FileDescriptorProto) string {
	content := ""
	for _, message := range f.MessageType {
		classText := makeMessageType(option, f, message)
		content += classText + "\n"
	}
	return content
}

// makeServices is エンドポイントを全て解決
func makeServices(option com.Option, f *descriptor.FileDescriptorProto) string {
	content := ""
	for _, s := range f.Service {
		content += makeService(option, f, *s)
	}
	return content
}

// makeDependencies is 依存関係の解決
func makeDependencies(option com.Option, files map[string]*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) string {
	content := ""
	for _, s := range f.Dependency {
		// import * as <Hoge> from './<Hoge>.proto;'
		dep := files[s]
		content += "import * as " + strings.Replace(*dep.Package, ".", "_", -1) + " from './" + s + "';\n"
	}
	return content
}

// process is 変換処理取りまとめ
func process(req *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	option := com.ParseArgument(req)

	files := make(map[string]*descriptor.FileDescriptorProto)
	for _, f := range req.ProtoFile {
		files[f.GetName()] = f
	}

	var res plugin.CodeGeneratorResponse
	for _, fname := range req.FileToGenerate {
		f := files[fname]
		out := fname + ".ts"
		packageName := ""

		if f.Package != nil {
			packageName = *f.Package
		}

		// 種別毎に並列で文字列を生成
		enumText := make(chan string)
		messageText := make(chan string)
		serviceText := make(chan string)

		depText := makeDependencies(option, files, f)

		go func() {
			enumText <- makeEnumTypes(option, f)
		}()
		go func() {
			messageText <- makeMessageTypes(option, f)
		}()
		go func() {
			serviceText <- makeServices(option, f)
		}()

		// 同期して結合
		content := depText + "\n\n" +
			"export const PackageName = '" + packageName + "';\n\n" +
			<-enumText + "\n" +
			<-messageText + "\n" +
			<-serviceText + "\n"

		res.File = append(res.File, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(out),
			Content: proto.String(content),
		})
	}
	return &res
}

// run is プロトコルバッファの情報を標準入力から取得し、
//        tsファイルを生成して標準出力へ
func run() error {
	req, err := com.ReadFrom(os.Stdin)
	if err != nil {
		return nil
	}
	res := process(req)
	return com.WriteTo(res, os.Stdout)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
