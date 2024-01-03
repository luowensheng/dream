package dream

import (
	"fmt"
	"log"
	"os"
	"strings"
	"net/http"
)

type Handler = func(http.ResponseWriter, *http.Request)

type App struct { endpoints map[string]Handler }

func CreateServer() App {
	app := App{}
	app.endpoints = make(map[string]Handler)
	return app
}

func (app *App) Route(method string, path string, handler func(http.ResponseWriter, *http.Request)){
    key := method + "@"+ path
	app.endpoints[key] = handler
}

func (app *App) Expose(path string, dir string){
    http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(dir))))
}

func (app *App) Listen(port uint){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    path := r.URL.Path
	    method := r.Method
		key := method + "@"+ path
		handle, ok := app.endpoints[key]
		if (!ok){
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		handle(w, r)
	})
	fmt.Printf(fmt.Sprintf("Starting server at port http://localhost:%d\n", port))
    if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil); err != nil {
        log.Fatal(err)
    }

}

func ReadFile(path string) (string, error){
    content, err := os.ReadFile(path)
    if err != nil {
        log.Fatal(err)
		// return "", err
    }
    return string(content), nil
}

func Render(w http.ResponseWriter, path string, vars map[string]any){
	content, err := ReadFile(path)
	if err!=nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if (vars != nil){
		for key, value := range vars {
            value_str := fmt.Sprintf("%v", value)
            key_str := fmt.Sprintf("{\"%s\"}", key)
			content = strings.Replace(content, key_str, value_str, -1)
		}	
	}
    fmt.Fprintf(w, content)
}

