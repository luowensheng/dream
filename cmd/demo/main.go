package main

import (
	"fmt"
	"github.com/luowensheng/dream/internal/demo"
	"github.com/luowensheng/dream"
)

func main(){
	dream.CreateApp("t", 9030, demo.App2)
	fmt.Println("Yellow")
}