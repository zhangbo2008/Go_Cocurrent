package main

import "fmt"
import (
	pt "awesomeProject8/gojsonpointer/pointer"
	"encoding/json"
)

var jsonText = `{
	"name": "Bobby B",
	"occupation": {
		"title" : "King",
		"years" : 15,
		"heir" : "Joffrey B"
	}
}`
func main(){
	var jsonDocument map[string]interface{}     // 建立一个 string 到 任意的 map.

	json.Unmarshal([]byte(jsonText), &jsonDocument) // unmarshal 把string 解析到map

	//create a JSON pointer
	pointerString := "/occupation/title"
	pointer2, _ := pt.NewJsonPointer(pointerString)

	//SET a new value for the "title" in the document
	pointer2.Set(jsonDocument, "Supreme Leader of Westeros")

	//GET the new "title" from the document
	title, _, _ := pointer2.Get(jsonDocument)
	fmt.Println(title) //outputs "Supreme Leader of Westeros"

	//DELETE the "heir" from the document

	deletePointer, _ := pt.NewJsonPointer("/occupation/heir")
	deletePointer.Delete(jsonDocument)

	b, _ := json.Marshal(jsonDocument)
	fmt.Println(string(b))
	//outputs `{"name":"Bobby B","occupation":{"title":"Supreme Leader of Westeros","years":15}}`
}
