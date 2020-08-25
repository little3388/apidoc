// SPDX-License-Identifier: MIT

package ast

import (
	"strconv"

	"github.com/issue9/is"
	"github.com/issue9/sliceutil"

	"github.com/caixw/apidoc/v7/core"
	"github.com/caixw/apidoc/v7/internal/locale"
	"github.com/caixw/apidoc/v7/internal/xmlenc"
)

// Sanitize token.Sanitizer
func (api *API) Sanitize(p *xmlenc.Parser) {
	for _, header := range api.Headers { // 报头不能为 object
		if header.Type.V() == TypeObject {
			p.Error(p.NewError(header.Type.Start, header.Type.End, "header", locale.ErrInvalidValue))
		}
	}

	// 对 Servers 和 Tags 查重
	indexes := sliceutil.Dup(api.Servers, func(i, j int) bool { return api.Servers[i].V() == api.Servers[j].V() })
	if len(indexes) > 0 {
		err := p.NewError(api.Servers[indexes[0]].Start, api.Servers[indexes[0]].End, "server", locale.ErrDuplicateValue)
		for _, srv := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: api.Servers[srv].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		p.Error(err)
	}
	indexes = sliceutil.Dup(api.Tags, func(i, j int) bool { return api.Tags[i].V() == api.Tags[j].V() })
	if len(indexes) > 0 {
		err := p.NewError(api.Tags[indexes[0]].Start, api.Tags[indexes[0]].End, "server", locale.ErrDuplicateValue)
		for _, tag := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: api.Tags[tag].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		p.Error(err)
	}
}

// Sanitize token.Sanitizer
func (e *Enum) Sanitize(p *xmlenc.Parser) {
	if e.Description.V() == "" && e.Summary.V() == "" {
		p.Error(p.NewError(e.Start, e.End, "summary", locale.ErrIsEmpty, "summary"))
	}
}

// Sanitize token.Sanitizer
func (p *Path) Sanitize(pp *xmlenc.Parser) {
	if p.Path == nil || p.Path.V() == "" {
		pp.Error(pp.NewError(p.Start, p.End, "path", locale.ErrIsEmpty, "path"))
	}

	params, err := parsePath(p.Path.V())
	if err != nil {
		pp.Error(pp.NewError(p.Path.Start, p.Path.End, "path", locale.ErrInvalidFormat))
	}
	if len(params) != len(p.Params) {
		pp.Error(pp.NewError(p.Start, p.End, "path", locale.ErrPathNotMatchParams))
	}
	for _, param := range p.Params {
		if _, found := params[param.Name.V()]; !found {
			pp.Error(pp.NewError(param.Start, param.End, "path", locale.ErrPathNotMatchParams))
		}
	}

	// 路径参数和查询参数不能为 object
	for _, item := range p.Params {
		if item.Type.V() == TypeObject {
			pp.Error(pp.NewError(item.Start, item.End, "type", locale.ErrInvalidValue))
		}
	}
	for _, q := range p.Queries {
		if q.Type.V() == TypeObject {
			pp.Error(pp.NewError(q.Start, q.End, "type", locale.ErrInvalidValue))
		}
	}
}

func parsePath(path string) (params map[string]struct{}, err error) {
	start := -1
	for i, b := range path {
		switch b {
		case '{':
			if start != -1 {
				return nil, locale.NewError(locale.ErrInvalidFormat)
			}

			start = i + 1
		case '}':
			if start == -1 {
				return nil, locale.NewError(locale.ErrInvalidFormat)
			}

			if params == nil {
				params = make(map[string]struct{}, 3)
			}
			params[path[start:i]] = struct{}{}
			start = -1
		default:
		}
	}

	if start != -1 { // 没有结束符号
		return nil, locale.NewError(locale.ErrInvalidFormat)
	}

	return params, nil
}

// Sanitize token.Sanitizer
func (r *Request) Sanitize(p *xmlenc.Parser) {
	if r.Type.V() == TypeObject && len(r.Items) == 0 {
		p.Error(p.NewError(r.Start, r.End, "param", locale.ErrIsEmpty, "param"))
	}
	if r.Type.V() == TypeNone && len(r.Items) > 0 {
		p.Error(p.NewError(r.Start, r.End, "type", locale.ErrInvalidValue))
	}

	checkDuplicateEnum(r.Enums, p)

	if err := chkEnumsType(r.Type, r.Enums, p); err != nil {
		p.Error(err)
	}

	if err := checkXML(r.Array.V(), len(r.Items) > 0, &r.XML, p); err != nil {
		p.Error(err)
	}

	if r.Mimetype.V() != "" {
		for _, exp := range r.Examples {
			if exp.Mimetype.V() != r.Mimetype.V() {
				p.Error(p.NewError(r.Mimetype.Start, r.Mimetype.End, "mimetype", locale.ErrInvalidValue))
			}
		}
	}

	// 报头不能为 object
	for _, header := range r.Headers {
		if header.Type.V() == TypeObject {
			p.Error(p.NewError(header.Type.Start, header.Type.End, "type", locale.ErrInvalidValue))
		}
	}

	checkDuplicateItems(r.Items, p)
}

// Sanitize token.Sanitizer
func (p *Param) Sanitize(pp *xmlenc.Parser) {
	if p.Type.V() == TypeNone {
		pp.Error(pp.NewError(p.Start, p.End, "type", locale.ErrIsEmpty, "type"))
	}
	if p.Type.V() == TypeObject && len(p.Items) == 0 {
		pp.Error(pp.NewError(p.Start, p.End, "param", locale.ErrIsEmpty, "param"))
	}

	if p.Type.V() != TypeObject && len(p.Items) > 0 {
		pp.Error(pp.NewError(p.Type.Value.Start, p.Type.Value.End, "type", locale.ErrInvalidValue))
	}

	checkDuplicateEnum(p.Enums, pp)

	if err := chkEnumsType(p.Type, p.Enums, pp); err != nil {
		pp.Error(err)
	}

	checkDuplicateItems(p.Items, pp)

	if err := checkXML(p.Array.V(), len(p.Items) > 0, &p.XML, pp); err != nil {
		pp.Error(err)
	}

	if p.Summary.V() == "" && p.Description.V() == "" {
		pp.Error(pp.NewError(p.Start, p.End, "summary", locale.ErrIsEmpty, "summary"))
	}
}

// 检测 enums 中的类型是否符合 t 的标准，比如 Number 要求枚举值也都是数值
func chkEnumsType(t *TypeAttribute, enums []*Enum, p *xmlenc.Parser) error {
	if len(enums) == 0 {
		return nil
	}

	switch t.V() {
	case TypeNumber:
		for _, enum := range enums {
			if !is.Number(enum.Value.V()) {
				return p.NewError(enum.Start, enum.End, enum.StartTag.String(), locale.ErrInvalidFormat)
			}
		}
	case TypeBool:
		for _, enum := range enums {
			if _, err := strconv.ParseBool(enum.Value.V()); err != nil {
				return p.NewError(enum.Start, enum.End, enum.StartTag.String(), locale.ErrInvalidFormat)
			}
		}
	case TypeObject, TypeNone:
		return p.NewError(t.Start, t.End, t.AttributeName.String(), locale.ErrInvalidValue)
	}

	return nil
}

func checkDuplicateEnum(enums []*Enum, p *xmlenc.Parser) {
	indexes := sliceutil.Dup(enums, func(i, j int) bool { return enums[i].Value.V() == enums[j].Value.V() })
	if len(indexes) > 0 {
		err := p.NewError(enums[indexes[0]].Start, enums[indexes[0]].End, "enum", locale.ErrDuplicateValue)
		for _, i := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: enums[i].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		p.Error(err)
	}
}

func checkDuplicateItems(items []*Param, p *xmlenc.Parser) {
	indexes := sliceutil.Dup(items, func(i, j int) bool { return items[i].Name.V() == items[j].Name.V() })
	if len(indexes) > 0 {
		err := p.NewError(items[indexes[0]].Start, items[indexes[0]].End, "param", locale.ErrDuplicateValue)
		for _, i := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: items[i].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		p.Error(err)
	}
}

func checkXML(isArray, hasItems bool, xml *XML, p *xmlenc.Parser) error {
	if xml.XMLAttr.V() {
		if isArray || hasItems {
			return p.NewError(xml.XMLAttr.Start, xml.XMLAttr.End, xml.XMLAttr.AttributeName.String(), locale.ErrInvalidValue)
		}

		if xml.XMLWrapped.V() != "" {
			return p.NewError(xml.XMLWrapped.Start, xml.XMLWrapped.End, xml.XMLWrapped.AttributeName.String(), locale.ErrInvalidValue)
		}

		if xml.XMLExtract.V() {
			return p.NewError(xml.XMLExtract.Start, xml.XMLExtract.End, xml.XMLExtract.AttributeName.String(), locale.ErrInvalidValue)
		}

		if xml.XMLCData.V() {
			return p.NewError(xml.XMLCData.Start, xml.XMLCData.End, xml.XMLCData.AttributeName.String(), locale.ErrInvalidValue)
		}
	}

	if xml.XMLWrapped.V() != "" && !isArray {
		return p.NewError(xml.XMLWrapped.Start, xml.XMLWrapped.End, xml.XMLWrapped.AttributeName.String(), locale.ErrInvalidValue)
	}

	if xml.XMLExtract.V() {
		if xml.XMLNSPrefix.V() != "" {
			return p.NewError(xml.XMLNSPrefix.Start, xml.XMLNSPrefix.End, xml.XMLNSPrefix.AttributeName.String(), locale.ErrInvalidValue)
		}
	}

	return nil
}

// Sanitize 检测内容是否合法
func (doc *APIDoc) Sanitize(p *xmlenc.Parser) {
	if err := doc.checkXMLNamespaces(p); err != nil {
		p.Error(err)
	}
	doc.URI = p.Location.URI

	for _, api := range doc.APIs {
		if api.doc == nil {
			api.doc = doc // 保证单文件的文档能正常解析
			api.URI = doc.URI
		}
		api.sanitizeTags(p)
	}
}

// Sanitize 检测内容是否合法
func (ns *XMLNamespace) Sanitize(p *xmlenc.Parser) {
	if ns.URN.V() == "" {
		p.Error(p.NewError(ns.Start, ns.End, "@urn", locale.ErrIsEmpty, "@urn"))
	}
}

func (doc *APIDoc) checkXMLNamespaces(p *xmlenc.Parser) error {
	if len(doc.XMLNamespaces) == 0 {
		return nil
	}

	// 按 URN 查重
	indexes := sliceutil.Dup(doc.XMLNamespaces, func(i, j int) bool {
		return doc.XMLNamespaces[i].URN.V() == doc.XMLNamespaces[j].URN.V()
	})
	if len(indexes) > 0 {
		curr := doc.XMLNamespaces[indexes[0]].URN
		err := p.NewError(curr.Start, curr.End, "@urn", locale.ErrDuplicateValue)
		for _, i := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: doc.XMLNamespaces[i].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		return err
	}

	// 按 prefix 查重
	indexes = sliceutil.Dup(doc.XMLNamespaces, func(i, j int) bool {
		return doc.XMLNamespaces[i].Prefix.V() == doc.XMLNamespaces[j].Prefix.V()
	})
	if len(indexes) > 0 {
		curr := doc.XMLNamespaces[indexes[0]].URN
		err := p.NewError(curr.Start, curr.End, "@prefix", locale.ErrDuplicateValue)
		for _, i := range indexes[1:] {
			err.Relate(core.Location{URI: p.Location.URI, Range: doc.XMLNamespaces[i].Range}, locale.Sprintf(locale.ErrDuplicateValue))
		}
		return err
	}

	return nil
}

func (doc *APIDoc) findTag(tag string) *Tag {
	for _, t := range doc.Tags {
		if t.Name.V() == tag {
			return t
		}
	}
	return nil
}

func (doc *APIDoc) findServer(srv string) *Server {
	for _, s := range doc.Servers {
		if s.Name.V() == srv {
			return s
		}
	}
	return nil
}

func (api *API) sanitizeTags(p *xmlenc.Parser) {
	if api.doc == nil {
		panic("api.doc 未获取正确的值")
	}
	api.checkDup(p)

	apiURI := api.URI
	if apiURI == "" {
		apiURI = api.doc.URI
	}

	for _, tag := range api.Tags {
		t := api.doc.findTag(tag.Content.Value)
		if t == nil {
			loc := core.Location{
				URI: api.URI,
				Range: core.Range{
					Start: tag.Content.Start,
					End:   tag.Content.End,
				},
			}
			p.Warning(loc.NewError(locale.ErrInvalidValue).AddTypes(core.ErrorTypeUnused))
			continue
		}

		tag.definition = &Definition{
			Location: core.Location{
				Range: t.R(),
				URI:   api.doc.URI,
			},
			Target: t,
		}
		t.references = append(t.references, &Reference{
			Location: core.Location{
				Range: tag.R(),
				URI:   apiURI,
			},
			Target: tag,
		})
	}

	for _, srv := range api.Servers {
		s := api.doc.findServer(srv.Content.Value)
		if s == nil {
			loc := core.Location{
				URI: api.URI,
				Range: core.Range{
					Start: srv.Content.Start,
					End:   srv.Content.End,
				},
			}
			p.Warning(loc.NewError(locale.ErrInvalidValue).AddTypes(core.ErrorTypeUnused))
			continue
		}

		srv.definition = &Definition{
			Location: core.Location{
				Range: s.R(),
				URI:   api.doc.URI,
			},
			Target: s,
		}
		s.references = append(s.references, &Reference{
			Location: core.Location{
				Range: srv.R(),
				URI:   apiURI,
			},
			Target: srv,
		})
	}
}

// 检测当前 api 是否与 apidoc.APIs 中存在相同的值
func (api *API) checkDup(p *xmlenc.Parser) {
	err := (core.Location{URI: api.URI, Range: api.Range}).NewError(locale.ErrDuplicateValue)

	for _, item := range api.doc.APIs {
		if item == api {
			continue
		}

		if api.Method.V() != item.Method.V() {
			continue
		}

		p := ""
		if api.Path != nil {
			p = api.Path.Path.V()
		}
		iip := ""
		if item.Path != nil {
			iip = item.Path.Path.V()
		}
		if p != iip {
			continue
		}

		// 默认服务器
		if len(api.Servers) == 0 && len(item.Servers) == 0 {
			err.Relate(core.Location{URI: item.URI, Range: item.Range}, locale.Sprintf(locale.ErrDuplicateValue))
			continue
		}

		// 判断是否拥有相同的 server 字段
		for _, srv := range api.Servers {
			s := sliceutil.Count(item.Servers, func(i int) bool { return srv.V() == item.Servers[i].V() })
			if s > 0 {
				err.Relate(core.Location{URI: item.URI, Range: item.Range}, locale.Sprintf(locale.ErrDuplicateValue))
				continue
			}
		}
	}

	if len(err.Related) > 0 {
		p.Error(err)
	}
}
