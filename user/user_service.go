package user

import (
	"context"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createUser(user *User) error {
	return mgm.Coll(user).Create(user)
}

func getUserByID(id string) (*User, error) {
	user := &User{}
	err := mgm.Coll(user).FindByID(id, user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return user, err
}

func updateUser(id string, updates map[string]interface{}) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidIdFormat
	}

	updates["updated_at"] = time.Now().UTC().Truncate(time.Second)
	result, err := mgm.Coll(&User{}).UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, ErrUserNotFound
	}

	return getUserByID(id)
}

func replaceUser(user *User) error {
	return mgm.Coll(user).Update(user)
}

func deleteUser(user *User) error {
	return mgm.Coll(user).Delete(user)
}
