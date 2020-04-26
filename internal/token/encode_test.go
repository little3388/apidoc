// SPDX-License-Identifier: MIT

package token

import (
	"testing"

	"github.com/issue9/assert"
)

func TestEncode(t *testing.T) {
	a := assert.New(t)

	type nestObject struct {
		ID   *intTest `apidoc:"id,elem,usage"`
		Name string   `apidoc:"name,attr,usage"`
	}

	data := []*struct {
		name   string
		object interface{}
		xml    string
		err    bool
	}{
		{},

		{
			name:   "apidoc",
			object: &struct{}{},
			xml:    "<apidoc></apidoc>",
		},

		{
			name: "apidoc",
			object: &struct {
				ID intTest `apidoc:"id,attr,usage"`
			}{
				ID: intTest{Value: 11},
			},
			xml: `<apidoc id="11"></apidoc>`,
		},

		{
			name: "apidoc",
			object: &struct {
				ID   intTest     `apidoc:"id,attr,usage"`
				Name *stringTest `apidoc:",attr,usage"`
			}{
				ID:   intTest{Value: 11},
				Name: &stringTest{Value: "name"},
			},
			xml: `<apidoc id="11" Name="name"></apidoc>`,
		},

		{
			name: "apidoc",
			object: &struct {
				ID   *intTest   `apidoc:"id,attr,usage"`
				Name stringTest `apidoc:"name,elem,usage"`
			}{
				ID:   &intTest{Value: 11},
				Name: stringTest{Value: "name"},
			},
			xml: `<apidoc id="11"><name>name</name></apidoc>`,
		},

		{
			name: "apidoc",
			object: &struct {
				ID    intTest `apidoc:"id,attr,usage"`
				CData CData   `apidoc:",cdata,"`
			}{
				ID:    intTest{Value: 11},
				CData: CData{Value: String{Value: "<h1>h1</h1>"}},
			},
			xml: `<apidoc id="11"><![CDATA[<h1>h1</h1>]]></apidoc>`,
		},

		{
			name: "apidoc",
			object: &struct {
				ID      int     `apidoc:"id,attr,usage"`
				Content *String `apidoc:",content"`
			}{
				ID:      11,
				Content: &String{Value: "<111"},
			},
			xml: `<apidoc id="11">&lt;111</apidoc>`,
		},

		{ // 嵌套
			name: "apidoc",
			object: &struct {
				Object *nestObject `apidoc:"object,elem,usage"`
			}{
				Object: &nestObject{
					ID:   &intTest{Value: 12},
					Name: "name",
				},
			},
			xml: `<apidoc><object name="name"><id>12</id></object></apidoc>`,
		},

		{ // 嵌套 cdata
			name: "apidoc",
			object: &struct {
				Cdata *CData `apidoc:",cdata"`
			}{
				Cdata: &CData{Value: String{Value: "12"}},
			},
			xml: `<apidoc><![CDATA[12]]></apidoc>`,
		},

		{ // 嵌套 content
			name: "apidoc",
			object: &struct {
				Content *String `apidoc:",content"`
			}{
				Content: &String{Value: "11"},
			},
			xml: `<apidoc>11</apidoc>`,
		},
	}

	for i, item := range data {
		xml, err := Encode("", item.name, item.object)

		if item.err {
			a.Error(err, "not error at %d", i).
				Nil(xml, "not nil at %d", i)
			continue
		}

		a.NotError(err, "err %s at %d", err, i).
			Equal(string(xml), item.xml, "not equal at %d\nv1=%s\nv2=%s", i, string(xml), item.xml)
	}

	// content 和 cdata 的类型不正确
	a.Panic(func() {
		Encode("", "root", &struct {
			Content string `apidoc:",content"`
		}{})
	})
}
