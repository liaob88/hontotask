package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Book struct {
	title   string
	chapter int
	section int
}

type TodoistBoard struct {
	projectId string
	sectionId string
}

func main() {
	fmt.Print("---タスク作成開始---\n")

	title, chaptersCount, sectionsByChapter := readBookInfo()
	projectId, sectionId := readTodoistInfo()
	start := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(chaptersCount)

	// MEMO: 章ごとに節数分のタスクを作成する
	for chapter := 1; chapter <= chaptersCount; chapter++ {
		go func(chapter int) {
			defer wg.Done()

			sectionsCount := sectionsByChapter[chapter-1]

			wg2 := sync.WaitGroup{}
			wg2.Add(sectionsCount)
			// MEMO: 節数分のタスクを作成する
			for section := 1; section <= sectionsCount; section++ {
				go createTaskFromBook(&wg2, &Book{title, chapter, section}, &TodoistBoard{projectId, sectionId})
			}
			wg2.Wait()
		}(chapter)
	}
	wg.Wait()

	end := time.Now()
	fmt.Print("---タスク作成完了---")
	fmt.Printf("---所要時間: %f秒---", (end.Sub(start)).Seconds())
}

func readBookInfo() (title string, chaptersCount int, sectionsByChapter []int) {
	title, _ = readTitle()
	chaptersCount, _ = readChaptersCount()

	for i := 1; i <= chaptersCount; i++ {
		sectionsCount, _ := readSectionsCount(i)
		sectionsByChapter = append(sectionsByChapter, sectionsCount)
	}
	return
}

func readTodoistInfo() (projectId string, sectionId string) {
	fmt.Print("Todoist の project id を入力してください: ")
	projectId = strings.TrimSuffix(readInput(), "\n")

	fmt.Print("Todoist の section id を入力してください: ")
	sectionId = strings.TrimSuffix(readInput(), "\n")

	return
}

func readTitle() (title string, err error) {
	fmt.Print("Enter book title: ")
	title = strings.TrimSuffix(readInput(), "\n")
	return
}

func readSectionsCount(chapter int) (sectionsCount int, err error) {
	fmt.Printf("Enter chapter%d's sections count: ", chapter)
	sectionsCount, _ = strconv.Atoi(strings.TrimSuffix(readInput(), "\n"))
	return
}

func readChaptersCount() (chaptersCount int, err error) {
	fmt.Print("Enter ChaptersCount: ")
	chaptersCount, _ = strconv.Atoi(strings.TrimSuffix(readInput(), "\n"))
	return
}

func readInput() (input string) {
	reader := bufio.NewReader(os.Stdin)
	// ReadString will block until the delimiter is entered
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	return input
}

func createTaskFromBook(wg *sync.WaitGroup, book *Book, todoistBoard *TodoistBoard) {
	defer wg.Done()

	values := url.Values{
		"content":    []string{fmt.Sprintf("[%s] %d_%d", book.title, book.chapter, book.section)},
		"project_id": []string{todoistBoard.projectId},
		"section_id": []string{todoistBoard.sectionId},
	}

	postCreateTask(values)
}

func postCreateTask(values url.Values) {
	endpoint := "https://api.todoist.com/rest/v1/tasks"

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TODOIST_API_TOKEN")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteArray))
}
