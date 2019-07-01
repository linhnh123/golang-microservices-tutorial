package transformer

import (
	"encoding/json"
	"fmt"
	"log"
)

func ProcessLogStatement(item map[string]interface{}) ([]byte, error) {
	if item["short_message"] != nil {
		shortMessageString := item["short_message"].(string)

		var shortMessage map[string]interface{}
		err := json.Unmarshal([]byte(shortMessageString), &shortMessage)
		if err != nil {
			log.Printf("Error parsing short_message: %v\n", err.Error())
			return nil, fmt.Errorf("Error parsing 'short_message' property.")
		}

		if shortMessage != nil {
			item["msg"] = shortMessage["msg"].(string)
			item["level"] = shortMessage["level"].(string)
			delete(item, "short_message")
		} else {
			log.Println("Found log item with unparsable short_message: " + shortMessageString)
			return nil, fmt.Errorf("Found log item with unparsable 'short_message' property.")
		}
		log.Println(item["msg"].(string))

		return json.Marshal(item)
	} else {
		return nil, fmt.Errorf("Could not process log statement, missing 'short_message' property.")
	}
}
