package helper

import (
	"encoding/json"
	"fmt"
	ddlambda "github.com/DataDog/datadog-lambda-go"
	"reflect"
	"regexp"
	"strings"
)

var (
	DataDogHandle DataDogHelper = DataDogHelperImpl{IsVerbose}
	IsVerbose     bool
)

const (
	metric_prefix = "tfm"
)

type datadogLogMessage struct {
	Level   string `json:"level"`
	Details string `json:"details"`
}
type DataDogHelper interface {
	LogError(msg ...interface{})
	LogInfo(msg ...interface{})
	LogWarn(msg ...interface{})
	LogDebug(msg ...interface{})
	AddMetric(metric string, value float64)
	LogErrorWithInterface(errObject interface{}, msg ...interface{})
}

type DataDogHelperImpl struct {
	IsVerbose bool
}

func (DataDogHelperImpl) LogError(msg ...interface{}) {
	message("ERROR ", msg...)
}

func (DataDogHelperImpl) LogInfo(msg ...interface{}) {
	message("INFO ", msg...)
}

func (DataDogHelperImpl) LogWarn(msg ...interface{}) {
	message("WARN ", msg...)
}

func (di DataDogHelperImpl) LogDebug(msg ...interface{}) {
	if di.IsVerbose {
		message("DEBUG_INFO ", msg...)
	}
}

func (DataDogHelperImpl) AddMetric(metric string, value float64) {
	metricName := fmt.Sprintf("%s.%s", metric_prefix, metric)
	message("Pushing metric: %s, value: %v", metricName, value)
	ddlambda.Metric(metricName, value)
}
func (DataDogHelperImpl) LogErrorWithInterface(errObject interface{}, msg ...interface{}) {
	defer func() {
		if err := recover(); err != nil {
			message("something went wrong in LogErrorWithInterface", err)
		}
	}()
	errorObjectMap := make(map[string]interface{})
	errorObjectType := reflect.TypeOf(errObject)
	if errorObjectType == nil {
		return
	}
	var errorText string
	if errorMethod, found := errorObjectType.MethodByName("Error"); found {
		result := errorMethod.Func.Call([]reflect.Value{reflect.ValueOf(errObject)})
		errorText = stringiFy(result[0].String())
		errorObjectMap["errorMessage"] = errorText
	} else {
		errorObjectMap = JMap(errObject)
	}
	if msg != nil {
		message("ERROR ", msg...)
	}
}

func stringiFy(text string) (errR string) {
	defer func() {
		if err := recover(); err != nil {
			errR = "some thing went wrong while stringiFying error "
		}
	}()
	reg, _ := regexp.Compile("[^a-zA-Z0-9\\s\\t]+")
	processedString := reg.ReplaceAllString(text, "-")
	return processedString
}

// a returns a well formatted json
func MapToString(jsonBody interface{}) string {
	s, _ := json.Marshal(&jsonBody)
	return fmt.Sprintf("%s", s)
}

// on error a panic returned
func JMap(objectToDeserialize interface{}) map[string]interface{} {
	// convert the interface to an string
	s := MapToString(objectToDeserialize)

	var iFace map[string]interface{}
	// deserialize object to map[string]interface{}
	err := json.Unmarshal([]byte(s), &iFace)
	if err != nil {
		DataDogHandle.LogError(err.Error())
	}
	return iFace
}

func message(level string, msg ...interface{}) {
	logMsg := createMessage(level)
	logMsg.Details = fmt.Sprintln(msg...)
	logMsg.Details = strings.TrimSuffix(logMsg.Details, "\n")
	fmt.Printf("%s\n", fmt.Sprintf("TFMLOG:%s", MapToString(logMsg)))
}

func createMessage(level string) datadogLogMessage {
	return datadogLogMessage{Level: level}
}
