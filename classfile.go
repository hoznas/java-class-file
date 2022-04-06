package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

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

func read_CAFEBABE(bu *bytes.Buffer) string {
	a := bu.Next(1)
	b := bu.Next(1)
	c := bu.Next(1)
	d := bu.Next(1)
	return fmt.Sprintf("%X%X%X%X\n", a, b, c, d)
}
func read(b *bytes.Buffer, n int) []uint8 {
	return b.Next(n)
}
func read_uint32(b *bytes.Buffer, n uint32) []uint8 {

	var result []uint8
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

type info struct {
	name string
	val  []uint8
}

type CP_INFO struct {
	name  string
	tag   uint8
	infos []info
}

func (cp CP_INFO) String() string {
	str := fmt.Sprintf("<%s(%d) ", cp.name, cp.tag)
	for _, i := range cp.infos {
		str += fmt.Sprintf("[%s:%v]", i.name, i.val)
	}
	return str + ">"

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

func read_CP_INFO(b *bytes.Buffer, count uint16) []CP_INFO {
	result := []CP_INFO{}
	for i := uint16(1); i < count; i++ {
		fmt.Println("**************")
		cp := read_cp(b)
		fmt.Println(i, cp)
		result = append(result, cp)
	}
	return result
}

func main() {
	filename := "src/Hello.class"
	b := file_read(filename)

	fmt.Println("MAGIC:", read_CAFEBABE(b))
	fmt.Println("MAJOR:", read_u2(b))
	fmt.Println("NINOR:", read_u2(b))
	cp := read_u2(b)
	fmt.Println("CONSTANT_POOL:", cp)
	cps := read_CP_INFO(b, cp)
	fmt.Println("CPS:", cps)

	fmt.Println("access_flags:", read_u2(b))
	fmt.Println("this_class:", read_u2(b))
	fmt.Println("super_class:", read_u2(b))
	fmt.Println("interfaces_count:", read_u2(b))
	// read_ifaces()
	fmt.Println("fields_count:", read_u2(b))
	// read_fields()
	method_count := read_u2(b)
	fmt.Println("methods_count:", method_count)
	read_METHOD_INFO(b, method_count)
	attributes_count := read_u2(b)
	fmt.Println("attributes_count:", attributes_count)
	read_ATTRIBUTE(b, attributes_count)
}
func read_METHOD_INFO(b *bytes.Buffer, count uint16) []METHOD_INFO {
	result := []METHOD_INFO{}
	for i := uint16(0); i < count; i++ {
		fmt.Println("**************")
		m := read_method(b)
		fmt.Println(i, m)
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
		fmt.Println("**************")
		attr := read_attibute(b)
		fmt.Println(attr)
		result = append(result, attr)
	}
	return result
}
func read_attibute(b *bytes.Buffer) ATTRIBUTE_INFO {
	var attr ATTRIBUTE_INFO
	attr.attribute_name_index = read_u2(b)
	fmt.Println("attr.attribute_name_index", attr.attribute_name_index)
	attr.attribute_length = read_u4(b)
	fmt.Println("attr.attribute_length", attr.attribute_length)
	attr.info = read_uint32(b, attr.attribute_length)
	return attr
}

type ClassFile struct {
	magic               uint32 //u4 magic;
	minor_version       uint16 //u2 minor_version;
	major_version       uint16 //u2 major_version;
	constant_pool_count uint16 //u2 constant_pool_count;
	//constant_pool []CP_INFO		//cp_info constant_pool[constant_pool_count-1];
	access_flags     uint16 //u2 access_flags;
	this_class       uint16 //u2 this_class;
	super_class      uint16 //u2 super_class;
	interfaces_count uint16 //u2 interfaces_count;
	// 				 									//u2 interfaces[interfaces_count];
	fields_count uint16 //u2 fields_count;
	//field_info fields[fields_count];
	methods_count uint16 //u2 methods_count;
	//method_info methods[methods_count];
	attributes_count uint16 //u2 attributes_count;
	//attribute_info attributes[attributes_count];
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
