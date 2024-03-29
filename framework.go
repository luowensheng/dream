package dream

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type jsCommand struct {
	CommandType string `json:"type"`
	Command     string `json:"command"`
}

type jSResponse struct {
	Call   string `json:"call"`
	Params Record `json:"params"`
}

type serverResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Commands []jsCommand `json:"commands"`
}

type EventHandler = func(Record)

type domVariableResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type elementCreator struct {
	setChild     func(*Element)
	firstElement *Element
}

func newCreator() *elementCreator {
	return &elementCreator{
		setChild:     nil,
		firstElement: nil,
	}
}

type PageContext struct {
	*elementCreator
	commands     []jsCommand
	events       map[string]EventHandler
	broacasts    map[string][]EventHandler
	context      map[string]string
	hasExcecuted bool
}

func createManager() *PageContext {
	return &PageContext{
		context:      map[string]string{},
		elementCreator:      newCreator(),
		events:       map[string]EventHandler{},
		broacasts:    map[string][]EventHandler{},
		hasExcecuted: false,
	}
}

func (p *PageContext)clone() *PageContext {
	return &PageContext{
		context:      p.context,
		elementCreator:      newCreator(),
		events:       p.events,
		broacasts:    p.broacasts,
		hasExcecuted: false,
	}
}


var COUNTER int = 0

func (manager *PageContext) reset() {
	manager.setChild = nil
	manager.firstElement = nil
}

func (manager *PageContext) handle(jsResponse *jSResponse) serverResponse {

	response := serverResponse{Success: false}

	handler := manager.events[jsResponse.Call]
	if handler == nil {
		if jsResponse.Call == "error" {
			response.Success = true
		} else {
			response.Message = fmt.Sprintf("No handler for \"%s\" was not found", jsResponse.Call)
		}
		return response
	}

	manager.resetCommands()
	manager.context = make(map[string]string)
	context := jsResponse.Params["dom"]

	var data []domVariableResponse

	json.Unmarshal([]byte(context), &data)

	for i := 0; i < len(data); i++ {
		manager.context[data[i].Name] = data[i].Value
	}

	handler(jsResponse.Params)

	response.Success = true
	response.Commands = manager.commands
	return response
}

func (manager *PageContext) resetCommands() {
	manager.commands = []jsCommand{}
}

func (manager *PageContext) addCommand(command jsCommand) {
	manager.commands = append(manager.commands, command)
}

func (manager *PageContext) addEvent(name string, handler EventHandler) {
	manager.events[name] = handler
}

func (manager *PageContext) removeEvent(name string) {
	delete(manager.events, name)
}

func (manager *PageContext) addBroadcastIfNotExists(name string) {
	if manager.broacasts[name] == nil {
		manager.broacasts[name] = []func(Record){}
	}
}

func (manager *PageContext) addBroadcastListener(name string, handler EventHandler) {
	manager.addBroadcastIfNotExists(name)
	manager.broacasts[name] = append(manager.broacasts[name], handler)
}


type Record = map[string]string

type Element struct {
	Tag         string
	TextContent string
	attributes  Record
	Children    []*Element
	ref *ElementRef
}

func nextCount() int {
	x := COUNTER
	COUNTER += 1
	return x
}

func (element Element) GetId() string {
	if element.attributes["id"] == "" {
		element.attributes["id"] = fmt.Sprintf("%s-element-%d", element.Tag, nextCount())
	}
	return element.attributes["id"]
}

func (element Element) String() string {

	if element.Tag == "*" {
		return element.TextContent
	}

	innerHTML := element.TextContent
	if len(element.Children) > 0 {
		innerHTML = ""
		for i := 0; i < len(element.Children); i++ {
			innerHTML += fmt.Sprintf("\n\t%s", (element.Children[i].String()))
		}
		innerHTML += "\n"
	}

	attributeStr := ""
	for key, value := range element.attributes {
		attributeStr += fmt.Sprintf(" %s=\"%s\"", key, value)
	}

	return fmt.Sprintf(
		"<%s%s>%s</%s>",
		element.Tag, attributeStr, innerHTML, element.Tag)
}

type A = Record

type ElementRef struct {
	element *Element
	id      string
	manager *PageContext
}

func (e *ElementRef)clone() *ElementRef {

	elementClone := Element{}
	elementClone.attributes = e.element.attributes
	elementClone.Children = []*Element{}
	elementClone.Tag = e.element.Tag
	elementClone.TextContent = e.element.TextContent
	
	return &ElementRef{
		element: &elementClone,
		id: e.id,
	}
}

func (manager *PageContext) ExecuteJS(command string) {
	manager.addCommand(jsCommand{CommandType: "execute", Command: command})
}

func (manager *PageContext) ExecuteWithResponse(command string, f func(string)) {

	eventId := generateUniqueName("execute")

	paramsStr := "const data = {dom: getAllDOMVariables()};"
	params := map[string]string{}
	params["output"] = command

	for key := range params {
		paramsStr = fmt.Sprintf("%s\ndata['%s']=%s", paramsStr, key, params[key])
	}

	manager.addEvent(eventId, func(r Record) {
		f(r["output"])
		manager.removeEvent(eventId)
	})

	manager.ExecuteJS(
		fmt.Sprintf(
			`{
				%s;
				callBackend('%s', data);
			}
		`, paramsStr, eventId),
	)
}

func NewHTMLRef(id string) *ElementRef {
	return &ElementRef{id: id}
}

type State[T any] struct {
	value T
	tasks []func(T)
}

type Comparable = interface{}

type DOMVariable[T Comparable] struct {
	value   T
	name    string
	tasks   []func(T)
	manager *PageContext
}

func ToJsonString[T any](value T) string {
	bytes, err := json.Marshal(value)

	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func SetDOMVariableCommand[T Comparable](name string, value T) string {

	cmd := fmt.Sprintf("setDOMVariable('%s', %s)", name, ToJsonString(value))
	return cmd
}
func NewNamedDOMVariable[T Comparable](manager *PageContext, name string, value T) *DOMVariable[T] {

	cmd := SetDOMVariableCommand(name, value)
	manager.ExecuteJS(cmd)

	item := &DOMVariable[T]{value: value, name: name, manager: manager}
	item.OnValueUpdated(func(t T) {
		json.Marshal(t)

		manager.ExecuteJS(SetDOMVariableCommand(name, t))
	})

	return item
}

// ////////////////////////////////
func (variable DOMVariable[T]) String() string {
	return fmt.Sprint(variable.value)
}

func (variable *DOMVariable[T]) Value() T {
	value, ok := variable.manager.context[variable.name]
	if !ok {
		return variable.value
	}
	json.Unmarshal([]byte(value), &variable.value)
	delete(variable.manager.context, variable.name)
	return variable.value
}

func (variable *DOMVariable[T]) SetValue(newValue T) {
	if fmt.Sprint(variable.value) == fmt.Sprint(newValue) {
		return
	}
	variable.value = newValue
	Each(variable.tasks, func(task func(T)) {
		task(variable.value)
	})
}

func (variable *DOMVariable[T]) OnValueUpdated(f func(T)) {
	variable.tasks = append(variable.tasks, f)
}

func (variable *DOMVariable[T]) UpdateValue(f func(T) T) {
	variable.SetValue(f(variable.value))
}

// /////////////////////////////////
func (state State[T]) String() string {
	return fmt.Sprint(state.value)
}

func (state *State[T]) Value() T {
	return state.value
}

func (state *State[T]) SetValue(newValue T) {
	state.value = newValue
	for index := range state.tasks {
		state.tasks[index](state.value)
	}
}

func (state *State[T]) OnValueUpdated(f func(T)) {
	state.tasks = append(state.tasks, f)
}

func (state *State[T]) UpdateValue(f func(T) T) {
	state.value = f(state.value)
}

func UseState[T any](value T) *State[T] {
	return &State[T]{value: value}
}

func (elementRef *ElementRef) DOMContent(content *DOMVariable[string]) *ElementRef {

	content.OnValueUpdated(func(s string) {
		elementRef.SetTextContent(fmt.Sprint(s))
	})

	return elementRef.Content(content.Value())
}

func (elementRef *ElementRef) StateContent(content *State[string]) *ElementRef {

	content.OnValueUpdated(func(s string) {
		elementRef.SetTextContent(fmt.Sprint(s))
	})

	return elementRef.Content(content.value)
}

func (elementRef *ElementRef) Content(content string) *ElementRef {

	textContent := fmt.Sprint(content)
	elementRef.element.TextContent = textContent
	return elementRef
}

func (elementRef *ElementRef) Class(class string) *ElementRef {
	return elementRef.Attr("class", class)
}

func (elementRef *ElementRef) Id(id string) *ElementRef {
	return elementRef.Attr("id", id)
}

func (elementRef *ElementRef) Style(style string) *ElementRef {
	return elementRef.Attr("style", style)
}

func (elementRef *ElementRef) Attr(key, value string) *ElementRef {
	if elementRef.element == nil {
		elementRef.ExecuteJS(fmt.Sprintf("{this}.setAttribute('%s', ''%s);", key, value))
		return elementRef
	}
	elementRef.element.attributes[key] = value
	return elementRef
}

func (elementRef *ElementRef) Inner(f func()) {
	elementRef.InnerRef(func(*ElementRef) { f() })
}

func (elementRef *ElementRef) InnerRef(f func(*ElementRef)) {

	if elementRef.element == nil {
		f(elementRef)
		return
	}

	setChild := elementRef.manager.setChild
	elementRef.manager.setChild = func(e *Element) {
		elementRef.element.Children = append(elementRef.element.Children, e)
	}
	f(elementRef)
	elementRef.manager.setChild = setChild
}

func (e *ElementRef) UpdateInner(f func(newContext *PageContext)) {

	manager := e.manager.clone()

	elementRef := e.clone()

	elementRef.manager = manager

	if elementRef.element == nil {
		f(manager)
		return
	}

	setChild := elementRef.manager.setChild
	elementRef.manager.setChild = func(e *Element) {
		elementRef.element.Children = append(elementRef.element.Children, e)
	}
	f(manager)
	elementRef.manager.setChild = setChild

	e.SetInnerHTML(elementRef.element.String())

	e.manager.commands = append(e.manager.commands, manager.commands...)

	elementRef.manager = e.manager
	for _, child := range elementRef.element.Children {
		child.ref.manager = e.manager
	}
}

func generateUniqueName(name string) string {
	return fmt.Sprintf("%s_%d", name, nextCount())
}

func (elementRef *ElementRef) On(eventName string, f func()) *ElementRef {
	eventId := generateUniqueName(eventName)

	elementRef.manager.addEvent(eventId, func(r Record) { f() })
	elementRef.ExecuteJS(
		fmt.Sprintf("{this}.addEventListener(`%s`, function(event){callBackend(`%s`, {}, event);})", eventName, eventId),
	)
	return elementRef
}

func (elementRef *ElementRef) Broadcast(name, eventName string) {
	eventId := generateUniqueName(eventName)

	elementRef.manager.addBroadcastIfNotExists(name)
	elementRef.manager.addEvent(eventId, func(r Record) {
		for i := range elementRef.manager.broacasts[name] {
			elementRef.manager.broacasts[name][i](r)
		}
	})
	elementRef.ExecuteJS(
		fmt.Sprintf("{this}.addEventListener(`%s`, function(event){callBackend(`%s`, {}, event);})", eventName, eventId),
	)
}

// func (elementRef *ElementRef) RemoveListenersForEvent(event string) {
// 	elementRef.ExecuteJS(fmt.Sprintf(`
// 	const listeners = {this}.events && {this}.events["%s"];
// 	console.log({listeners})
// 	// Remove each click event listener
// 	if (listeners && listeners.length > 0) {
// 		for (const i = 0; i < listeners.length; i++) {
// 			{this}.removeEventListener("%s", listeners[i]);
// 		}
// 	}`, event, event))
// }

func (elementRef *ElementRef) OnWithParams(eventName string, f func(Record), params Record) {
	eventId := generateUniqueName(eventName)

	paramsStr := "const data = {dom: getAllDOMVariables()};"
	tag := ""
	value := ""
	for key := range params {
		value = strings.TrimSpace(params[key])
		if tag == "" && strings.HasPrefix(value, "await") {
			tag = "async "
		}
		paramsStr = fmt.Sprintf("%s\ndata['%s']=%s", paramsStr, key, value)
	}
	elementRef.manager.addEvent(eventId, f)
	elementRef.ExecuteJS(
		fmt.Sprintf(
			`{this}.addEventListener('%s', %sfunction(event){
				%s
				callBackend('%s', data, event);
			})
		`, eventName, tag, paramsStr, eventId),
	)
}

func (elementRef *ElementRef) BroadcastWithParams(name, eventName string, params Record) {
	eventId := generateUniqueName(eventName)

	paramsStr := "const data = {dom: getAllDOMVariables()};"
	for key := range params {
		paramsStr = fmt.Sprintf("%s\ndata['%s']=%s", paramsStr, key, params[key])
	}
	elementRef.manager.addBroadcastIfNotExists(name)
	elementRef.manager.addEvent(eventId, func(r Record) {
		for i := range elementRef.manager.broacasts[name] {
			elementRef.manager.broacasts[name][i](r)
		}
	})
	elementRef.ExecuteJS(
		fmt.Sprintf(
			`{this}.addEventListener('%s', function(event){
				%s
				callBackend('%s', data, event);
			})
		`, eventName, paramsStr, eventId),
	)
}

func (manager *PageContext) OnBroadcast(name string, f func()) {
	manager.addBroadcastListener(name, func(r Record) { f() })
}

func (manager *PageContext) OnBroadcastWithParams(name string, f EventHandler) {
	manager.addBroadcastListener(name, f)
}

func createParamRegex(paramName string) string {
	return "{\\s*" + paramName + "\\s*}"
}

func (elementRef *ElementRef) CreateQueryFromCommand(command string) string {
	regex := regexp.MustCompile(createParamRegex("this"))
	query := fmt.Sprintf("document.getElementById(`%s`)", elementRef.id)
	return regex.ReplaceAllString(command, query)
}

func (elementRef *ElementRef) ExecuteJS(command string) {
	elementRef.manager.addCommand(jsCommand{
		CommandType: "execute",
		Command:     elementRef.CreateQueryFromCommand(command),
	})
}
func (elementRef *ElementRef) This(command string) string {
	return elementRef.CreateQueryFromCommand(fmt.Sprintf("{this}.%v;", command))
}

func (elementRef *ElementRef) GetTextContent() string {
	return elementRef.CreateQueryFromCommand("{this}.textContent;")
}

func (elementRef *ElementRef) GetValue() string {
	return elementRef.CreateQueryFromCommand("{this}.value;")
}

func (elementRef *ElementRef) GetAttribute(key string) string {
	return elementRef.CreateQueryFromCommand(fmt.Sprintf("{this}.getAttribute('%s');", key))
}

func (elementRef *ElementRef) GetStyle(key string) string {
	return elementRef.CreateQueryFromCommand(fmt.Sprintf("{this}.style['%s'];", key))
}

func (elementRef *ElementRef) SetTextContent(textContent string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.textContent=`%s`;", textContent))
}

func (elementRef *ElementRef) SetAttribute(key, value string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.setAttribute('%s',`%s`);", key, value))
}

func (elementRef *ElementRef) SetStyle(key, value string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.style['%s']=`%s`;", key, value))
}

func (elementRef *ElementRef) SetValue(value string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.value=`%s`;", value))
}

func (elementRef *ElementRef) SetInnerHTML(value string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.innerHTML=`%s`;", value))
}

func (elementRef *ElementRef) RemoveAttribute(key string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.removeAttribute(`%s`);", key))
}

func (elementRef *ElementRef) AddClass(class string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.classList.add(`%s`);", class))
}

func (elementRef *ElementRef) RemoveClass(class string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.classList.remove(`%s`);", class))
}

func (elementRef *ElementRef) ToggleClass(class string) {
	elementRef.ExecuteJS(fmt.Sprintf("{this}.classList.toggle(`%s`);", class))
}

func (manager *PageContext) El(tag string) *ElementRef {
	if tag == "" {
		tag = "div"
	}
    element := Element{
		Tag:        tag,
		attributes: A{},
	}

	element.GetId()

	if manager.firstElement == nil {
		manager.firstElement = &element
	}
	if manager.setChild != nil {
		manager.setChild(&element)
	}

	elementRef := &ElementRef{element: &element, id: element.GetId(), manager: manager}
	element.ref = elementRef

	return elementRef
}

func (manager *PageContext) LoadHTML(localPath string) {
	content, err := ReadFile(localPath)
	if err != nil {
		log.Fatal(err)
	}
	manager.El("*").Content(fmt.Sprintf("\n%s\n", content))
}

func isLocalPath(path string) bool {
	return strings.HasPrefix(path, "./")
}

func (manager *PageContext) LoadJS(path string) {

	if !isLocalPath(path) {
		manager.El("script").Attr("src", path)
	} else {
		content, err := ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		manager.El("script").Content(fmt.Sprintf("\n%s\n", content))
	}
}

func (manager *PageContext) LoadCSS(path string) {

	if !strings.HasPrefix(path, "./") {
		manager.El("link").Attr("rel", "stylesheet").Attr("href", path)
	} else {
		content, err := ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		manager.El("style").Content(fmt.Sprintf("\n%s\n", content))
	}
}

func createHTML(manager *PageContext, title string, app func(*PageContext)) string {

	manager.El("body").Inner(func() {

		app(manager)

		if len(manager.commands) > 0 {
			manager.El("script").Content(fmt.Sprintf("\n%s\n", jsCONTENT))

			jsScript, err := json.Marshal(manager.commands)
			if err != nil {
				log.Fatal(err)
			}
			manager.El("script").Content(fmt.Sprintf(
				"document.addEventListener('DOMContentLoaded', async()=>{handleServerCommand(%s);} )", jsScript),
			)
		}

	})

	appHTML := manager.firstElement.String()
	manager.hasExcecuted = true

	html := fmt.Sprintf(
		`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
</head>
%s
		`, title, appHTML,
	)

	dirPath := "./output"

	// Check if the directory already exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		err := os.Mkdir(dirPath, 0755) // 0755 is the permission mode
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	err := os.WriteFile("./output/index.html", []byte(html), 0644)
	if err != nil {
		log.Fatal(err)
	}
	return html
}

func CreateApp(title string, port uint, app func(*PageContext)) {

	manager := createManager()
	manager.reset()

	server := CreateServer()
	html := createHTML(manager, title, app)

	server.Expose("/static/", "./public")

	server.Route("GET", "/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, html)
	})

	if len(manager.commands) > 0 {
		server.Route("POST", "/api/v1/call", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Can't read Body", http.StatusBadRequest)
				return
			}

			data := jSResponse{}
			if err := json.Unmarshal(bytes, &data); err != nil {
				http.Error(w, "Invalid json", http.StatusBadRequest)
				return
			}

			clientCommands := manager.handle(&data)

			json.NewEncoder(w).Encode(clientCommands)
		})
	}

	server.Listen(port)
}
