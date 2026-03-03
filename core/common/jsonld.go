package common

import (
	"encoding/json"
)

// JSONLD is a concrete struct for LD-JSON data.
type JSONLD struct {
	Context string
	Type    string
	Data    interface{}
}

// MarshalJSON marshals the JSONLD object, injecting @context and @type.
func (j JSONLD) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(j.Data)
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
	if j.Context != "" {
		m["@context"] = j.Context
	} else {
		m["@context"] = "https://schema.org"
	}
	if j.Type != "" {
		m["@type"] = j.Type
	}
	return json.Marshal(m)
}

// UnmarshalJSON unmarshals the JSONLD object, switching on @type to instantiate the correct Data struct.
func (j *JSONLD) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if ctx, ok := raw["@context"]; ok {
		var s string
		if err := json.Unmarshal(ctx, &s); err == nil {
			j.Context = s
		}
	}
	if typ, ok := raw["@type"]; ok {
		var s string
		if err := json.Unmarshal(typ, &s); err == nil {
			j.Type = s
		}
	}

	switch j.Type {
	case "Article":
		var a Article
		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}
		j.Data = a
	case "NewsArticle":
		var a NewsArticle
		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}
		j.Data = a
	case "BlogPosting":
		var a BlogPosting
		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}
		j.Data = a
	case "DiscussionForumPosting":
		var a DiscussionForumPosting
		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}
		j.Data = a
	case "Person":
		var p Person
		if err := json.Unmarshal(data, &p); err != nil {
			return err
		}
		j.Data = p
	case "Organization":
		var o Organization
		if err := json.Unmarshal(data, &o); err != nil {
			return err
		}
		j.Data = o
	default:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		delete(m, "@context")
		delete(m, "@type")
		j.Data = m
	}
	return nil
}

type Article struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        *Person  `json:"author,omitempty"`
}

type NewsArticle struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        *Person  `json:"author,omitempty"`
}

type BlogPosting struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        *Person  `json:"author,omitempty"`
}

type DiscussionForumPosting struct {
	Headline      string   `json:"headline,omitempty"`
	Image         []string `json:"image,omitempty"`
	DatePublished string   `json:"datePublished,omitempty"`
	DateModified  string   `json:"dateModified,omitempty"`
	Author        *Person  `json:"author,omitempty"`
}

type Person struct {
	Type string `json:"@type,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (p Person) MarshalJSON() ([]byte, error) {
	type Alias Person
	a := Alias(p)
	if a.Type == "" {
		a.Type = "Person"
	}
	return json.Marshal(a)
}

type Organization struct {
	Type string `json:"@type,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

func (o Organization) MarshalJSON() ([]byte, error) {
	type Alias Organization
	a := Alias(o)
	if a.Type == "" {
		a.Type = "Organization"
	}
	return json.Marshal(a)
}
