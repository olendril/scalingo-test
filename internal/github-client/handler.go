package github_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v66/github"
	"github.com/olendril/scalingo-test/api"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type Server struct {
	client *github.Client
	logger logrus.FieldLogger
}

type Language map[string]struct {
	Bytes *int
}

func NewServer(log logrus.FieldLogger, accessToken string) *Server {

	var client *github.Client
	if accessToken == "" {
		client = github.NewClient(nil)

	} else {
		client = github.NewClient(nil).WithAuthToken(accessToken)
	}

	if client == nil {
		log.Error("Failed to create github client")
		return nil
	}

	return &Server{
		client: client,
		logger: log,
	}
}

func (Server) GetPing(w http.ResponseWriter, r *http.Request) {
	resp := api.Pong{
		Ping: "pong",
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s Server) GetRepos(w http.ResponseWriter, r *http.Request, params api.GetReposParams) {

	query := queryBuilder(params)

	s.logger.WithFields(logrus.Fields{
		"query": query,
	}).Info()

	repos, rsp, err := s.client.Search.Repositories(context.Background(), query, &github.SearchOptions{
		Sort:  "created",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	if err != nil {
		if rsp.StatusCode == http.StatusUnprocessableEntity {
			s.logger.WithError(err).Error("Failed to get repos")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else {
			s.logger.WithError(err).Error("Failed to get repos")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	s.logger.WithFields(logrus.Fields{
		"resp": rsp,
		"err":  err,
	}).Info()

	repoChan := make(chan api.Repo, 100)

	var wg sync.WaitGroup

	// Make request for each repository to get language details using goroutine
	for _, repo := range repos.Repositories {
		wg.Add(1)
		go func(owner string, repoName string, repoFullName string, language string) {
			defer wg.Done()

			languages, rsp, err := s.client.Repositories.ListLanguages(context.Background(), owner, repoName)

			if err != nil {
				s.logger.WithError(err).Errorf("Failed to list language for repo %s", repoName)
				return
			}

			if rsp.StatusCode != 200 {
				s.logger.WithFields(logrus.Fields{
					"rsp": rsp,
				}).Error("Failed to list language for repo %s", repoName)
				return
			}

			repoResp := api.Repo{
				FullName:   repoFullName,
				Owner:      owner,
				Repository: repoName,
				Languages: map[string]struct {
					Bytes int `json:"bytes"`
				}{
					language: {
						Bytes: languages[language],
					},
				},
			}

			repoChan <- repoResp
			return
		}(*repo.Owner.Login, *repo.Name, *repo.FullName, *repo.Language)
	}

	wg.Wait()
	close(repoChan)

	response := []api.Repo{}

	for i := 0; i < len(repos.Repositories); i++ {
		tmp := <-repoChan
		if tmp.FullName != "" {
			response = append(response, tmp)
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func queryBuilder(params api.GetReposParams) string {
	query := `is:public`

	if params.Language != nil {
		query += fmt.Sprintf(" language:%s", *params.Language)
	}

	if params.License != nil {
		query += fmt.Sprintf(" license:%s", *params.License)
	}

	return query
}
