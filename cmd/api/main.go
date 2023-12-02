package main

import (
	"fmt"
	_ "go-chi-api/docs"
	"go-chi-api/internal/server"
	"log"
)

//	@title			Go Chi API
//	@version		1.0
//	@description	This is a sample server.

// @license.name	MIT
// @license.url	http://github.com/ryanseipp/go-chi-api/LICENSE

// @host localhost:3000
// @BasePath /
func main() {
	server := server.NewServer()

	log.Println(fmt.Sprintf("Starting server at %s", server.Addr))
	err := server.ListenAndServe()
	if err != nil {
		panic("cannot start server")
	}
}
