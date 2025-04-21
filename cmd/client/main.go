package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"
	// data container for the request
	data := url.Values{}

	fmt.Println("Введите длинный URL")

	reader := bufio.NewReader(os.Stdin)

	long, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Ошибка чтения URL: %v", err)
	}
	long = strings.TrimSuffix(long, "\n")

	// URL validation
	if long == "" {
		log.Fatalf("URL не может быть пустым")
	}

	_, err = url.ParseRequestURI(long)
	if err != nil {
		log.Fatalf("Некорректный URL: %v", err)
	}

	data.Set("url", long)

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Ошибка отправки запроса: %v", err)
	}

	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}

	fmt.Println(string(body))
}
