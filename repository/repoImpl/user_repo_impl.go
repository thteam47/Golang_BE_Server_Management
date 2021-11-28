package repoimpl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/thteam47/server_management/drive"
	"github.com/thteam47/server_management/global"
	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)


type UserRepositoryImpl struct {
	DB           *mongo.Database
	MyRediscache *cache.Cache
	jwtManager   repo.JwtRepository
}

func NewUserRepo(mongodb *mongo.Database, redis *cache.Cache, jwtManager repo.JwtRepository) repo.UserRepository {
	return &UserRepositoryImpl{
		DB:           mongodb,
		MyRediscache: redis,
		jwtManager:   jwtManager,
	}
}
func (u *UserRepositoryImpl) Login(username string, password string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	err := u.DB.Collection(vi.GetString("collectionUser")).FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return false, "", "", errors.New("User not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, "", "", errors.New("Incorrect password")
	}
	token, err := u.jwtManager.Generate(&user)
	if err != nil {
		return false, "", "", errors.New("Could not login")
	}
	if err := u.MyRediscache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   user.ID.Hex(),
		Value: token,
		TTL:   30 * time.Minute,
	}); err != nil {
		panic(err)
	}
	return true, token, user.Role, nil
}
func (u *UserRepositoryImpl) Logout(ctx context.Context) (bool, error) {
	idUser, err := u.GetIdUser(ctx)

	err = RemoveValueCache(u.MyRediscache, idUser)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (u *UserRepositoryImpl) GetIdUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) < 1 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := strings.TrimPrefix(values[0], "Bearer ")
	if accessToken == "undefined" {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	claims, err := u.jwtManager.Verify(accessToken)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	return claims.Issuer, nil
}
func (u *UserRepositoryImpl) GetListUser() ([]*models.User, error) {
	collection := u.DB.Collection(vi.GetString("collectionUser"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var ltUser []*models.User
	for cur.Next(context.TODO()) {
		var elem models.User
		er := cur.Decode(&elem)
		if er != nil {
			log.Fatal(err)
		}
		userResp := &models.User{
			ID:       elem.ID,
			Username: elem.Username,
			Password: elem.Password,
			Email:    elem.Email,
			Role:     elem.Role,
			FullName: elem.FullName,
			Action:   elem.Action,
		}
		ltUser = append(ltUser, userResp)
	}
	return ltUser, nil
}
func (u *UserRepositoryImpl) GetUser(idUser string) (*models.User, error) {
	var user models.User
	id, _ := primitive.ObjectIDFromHex(idUser)
	collection := u.DB.Collection(vi.GetString("collectionUser"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("Id User not found")
	}
	return &models.User{
		ID:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		Action:   user.Action,
	}, nil

}
func (u *UserRepositoryImpl) AddUser(user *models.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	insertResult, err := u.DB.Collection(vi.GetString("collectionUser")).InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("%v", insertResult.InsertedID)
	idResp := strings.Split(str, "\"")
	return idResp[1], nil
}
func (u *UserRepositoryImpl) ChangeActionUser(idUser string, role string, a []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var roleUser string
	if role == "" {
		roleUser = "staff"
	} else {
		roleUser = role
	}
	var actionList []string
	if roleUser == "admin" {
		actionList = append(actionList, "All Rights")
	} else if roleUser == "assistant" {
		actionList = []string{"Add Server", "Update Server", "Detail Status", "Export", "Connect", "Disconnect", "Delete Server", "Change Password"}
	} else {
		actionList = a
	}
	id, _ := primitive.ObjectIDFromHex(idUser)
	filterUser := bson.M{"_id": id}
	updateUser := bson.M{"$set": bson.M{
		"role":   roleUser,
		"action": actionList,
	}}
	_, err := global.DB.Collection(vi.GetString("collectionUser")).UpdateOne(ctx, filterUser, updateUser)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepositoryImpl) UpdateUser(idUser string, user *models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id primitive.ObjectID
	id, _ = primitive.ObjectIDFromHex(idUser)

	filterUser := bson.M{"_id": id}
	updateUser := bson.M{"$set": bson.M{
		"fullname": user.FullName,
		"username": user.Username,
		"email":    user.Email,
	}}
	_, err := u.DB.Collection(vi.GetString("collectionUser")).UpdateOne(ctx, filterUser, updateUser)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:       id,
		Username: user.FullName,
		FullName: user.FullName,
		Email:    user.Email,
	}, nil
}
func (u *UserRepositoryImpl) ChangePassUser(idUser string, pass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id primitive.ObjectID
	id, _ = primitive.ObjectIDFromHex(idUser)

	passHash, _ := drive.HashPassword(pass)
	filterUser := bson.M{"_id": id}
	updateUser := bson.M{"$set": bson.M{
		"password": passHash,
	}}
	_, err := u.DB.Collection(vi.GetString("collectionUser")).UpdateOne(ctx, filterUser, updateUser)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserRepositoryImpl) DeleteUser(idUser string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id, _ := primitive.ObjectIDFromHex(idUser)
	result, err := u.DB.Collection(vi.GetString("collectionUser")).DeleteOne(ctx, bson.M{"_id": id})

	if err != nil {
		log.Fatal(err)
	}
	if result.DeletedCount == 0 {
		return errors.New("Id incorrect")
	}
	return nil
}
