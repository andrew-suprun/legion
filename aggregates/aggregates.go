package aggregates

import (
	"legion/es"
	"log"
)

func Aggregate(entity es.Info, event es.Info) {
	aggregateInfo(entity, event)
}

func aggregateInfo(entity es.Info, event es.Info) {
	for k, eventValue := range event {
		if eventValue == nil {
			delete(entity, k)
			continue
		}
		switch eventValueInfo := eventValue.(type) {
		case int, float64, string:
			entity[k] = eventValue
		case es.Info:
			if entityValueInfo, ok := entity[k].(es.Info); ok {
				aggregateInfo(entityValueInfo, eventValueInfo)
			} else {
				entity[k] = eventValue
			}
		default:
			log.Fatalf("ERROR: Diff doesn't support %[1]v %[1]T\n", eventValue)
		}
	}
}

func Diff(oldEntity, newEntity es.Info) (diff es.Info) {
	diff = es.Info{}
	diffInfoRemoved(oldEntity, newEntity, diff)
	diffInfo(oldEntity, newEntity, diff)
	return diff
}

func diffInfo(oldInfo, newInfo, diff es.Info) {
	for k, newValue := range newInfo {
		if newValue == nil {
			continue
		}
		oldValue := oldInfo[k]
		switch newElement := newValue.(type) {
		case int:
			if oldString, ok := oldValue.(int); !ok || oldString != newElement {
				diff[k] = newValue
			}
		case float64:
			if oldString, ok := oldValue.(float64); !ok || oldString != newElement {
				diff[k] = newValue
			}
		case string:
			if oldString, ok := oldValue.(string); !ok || oldString != newElement {
				diff[k] = newValue
			}
		case es.Info:
			if oldInfoElement, ok := oldValue.(es.Info); ok {
				elementDiff := es.Info{}
				diffInfo(oldInfoElement, newElement, elementDiff)
				if len(elementDiff) > 0 {
					diff[k] = elementDiff
				}
			} else {
				diff[k] = newValue
			}
		default:
			log.Fatalf("ERROR: Diff doesn't support %[1]v %[1]T\n", newValue)
		}
	}
}

func diffInfoRemoved(oldInfo, newInfo, diff es.Info) {
	for k, oldValue := range oldInfo {
		newValue, ok1 := newInfo[k]
		newValueInfo, ok2 := newValue.(es.Info)
		if !ok1 || !ok2 {
			diff[k] = nil
		} else if oldValueInfo, ok := oldValue.(es.Info); ok {
			elementDiff := es.Info{}
			diffInfoRemoved(oldValueInfo, newValueInfo, elementDiff)
			if len(elementDiff) > 0 {
				diff[k] = elementDiff
			}
		}
	}
}
