package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-redis/internal/pkg/cache"
	"go-redis/internal/pkg/db"
	"go-redis/internal/pkg/repository"
	"go-redis/internal/pkg/repository/postgresql"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type serverUserData struct {
	Id   int64
	Name string
}

type UserServer struct {
	userRepo repository.UsersRepo
}

var ctx = context.Background()

func main() {

	_, database, err := db.NewDB(ctx)
	defer database.GetPool(ctx).Close()

	userRepo := postgresql.NewUsers(database)
	cache.ConnectRedisCache()

	if err != nil {
		return
	}

	mux := CreateUserServer(ctx, userRepo)

	if err := http.ListenAndServe(":9000", mux); err != nil {
		log.Fatal(err)
	}
}

func (s *UserServer) getUser(cxt context.Context, req *http.Request) ([]byte, int) {
	id, err := getUserID(req.URL)
	if err != nil {
		fmt.Errorf("can't parse id: %s", err)
		return nil, http.StatusBadRequest
	}

	res := cache.GetFromCache(cache.REDIS, id)

	if res != nil {
		data, err := json.Marshal(res)
		if err != nil {
			fmt.Errorf("can't marshal user with id: %d. Error: %s", id, err)
			return nil, http.StatusInternalServerError
		}
		return data, http.StatusOK
	}

	var user *repository.User

	user, err = s.userRepo.GetById(cxt, int64(id))
	if err != nil {
		fmt.Errorf("can't parse id: %s", err)
		return nil, http.StatusInternalServerError

	}

	su := &serverUserData{}
	su.Id = user.Id
	su.Name = user.Name

	data, err := json.Marshal(su)
	if err != nil {
		fmt.Errorf("can't marshal user with id: %d. Error: %s", id, err)
		return nil, http.StatusInternalServerError
	}

	return data, http.StatusOK
}

func (s *UserServer) createUser(cxt context.Context, req *http.Request) (uint, int) {
	user, err := getUserData(req.Body)
	if err != nil {
		fmt.Println(err)
		return 0, http.StatusBadRequest
	}
	id, err := s.userRepo.Add(cxt, user.Name)
	if err != nil {
		fmt.Println(err)
		return 0, http.StatusInternalServerError
	}

	return uint(id), http.StatusOK
}

func (s *UserServer) updateUser(ctx context.Context, req *http.Request) int {
	user, err := getUserData(req.Body)
	if err != nil {
		fmt.Println(err)
		return http.StatusBadRequest
	}

	u := &repository.User{Id: user.Id, Name: user.Name}

	updated, err := s.userRepo.Update(ctx, u)

	cache.SetInCache(cache.REDIS, user.Id, user)

	if err != nil {
		fmt.Println(err)
		return http.StatusBadRequest
	}

	if !updated {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (s *UserServer) deleteUser(ctx context.Context, req *http.Request) int {
	id, err := getUserID(req.URL)
	if err != nil {
		fmt.Println(err)
		return http.StatusBadRequest
	}

	cache.DeleteFromCache(cache.REDIS, id)

	e := s.userRepo.Delete(ctx, int64(id))
	if e != nil {
		fmt.Println(err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// CreateUserServer Эта функция для создания хендла пользователей
func CreateUserServer(ctx context.Context, ur repository.UsersRepo) *http.ServeMux {
	serv := UserServer{
		userRepo: ur,
	}
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/user", func(res http.ResponseWriter, req *http.Request) {
		if req == nil {
			return
		}

		switch req.Method {
		case http.MethodGet:
			data, status := serv.getUser(ctx, req)
			res.WriteHeader(status)
			res.Write(data)
		case http.MethodPost:
			_, status := serv.createUser(ctx, req)
			res.WriteHeader(status)
		case http.MethodDelete:
			status := serv.deleteUser(ctx, req)
			res.WriteHeader(status)
		case http.MethodPut:
			status := serv.updateUser(ctx, req)
			res.WriteHeader(status)

		default:
			fmt.Printf("unsupported method: [%s]", req.Method)
			res.WriteHeader(http.StatusNotImplemented)
		}
	})

	return serveMux
}

func getUserData(reader io.ReadCloser) (serverUserData, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return serverUserData{}, err
	}

	data := serverUserData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func getUserID(reqUrl *url.URL) (int64, error) {
	idStr := reqUrl.Query().Get("id")
	if len(idStr) == 0 {
		return 0, errors.New("can't get id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("can't parse id: %s", err)
	}

	return int64(id), nil
}
