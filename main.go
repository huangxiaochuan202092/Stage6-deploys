package main

import (
	"fmt"
	"log"
	"os"
	"proapp/routes"
)

func main() {
	// 添加工作目录检查
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("当前工作目录: %s\n", dir)

	r := routes.InitRouter()

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
