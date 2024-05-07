package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Course struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Date     primitive.DateTime `json:"date"          bson:"date"`
	Name     string             `json:"name"          bson:"name"`
	Description  string             `json:"description"      bson:"description"`
	Image	 string             `json:"image"         bson:"image"`
	Link	 string             `json:"link"          bson:"link"`
}

type User struct {
    Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Date        primitive.DateTime `json:"date"          bson:"date"`
    Email       string             `json:"email"          bson:"email"`
    Professor   bool               `json:"professor"     bson:"professor"`
    Password    string             `json:"password"      bson:"password"`
    Cursos      []primitive.ObjectID `json:"cursos"        bson:"cursos"` // Corrigido para []primitive.ObjectID
}
