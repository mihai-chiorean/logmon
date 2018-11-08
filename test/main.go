package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	f, err := os.OpenFile("../access.log", os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	methods := []string{
		"GET",
		"POST",
		"HEAD",
		"OPTIONS",
		"PUT",
	}
	paths := []string{
		"/pages/create",
		"/pages/read",
		"/pages/list",
		"/pages/subpages/create",
		"/pages/subpages/read",
		"/images/1.img",
		"/images/2.img",
		"/assets/images/",
		"/groups/users/preferences/1",
		"/resources/lorem/ipsum",
	}
	status := []int{
		200,
		201,
		302,
		404,
		400,
		500,
	}
	dateFormat := "02/Jan/2006:15:04:05 -0700"
	logFormat := `127.0.0.1 - lol [%s] "%s %s HTTP/1.0" %d %d`
	for i := 0; i < 10000; i++ {
		t := time.Now()
		l := fmt.Sprintf(
			logFormat,
			t.Format(dateFormat),
			methods[rand.Intn(len(methods))],
			paths[rand.Intn(len(paths))],
			status[rand.Intn(len(status))],
			rand.Intn(10000),
		)
		fmt.Fprintln(w, l)
		err := w.Flush()
		fmt.Println(l)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}
