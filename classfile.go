package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

///// file io / byte io /////
func file_read(fname string) *bytes.Buffer {
	fp, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	var b bytes.Buffer
	buf := make([]byte, 10)
	for {
		n, err := fp.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			panic(err)
		}
		b.Write(buf[0:n])
	}
	return &b
}
func read(b *bytes.Buffer, n int) []uint8 {
	return b.Next(n)
}
func read_uint32(b *bytes.Buffer, n uint32) []byte {
	var result []byte
	var i uint32
	for i = 0; i < n; i++ {
		x := b.Next(1)[0]
		result = append(result, x)
	}
	return result
}
func read_u1(b *bytes.Buffer) uint8 {
	foo := b.Next(1)
	return uint8(foo[0])
}
func read_u2(b *bytes.Buffer) uint16 {
	bs := b.Next(2)
	return binary.BigEndian.Uint16(bs)
}
func read_u4(b *bytes.Buffer) uint32 {
	bs := b.Next(4)
	return binary.BigEndian.Uint32(bs)
}

///// class file structure /////
type ClassFile struct {
	magic               []uint8   //uint32 //u4 magic;
	minor_version       uint16    //u2 minor_version;
	major_version       uint16    //u2 major_version;
	constant_pool_count uint16    //u2 constant_pool_count;
	constant_pool       []CP_INFO //cp_info constant_pool[constant_pool_count-1];
	access_flags        uint16    //u2 access_flags;
	this_class          uint16    //u2 this_class;
	super_class         uint16    //u2 super_class;
	interfaces_count    uint16    //u2 interfaces_count;
	//u2 interfaces[interfaces_count];
	fields_count uint16 //u2 fields_count;
	//field_info fields[fields_count];
	methods_count    uint16           //u2 methods_count;
	method_info      []METHOD_INFO    //methods[methods_count];
	attributes_count uint16           //u2 attributes_count;
	attribute_info   []ATTRIBUTE_INFO //attributes[attributes_count];
}

type info struct {
	name string
	val  []uint8
}

type CP_INFO struct {
	name  string
	tag   uint8
	infos []info
}

type METHOD_INFO struct {
	access_flags     uint16 //u2
	name_index       uint16 //u2
	descriptor_index uint16 //u2
	attributes_count uint16 //u2
	attribute_info   []ATTRIBUTE_INFO
}
type ATTRIBUTE_INFO struct {
	attribute_name_index uint16  //u2
	attribute_length     uint32  //u4
	info                 []uint8 //u1
}

///// class file toString() /////
func (cp CP_INFO) String() string {
	str := fmt.Sprintf("<%s(%d) ", cp.name, cp.tag)
	for _, i := range cp.infos {
		str += fmt.Sprintf("[%s:%v]", i.name, i.val)
	}
	if cp.tag == uint8(1) {
		return str + string(cp.infos[1].val) + ">"
	} else {
		return str + ">"
	}
}
func (m METHOD_INFO) String() string {
	result := "<"
	result += "flags:" + str_access_flag(m.access_flags, "m_a_p") + " "
	result += fmt.Sprintf("name_idx:%d ", m.name_index)
	result += fmt.Sprintf("desc_idx:%d ", m.descriptor_index)
	var i uint16 = 0
	for i = 0; i < m.attributes_count; i++ {
		result += fmt.Sprintf("attr[%d]:%s ", i, m.attribute_info[i])
	}
	return result + ">"
}
func (a ATTRIBUTE_INFO) String() string {
	result := "<"
	result += fmt.Sprintf("attribute_name_index:%d ", a.attribute_name_index)
	for i := 0; i < len(a.info); i++ {
		result += fmt.Sprintf("%02X ", a.info[i])
	}
	return result + ">"
}

///// read /////
func read_class_file(b *bytes.Buffer) ClassFile {
	var cf ClassFile
	cf.magic = read_CAFEBABE(b)
	cf.major_version = read_u2(b)
	cf.minor_version = read_u2(b)
	cf.constant_pool_count = read_u2(b)
	cf.constant_pool = read_CP_INFO(b, cf.constant_pool_count)
	cf.access_flags = read_u2(b)
	cf.this_class = read_u2(b)
	cf.super_class = read_u2(b)
	cf.interfaces_count = read_u2(b)
	if cf.interfaces_count != 0 {
		panic("interface is not supported")
	}
	cf.fields_count = read_u2(b)
	if cf.fields_count != 0 {
		panic("field is not supported")
	}
	cf.methods_count = read_u2(b)
	cf.method_info = read_METHOD_INFO(b, cf.methods_count)
	cf.attributes_count = read_u2(b)
	cf.attribute_info = read_ATTRIBUTE(b, cf.attributes_count)
	return cf
}
func read_CAFEBABE(b *bytes.Buffer) []uint8 {
	return read(b, 4)
}
func read_CP_INFO(b *bytes.Buffer, count uint16) []CP_INFO {
	result := []CP_INFO{}
	for i := uint16(1); i < count; i++ {
		cp := read_cp(b)
		result = append(result, cp)
	}
	return result
}

func read_cp(b *bytes.Buffer) CP_INFO {
	var cp CP_INFO
	cp.tag = read_u1(b)
	switch cp.tag {
	case 7:
		cp.name = "Class"
		cp.infos = []info{{"name_index", read(b, 2)}}
	case 9: // [:Fieldref,[2,2]],
		cp.name = "Fieldref"
		cp.infos = []info{{"class_index", read(b, 2)},
			{"name_and_type_index;", read(b, 2)}}
	case 10: //[:Methodref,[2,2]],
		cp.name = "Methodref"
		cp.infos = []info{{"class_index", read(b, 2)},
			{"name_and_type_index", read(b, 2)}}
	case 11: //[:InterfaceMethodref,[2,2]],
		cp.name = "InterfaceMethodref"
		cp.infos = []info{{"class_index", read(b, 2)},
			{"name_and_type_index;", read(b, 2)}}
	case 8: //[:String,[2]],
		cp.name = "String"
		cp.infos = []info{{"string_index", read(b, 2)}}
	case 3: // [:Integer,[4]],
		cp.name = "Integer"
		cp.infos = []info{{"bytes", read(b, 4)}}
	case 4: //[:Float,[4]],
		cp.name = "Float"
		cp.infos = []info{{"bytes;", read(b, 4)}}
	case 5: //[:Long,[4,4]],
		cp.name = "Long"
		cp.infos = []info{{"high_bytes", read(b, 4)},
			{"low_bytes", read(b, 4)}}
	case 6: //[:Double,[4,4]],
		cp.name = "Double"
		cp.infos = []info{{"high_bytes", read(b, 4)},
			{"low_bytes", read(b, 4)}}
	case 12: //[:NameAndType,[2,2]],
		cp.name = "NameAndType"
		cp.infos = []info{{"name_index", read(b, 2)},
			{"descriptor_index", read(b, 2)}}
	case 1: //[:Utf8,[2]], #
		cp.name = "Utf8"
		size := read(b, 2)
		cp.infos = []info{{"length", size}}
		i := int(binary.BigEndian.Uint16(size))
		foo := info{"bytes", read(b, i)}
		cp.infos = append(cp.infos, foo)
	case 15: //[:MethodHandle,[1,2]],
		cp.name = "MethodHandle"
		cp.infos = []info{{"reference_kind", read(b, 1)},
			{"reference_index", read(b, 2)}}
	case 16: // [:MethodType,[2]],
		cp.name = "MethodType"
		cp.infos = []info{{"descriptor_index", read(b, 2)}}
	case 18: //[:InvokeDynamic,[2,2]]}
		cp.name = "InvokeDynamic"
		cp.infos = []info{{"bootstrap_method_attr_index", read(b, 2)},
			{"name_and_type_index", read(b, 2)}}
	default:
		panic(fmt.Sprintf("cp_info(%d)", cp.tag))
	}
	return cp
}
func read_METHOD_INFO(b *bytes.Buffer, count uint16) []METHOD_INFO {
	result := []METHOD_INFO{}
	for i := uint16(0); i < count; i++ {
		m := read_method(b)
		result = append(result, m)
	}
	return result
}
func read_method(b *bytes.Buffer) METHOD_INFO {
	var meth METHOD_INFO
	meth.access_flags = read_u2(b)
	meth.name_index = read_u2(b)
	meth.descriptor_index = read_u2(b)
	meth.attributes_count = read_u2(b)
	meth.attribute_info = read_ATTRIBUTE(b, meth.attributes_count)
	return meth
}
func read_ATTRIBUTE(b *bytes.Buffer, count uint16) []ATTRIBUTE_INFO {
	result := []ATTRIBUTE_INFO{}
	for i := uint16(0); i < count; i++ {
		attr := read_attibute(b)
		result = append(result, attr)
	}
	return result
}
func read_attibute(b *bytes.Buffer) ATTRIBUTE_INFO {
	var attr ATTRIBUTE_INFO
	attr.attribute_name_index = read_u2(b)
	attr.attribute_length = read_u4(b)
	attr.info = read_uint32(b, attr.attribute_length)
	return attr
}

///// print  /////
func print(cf ClassFile) {
	fmt.Printf("MAGIC:%X%X%X%X\n", cf.magic[0], cf.magic[1], cf.magic[2], cf.magic[3])
	fmt.Println("MAJOR:", cf.major_version)
	fmt.Println("NINOR:", cf.minor_version)
	fmt.Println("CONSTANT_POOL_COUNT:", cf.constant_pool_count)
	for i := 0; i < len(cf.constant_pool); i++ {
		fmt.Println("CONSTANT_POOL:", i+1, cf.constant_pool[i])
	}
	fmt.Println("access_flags:", cf.access_flags, str_access_flag(cf.access_flags, "class"))
	fmt.Println("this_class:", cf.this_class)
	fmt.Println("super_class:", cf.super_class)
	fmt.Println("interfaces_count:", cf.interfaces_count)
	// interface_info
	fmt.Println("fields_count:", cf.fields_count)
	// field_info
	fmt.Println("methods_count:", cf.methods_count)
	for i := 0; i < len(cf.method_info); i++ {
		fmt.Printf("method[%d]:%s\n", i, cf.method_info[i])
	}
	fmt.Println("attributes_count:", cf.attributes_count)
	for i := 0; i < len(cf.attribute_info); i++ {
		fmt.Printf("attr[%d]:%s\n", i, cf.attribute_info[i])
	}
}

func str_access_flag(x uint16, opt string) string {
	result := ""
	if (x & 0x0001) != 0 { //ACC_PUBLIC	0x0001
		result += "public "
	}
	if (x & 0x0002) != 0 { //ACC_PRIVATE	0x0002
		result += "private "
	}
	if (x & 0x0004) != 0 { //ACC_PROTECTED	0x0004
		result += "protected "
	}
	if (x & 0x0008) != 0 { //ACC_STATIC	0x0008
		result += "static "
	}
	if (x & 0x0010) != 0 { //ACC_FINAL	0x0010
		result += "final "
	}
	if (x & 0x0100) != 0 { //ACC_NATIVE
		result += "native "
	}
	if (x & 0x0200) != 0 { //ACC_INTERFACE
		result += "interface "
	}
	if (x & 0x0400) != 0 { //ACC_ABSTRACT
		result += "abstract "
	}
	if (x & 0x0800) != 0 { //ACC_STRICT
		result += "strict "
	}
	if (x & 0x1000) != 0 { //ACC_SYNTHETIC
		result += "synthetic "
	}
	if (x & 0x2000) != 0 { //ACC_ANNOTATION
		result += "annotation "
	}
	if (x & 0x4000) != 0 { //ACC_ENUM
		result += "enum "
	}
	///////////////////////////////////////
	if (x&0x0080) != 0 && (opt == "field") { //ACC_TRANSIENT
		result += "transient "
	}
	if (x&0x0040) != 0 && (opt == "field") { //ACC_VOLATILE
		result += "volatile "
	}
	if (x&0x0020) != 0 && (opt == "class") { //ACC_SUPER
		result += "super "
	}
	if (x&0x0020) != 0 && (opt == "m_a_p") { //ACC_SYNCRONIZED
		result += "super "
	}
	if (x&0x0040) != 0 && (opt == "m_a_p") { //ACC_BRIDGE
		result += "bridge "
	}
	if (x&0x0080) != 0 && (opt == "m_a_p") { //ACC_VARARGS
		result += "varargs "
	}
	return result
}

///// as []byte /////
func u2_to_bytes(u uint16) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint16(bytes, u)
	return bytes
}
func as_byte(cf ClassFile) []byte {
	result := []byte{}
	result = append(result, cf.magic...)
	result = append(result, u2_to_bytes(cf.major_version)...)
	result = append(result, u2_to_bytes(cf.minor_version)...)
	result = append(result, u2_to_bytes(cf.constant_pool_count)...)
	for i := 0; i < len(cf.constant_pool); i++ {
		result = append(result, cf.constant_pool[i].tag)
		for j := 0; j < len(cf.constant_pool[i].infos); j++ {
			result = append(result, cf.constant_pool[i].infos[j].val...)
		}
	}
	result = append(result, u2_to_bytes(cf.access_flags)...)
	result = append(result, u2_to_bytes(cf.this_class)...)
	result = append(result, u2_to_bytes(cf.super_class)...)
	result = append(result, u2_to_bytes(cf.interfaces_count)...)
	result = append(result, u2_to_bytes(cf.fields_count)...)
	result = append(result, u2_to_bytes(cf.methods_count)...)
	for i := 0; i < len(cf.method_info); i++ {
		//under constructing
	}

	return result
}

///// main /////
func main() {
	filename := "src/Hello.class"
	b := file_read(filename)
	cf := read_class_file(b)
	print(cf)

	fmt.Println("aaaaaaaaaaaaaaaaaaaaaaa")

	u64 := uint64(12345)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, u64)
	fmt.Printf("%v\n", bytes)

	bytes = []byte{0xCA, 0xFE, 0xBA, 0xBE}
	u32 := binary.BigEndian.Uint32(bytes)
	fmt.Printf("%X\n", u32)

}
