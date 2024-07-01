package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/joho/godotenv"
)

type GitQuery struct {
	Query string `json:"query"`
}

type GitResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Node struct {
					Login string `json:"login"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"search"`
	} `json:"data"`
}

var NUM_USERS = 200

func main() {

	now := time.Now()
	defer func(now time.Time) {
		fmt.Println("Total Time Take: ", time.Since(now).Seconds())
	}(now)

	if err := godotenv.Load(); err != nil {
		log.Panic(err)
	}

	// LOAD ENV VARIABLES
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		log.Panic("Error loading the environment variables")
	}

	userList, err := getUsernames(token)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("User list size: ", len(userList.Data.Search.Edges))
	fmt.Println("Start Fetching ... ")

	wg := sync.WaitGroup{}
	wg.Add(len(userList.Data.Search.Edges))

	for _, user := range userList.Data.Search.Edges {
		go func(username string) {
			defer wg.Done()
			m := GetMaroto(username)
			document, err := m.Generate()
			if err != nil {
				log.Println(err)
			}

			err = document.Save(fmt.Sprintf("./assets/%s.pdf", username))
			if err != nil {
				log.Println(err)
			}
		}(user.Node.Login)
	}
	wg.Wait()

	fmt.Println("Process Complete: RESUMES GENERATED")
}

func getUsernames(token string) (*GitResponse, error) {

	query := fmt.Sprintf(`{
				search(query: "type:user", type: USER, first: %d) {
					edges {
						node {
							... on User {
								login
							}
						}
					}
				}
			}`, NUM_USERS)

	gqlQuery := GitQuery{
		Query: query,
	}

	reqBody, err := json.Marshal(gqlQuery)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	userList := GitResponse{}
	if err := json.Unmarshal(responseBody, &userList); err != nil {
		return nil, err
	}

	return &userList, nil
}

func GetMaroto(username string) core.Maroto {

	m := maroto.New()

	m.AddRow(20, text.NewCol(4, username, props.Text{
		Style:  "B",
		Size:   20.0,
		Family: "helvetica",
	}))

	m.AddRow(30, text.NewCol(12, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."))
	m.AddRow(10, text.NewCol(4, fmt.Sprintf("https://www.github.com/%s", username)))

	return m
}
