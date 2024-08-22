package services

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Todo struct {
	ID    primitive.ObjectID `bson:"_id"`
	Title string
	Done  bool
}

type TodoService struct {
	Collection *mongo.Collection
}

func (s *TodoService) AllTodos() ([]Todo, error) {
	cursor, err := s.Collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	var results []Todo
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *TodoService) GetTodo(id interface{}) (Todo, error) {
	var result Todo
	err := s.Collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		return Todo{}, err
	}
	return result, nil
}

func (s *TodoService) AddTodo(t Todo) (interface{}, error) {
	document := bson.D{{Key: "title", Value: t.Title}, {Key: "done", Value: false}}
	result, err := s.Collection.InsertOne(context.TODO(), document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (s *TodoService) UpdateTodo(t Todo) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: t.ID}}
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "title", Value: t.title}, {Key: "done", Value: t.done}}}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "done", Value: t.Done}}}}
	result, err := s.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}
	return result.UpsertedID, nil
}

func (s *TodoService) DeleteTodo(id interface{}) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	result, err := s.Collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}
