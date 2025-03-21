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
	// контейнер данных для запроса
	data := url.Values{}
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Ошибка чтения URL: %v", err)
	}
	long = strings.TrimSuffix(long, "\n")

	// Валидация URL
	if long == "" {
		log.Fatalf("URL не может быть пустым")
	}

	_, err = url.ParseRequestURI(long)
	if err != nil {
		log.Fatalf("Некорректный URL: %v", err)
	}

	// заполняем контейнер данными
	data.Set("url", long)
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Ошибка отправки запроса: %v", err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}
	// и печатаем его
	fmt.Println(string(body))
}
