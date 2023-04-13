package repository

import (
	"context"
	"fmt"
	notemodel "github.com/bruma1994/dyngo/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"os/user"
	"time"
)

func NewNotesRepository(collection *mongo.Collection, counter int) NotesRepository {
	return &NotesRepositoryImpl{
		collection: collection,
		counter:    counter,
	}
}

type NotesRepositoryImpl struct {
	collection *mongo.Collection
	counter    int
}

func (r *NotesRepositoryImpl) Add(ctx context.Context, note notemodel.Note) error {
	author, _ := user.Current()

	_, err := r.collection.InsertOne(ctx, bson.D{
		{"name", note.Name},
		{"text", note.Text},
		{"tag", note.Tag},
		{"author", author.Username},
		{"created_at", time.Now()},
		{"id", r.counter},
	})

	if err != nil {
		return fmt.Errorf("add note: %w", err)
	}
	r.counter++

	return nil
}

func (r *NotesRepositoryImpl) Delete(ctx context.Context, id int32) error {
	filter := bson.D{primitive.E{Key: "id", Value: id}}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	return nil
}

func (r *NotesRepositoryImpl) ByAuthor(ctx context.Context, author string) (notemodel.Note, error) {
	note := notemodel.Note{}

	filter := bson.D{primitive.E{Key: "author", Value: author}}
	result := r.collection.FindOne(ctx, filter)

	err := result.Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return notemodel.Note{}, fmt.Errorf("no documents by given filter")
		}
		return notemodel.Note{}, fmt.Errorf("show note by author: %w", err)
	}

	return note, nil
}

func (r *NotesRepositoryImpl) ByName(ctx context.Context, name string) (notemodel.Note, error) {
	note := notemodel.Note{}

	filter := bson.D{primitive.E{Key: "name", Value: name}}
	result := r.collection.FindOne(ctx, filter)

	err := result.Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return notemodel.Note{}, fmt.Errorf("no documents by given filter")
		}
		return notemodel.Note{}, fmt.Errorf("show note by author: %w", err)
	}

	return note, nil
}

func (r *NotesRepositoryImpl) All(ctx context.Context) ([]notemodel.Note, error) {
	var notes []notemodel.Note

	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("get all notes: %w", err)
	}

	err = cursor.All(ctx, &notes)
	if err != nil {
		fmt.Errorf("decode to slice: %w", err)
	}

	return notes, nil
}

type NotesRepository interface {
	Add(context.Context, notemodel.Note) error
	Delete(context.Context, int32) error
	ByAuthor(context.Context, string) (notemodel.Note, error)
	ByName(context.Context, string) (notemodel.Note, error)
	All(context.Context) ([]notemodel.Note, error)
}
