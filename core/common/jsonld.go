package common

import (
	"encoding/json"
	"fmt"
)

// JSONLDer defines an interface for any object that can be represented as JSON-LD.
type JSONLDer interface {
	LDType() string
	MarshalJSONLD() ([]byte, error)
}

// LDContext is the default schema.org context.
const LDContext = "https://schema.org"

// unmarshalHelper assists in extracting the @type for unmarshaling interfaces.
type unmarshalHelper struct {
	Type string `json:"@type"`
}

// Person represents a Person in JSON-LD.
type Person struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (p Person) LDType() string { return "Person" }
func (p Person) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(p, p.LDType())
}

// Organization represents an Organization in JSON-LD.
type Organization struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (o Organization) LDType() string { return "Organization" }
func (o Organization) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(o, o.LDType())
}

// Article represents an Article in JSON-LD.
type Article struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        JSONLDer `json:"author,omitempty"`
}

func (a Article) LDType() string { return "Article" }
func (a Article) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(a, a.LDType())
}

// NewsArticle represents a NewsArticle in JSON-LD.
type NewsArticle struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        JSONLDer `json:"author,omitempty"`
}

func (a NewsArticle) LDType() string { return "NewsArticle" }
func (a NewsArticle) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(a, a.LDType())
}

// BlogPosting represents a BlogPosting in JSON-LD.
type BlogPosting struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        JSONLDer `json:"author,omitempty"`
}

func (a BlogPosting) LDType() string { return "BlogPosting" }
func (a BlogPosting) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(a, a.LDType())
}

// DiscussionForumPosting represents a DiscussionForumPosting in JSON-LD.
type DiscussionForumPosting struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        JSONLDer `json:"author,omitempty"`
}

func (a DiscussionForumPosting) LDType() string { return "DiscussionForumPosting" }
func (a DiscussionForumPosting) MarshalJSONLD() ([]byte, error) {
	return marshalWithContextAndType(a, a.LDType())
}

// marshalWithContextAndType is an internal helper to inject @context and @type.
func marshalWithContextAndType(v interface{}, ldType string) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	m["@context"] = LDContext
	m["@type"] = ldType
	return json.Marshal(m)
}

// UnmarshalJSONLD parses bytes into a JSONLDer.
func UnmarshalJSONLD(data []byte) (JSONLDer, error) {
	var h unmarshalHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}

	var res JSONLDer
	switch h.Type {
	case "Article":
		var a Article
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		res = a
	case "NewsArticle":
		var a NewsArticle
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		res = a
	case "BlogPosting":
		var a BlogPosting
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		res = a
	case "DiscussionForumPosting":
		var a DiscussionForumPosting
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		res = a
	case "Person":
		var p Person
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		res = p
	case "Organization":
		var o Organization
		if err := json.Unmarshal(data, &o); err != nil {
			return nil, err
		}
		res = o
	default:
		return nil, fmt.Errorf("unknown @type: %s", h.Type)
	}

	return res, nil
}
