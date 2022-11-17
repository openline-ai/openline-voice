package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User schema.
type OpenlineCarrier struct {
	ent.Schema
}

// Annotations of the User.
func (OpenlineCarrier) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_carrier"},
	}
}

// Fields of the user.
func (OpenlineCarrier) Fields() []ent.Field {
	return []ent.Field{
		field.String("carrier_name"),
		field.String("username").
			Unique(),
		field.String("realm"),
		field.String("ha1"),
		field.String("domain"),
	}
}
