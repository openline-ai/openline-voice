package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User schema.
type OpenlineNumberMapping struct {
	ent.Schema
}

// Annotations of the User.
func (OpenlineNumberMapping) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_number_mapping"},
	}
}

// Fields of the user.
func (OpenlineNumberMapping) Fields() []ent.Field {
	return []ent.Field{
		field.String("e164").
			Unique(),
		field.String("alias"),
		field.String("sipuri").
			Unique(),
		field.String("carrier_name"),
	}
}
