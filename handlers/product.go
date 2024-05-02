package handlers

import (
	"context"
	"time"

	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostCourse(read model.Course) (model.Course, error) {
	toInsert := model.Course{}
	toInsert.Id = primitive.NewObjectID()
	toInsert.Date = primitive.NewDateTimeFromTime(time.Now())
	toInsert.Name = read.Name
	toInsert.Description = read.Description
	toInsert.Link = read.Link
	toInsert.Image = read.Image

	_, err := db.InsertOne(
		db.Instance.Client,
		db.Instance.Context,
		db.Instance.Dbname,
		db.CourseCollection,
		toInsert,
	)

	return toInsert, err
}

func GetCourseByID(courseID primitive.ObjectID) (*model.Course, error) {
	return getCourseByID(courseID)
}

func getCourseByID(id primitive.ObjectID) (*model.Course, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).
		Collection(db.CourseCollection)

	var result model.Course
	err := collection.FindOne(db.Instance.Context, bson.M{"_id": id}).
		Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func readAllCourse(
	client *mongo.Client,
	ctx context.Context,
	dataBase, col string,
) (*mongo.Cursor, error) {
	collection := client.Database(dataBase).Collection(col)

	cur, err := collection.Find(ctx, bson.M{})
	return cur, err
}

func GetAllCourses() ([]model.Course, error) {
	cur, err := readAllCourse(
		db.Instance.Client,
		db.Instance.Context,
		db.Instance.Dbname,
		db.ProductCollection,
	)
	if err != nil {
		return nil, err
	}

	var results []model.Course
	for cur.Next(db.Instance.Context) {
		var elem model.Course
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func updateCourseByID(id primitive.ObjectID, updateData bson.M) error {
	collection := db.Instance.Client.Database(db.Instance.Dbname).
		Collection(db.ProductCollection)
	_, err := collection.UpdateOne(
		db.Instance.Context,
		bson.M{"_id": id},
		bson.M{"$set": updateData},
	)
	return err
}

func UpdateCourseValue(
	courseID primitive.ObjectID,
	name string,
	description string,
	link string,
	image string,
) error {

	updateData := bson.M{
		"Name":        name,
		"Description": description,
		"Link":       link,
	}
	return updateCourseByID(courseID, updateData)
}

func deleteCourseByID(id primitive.ObjectID) error {
	collection := db.Instance.Client.Database(db.Instance.Dbname).
		Collection(db.CourseCollection)
	_, err := collection.DeleteOne(db.Instance.Context, bson.M{"_id": id})
	return err
}

func DeleteCourse(courseID primitive.ObjectID) error {
	return deleteCourseByID(courseID)
}
