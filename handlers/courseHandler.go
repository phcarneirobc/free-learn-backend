// handlers/handlers.go

package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SearchCourses(query string) ([]model.Course, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var courses []model.Course
	if err = cursor.All(context.Background(), &courses); err != nil {
		return nil, err
	}
	return courses, nil
}

func PostCourse(read model.Course, creatorID primitive.ObjectID) (model.Course, error) {
	toInsert := model.Course{
		Id:          primitive.NewObjectID(),
		Date:        primitive.NewDateTimeFromTime(time.Now()),
		Name:        read.Name,
		Description: read.Description,
		Image:       read.Image,
		Link:        read.Link,
		Modules:     read.Modules,
		CreatorID:   creatorID,
	}

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
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var result model.Course
	err := collection.FindOne(db.Instance.Context, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func GetAllCourses() ([]model.Course, error) {
	cur, err := readAllCourse(db.Instance.Client, db.Instance.Context, db.Instance.Dbname, db.CourseCollection)
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

func readAllCourse(client *mongo.Client, ctx context.Context, dataBase, col string) (*mongo.Cursor, error) {
	collection := client.Database(dataBase).Collection(col)
	cur, err := collection.Find(ctx, bson.M{})
	return cur, err
}

func UpdateCourseValue(courseID primitive.ObjectID, name, description, link, image string, modules []model.Module) error {
	updateData := bson.M{
		"name":        name,
		"description": description,
		"link":        link,
		"image":       image,
		"modules":     modules,
	}
	return updateCourseByID(courseID, updateData)
}

func updateCourseByID(id primitive.ObjectID, updateData bson.M) error {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	_, err := collection.UpdateOne(db.Instance.Context, bson.M{"_id": id}, bson.M{"$set": updateData})
	return err
}

func DeleteCourse(userID, courseID primitive.ObjectID) error {
	courseCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var course model.Course

	err := courseCollection.FindOne(context.Background(), bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("course not found")
		}
		return err
	}

	if course.CreatorID != userID {
		return errors.New("user is not the creator of the course")
	}

	_, err = courseCollection.DeleteOne(context.Background(), bson.M{"_id": courseID})
	return err
}

func RateCourse(userID, courseID primitive.ObjectID, rating model.Rating) error {
	courseCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var course model.Course

	err := courseCollection.FindOne(context.Background(), bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		return err
	}

	if course.CreatorID == userID {
		return errors.New("professors cannot rate their own courses")
	}

	for _, existingRating := range course.Ratings {
		if existingRating.UserID == userID {
			return errors.New("user has already rated this course")
		}
	}

	course.Ratings = append(course.Ratings, rating)
	_, err = courseCollection.UpdateOne(context.Background(), bson.M{"_id": courseID}, bson.M{"$set": bson.M{"ratings": course.Ratings}})
	return err
}
