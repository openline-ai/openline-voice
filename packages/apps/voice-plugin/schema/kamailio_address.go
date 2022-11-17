package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User schema.
type KamailioAddress struct {
	ent.Schema
}

// Annotations of the User.
func (KamailioAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kamailio_address"},
	}
}

// Fields of the user.
func (KamailioAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Int("grp").Default(1),
		field.String("ip_addr").
			Unique(),
		field.Int8("mask"),
		field.Int16("port"),
		field.String("tag"),
	}
}
