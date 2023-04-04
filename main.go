package main

// Проверять тут https://confluence.veon.com/pages/viewpage.action?pageId=190188828
// отпуска https://confluence.veon.com/pages/viewpage.action?pageId=227713180
import (
	"log"
	"net/http"

	sw "github.com/moznobkin/jira-timesheet/generated/go"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
