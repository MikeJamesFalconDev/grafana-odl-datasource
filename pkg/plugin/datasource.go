package plugin

import (
	"errors"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"strconv"
	"regexp"

    "net/http"
    "io"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/PaesslerAG/jsonpath"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

var logger = log.New()

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, _ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()
	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type (
	columnType struct 	{ 
		Name  		string
		Path  		string
		Regex 		string
		Converter	string
	}
	queryModel struct	{ 
		Columns 	[]columnType
		LoopPath 	string
		Uri 		string
	}
	datasourceSettingsStruct struct{ BaseUrl string }
)


func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames
	frame := data.NewFrame("response")

	// input := []byte(`{ "uri" : "/rests/data/network-topology:network-topology" }`)
	
	var queryData queryModel

	logger.Debug(fmt.Sprintf("Raw Query: %s", query.JSON))
	err := json.Unmarshal(query.JSON, &queryData)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshalling query.JSON, %s", err.Error()))
		return backend.ErrDataResponse(backend.StatusBadRequest,fmt.Sprintf("Error unmashalling query.JSON %v", err.Error()))
	}
	logger.Debug(fmt.Sprintf("Query: uri %s, loopPath %s", queryData.Uri, queryData.LoopPath))

	var odlResponse map[string]interface{}

	err1 := d.OdlGet(pCtx, queryData.Uri, &odlResponse)
	if err1!= nil {
        logger.Error(err1.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, err1.Error())
	}
	
	// logger.Info(fmt.Sprintf("odlResponseKeys %d", len(odlResponse)))
	for _, column := range queryData.Columns {
		logger.Debug(fmt.Sprintf("Column name: %s, path: %s", column.Name, column.Path))
		frame.Fields = append(frame.Fields, data.NewField(column.Name, nil, d.GetValues(queryData.LoopPath, column, odlResponse)))
	}

	response.Frames = append(response.Frames, frame)

	return response
}

func (d *Datasource) OdlGet(pCtx backend.PluginContext, uri string, response *map[string]any ) error {

	var pluginData map[string]interface{}
	err := json.Unmarshal(pCtx.DataSourceInstanceSettings.JSONData, &pluginData)
    if err != nil {
		return errors.New(fmt.Sprintf("Error unmashalling PluginContext.JSONData %v", err.Error()))		
    }
	logger.Debug("Extracting parameters")

	url := d.GetUrl(pluginData, uri)
    req, err1 := http.NewRequest("GET", url, nil)
    if err1 != nil {
		return errors.New(fmt.Sprintf("Error building http request %v", err1.Error()))
    }
	logger.Debug("Building header")
	req.Header.Add("Accept", "application/json")
	logger.Debug("Performing request")
	res, err2 := http.DefaultClient.Do(req)
    if err2 != nil {
		return errors.New(fmt.Sprintf("Error in http get %v", err2.Error()))
    }
	logger.Debug("Deferred close")
	defer res.Body.Close()
	logger.Debug("Reading response")
	body, readErr := io.ReadAll(res.Body)
    if readErr != nil || body == nil {
		return errors.New(fmt.Sprintf("Error extracting http response %v", readErr.Error()))
    }
	logger.Debug(fmt.Sprintf("odlResponse: %s", body[:150]))

	err3 := json.Unmarshal(body, response)
	if err3 != nil {
		return errors.New(fmt.Sprintf("Error unmashalling query.JSON %v", err3.Error()))
	}
	logger.Debug("End odlGet")
	return nil
}

func (d *Datasource) GetUrl(pluginData map[string]any, uri string) string {
	baseUrl := pluginData["baseUrl"]
	logger.Debug(fmt.Sprintf("Base URL: %s\tURI:%s", baseUrl, uri))
	return fmt.Sprintf("%s%s",baseUrl, uri)
}

func (d *Datasource) GetValues(loopPath string, column columnType, odlResponse map[string]any) []string {
	loopNodesAny, err := jsonpath.Get(loopPath, odlResponse)
	// loopNodesAny, err := d.GetNode(loopPath, odlResponse)
	logger.Info(fmt.Sprintf("Looping on %s", loopPath))
	var response []string
	if err != nil {
		logger.Error(err.Error())
		return response
	}
	loopNodes, ok := loopNodesAny.([]any)
	logger.Debug(fmt.Sprintf("Casting loop nodes. ok: %s. Actual type: %T", ok, loopNodesAny))
	if ok {
		logger.Info(fmt.Sprintf("Casting loop list worked (%d)", len(loopNodes)))
		var value any
		var loopNode map[string]any
		var err1 error
		for _, loopNodeAny := range loopNodes {
			logger.Debug(fmt.Sprintf("Searching for column value %s", column.Path))
			loopNode, ok = loopNodeAny.(map[string]any)
			if ok {
				value, err1 = jsonpath.Get(column.Path, loopNode)
			// 	valueAny, err1 = d.GetNode(column.Path, loopNode)
				if err1 != nil {
					logger.Error(err1.Error())
					value = ""
				}
				var values []string
				values, err = d.applyRegex(value, column.Regex)
				if err != nil {
					logger.Error(fmt.Sprintf("Error applying regex. %s", err.Error()))
				}
				value, err = d.applyConverter(values, column.Converter)
				if err != nil {
					logger.Error(fmt.Sprintf("Error applying converter. %s", err.Error()))
				}
				logger.Debug(fmt.Sprintf("Appending value %s", value))
				valueStr, ok := value.(string)
				if ok {
					response = append(response, valueStr)
				}
			} else {
				logger.Error(fmt.Sprintf("Could not convert %T to map[string]any", loopNodeAny))
			}
		}
	} else {
		logger.Error(fmt.Sprintf("Failed to typify loopNode list. %T", loopNodesAny))
	}
	logger.Debug(fmt.Sprintf("Finished getting values. Response: %s",response))
	return response
}

func (d *Datasource) applyRegex(valueAny any, columnRegex string) ([]string, error) {
	var valueArr []any
	val, ok := valueAny.(string)
	if ok {
		valueArr = []any { val }
	} else {
		valueArr, ok = valueAny.([]any)
		if !ok {
			return nil, errors.New(fmt.Sprintf("%s (%T) is not a string or []any", valueAny, valueAny))
		}
	}
	// logger.Info(fmt.Sprintf("Column Regex %s", columnRegex))
	var response []string
	if (columnRegex != "") {
		var compiledRegex = regexp.MustCompile(columnRegex)
		for _, valueAny := range valueArr {
			value, ok := valueAny.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Could not convert %T to string", valueAny))
				continue
			}
			match := compiledRegex.FindStringSubmatch(value)
			logger.Info(fmt.Sprintf("Applying %s to %s generated %d matches",columnRegex, value, len(match)))
			if (len(match) > 0) {
				logger.Info(fmt.Sprintf("New value after regex %s",match[1]))
				response = append(response, match[1])
			} else {
				response = append(response, "")
				logger.Error(fmt.Sprintf("Regex %s did not match %s", columnRegex, value))
			}
		}
		return response, nil
	} else {
		for _, valAny := range valueArr {
			val, ok := valAny.(string)
			if ok {
				response = append(response, val)
			}
		}
		return response, nil
	}
}

func (d *Datasource) sum(values []string) (string, error) {
	accum := 0
	for _, valueStr := range values {
		val, err := strconv.ParseFloat(valueStr, 64)
		if err == nil {
			accum += int(val)
		} else {
			logger.Error(fmt.Sprintf("Could not convert %s to integer", valueStr))
		}
	}
	return strconv.Itoa(accum), nil
}

func (d *Datasource) int2Ip(valueStr []string) (string, error) {
	if (len(valueStr) != 1) {
		return "", errors.New(fmt.Sprintf("int2Ip received %d values, it should have received 1", len(valueStr)))
	}
	value, err := strconv.Atoi(valueStr[0])
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error converting %s to integer", valueStr))
	}
	b1 := (value>>24)&0xff
	b2 := (value>>16)&0xff
	b3 := (value>>8)&0xff
	b4 := value &0xff
	ip := fmt.Sprintf("%d.%d.%d.%d",b1,b2,b3,b4)
	return ip, nil
}

func (d *Datasource) First(values []string) (string, error) {
	value := values[0]
	valueF, err := strconv.ParseFloat(value, 64)
	if (err == nil) {
		valueI := int(valueF)
		value = strconv.Itoa(valueI)
	}
	return value, nil
}

func (d *Datasource) applyConverter(values []string, converterName string) (string, error) {
	converters := map[string]func([]string)(string,error) { "int2ip" : d.int2Ip, "none": d.First, "sum": d.sum}
	newValue, err := converters[converterName](values)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error applying converter: %s", err.Error()))
	}
	return newValue, nil
}

func (d *Datasource) GetKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (d *Datasource) GetNode(path string, odlResponse map[string]any) (any, error) {
	pathNodes := strings.Split(path,"/")
	var ok bool
	var arrayNode []any
	var previousNode map[string]any
	currentNode := odlResponse
	logger.Debug(fmt.Sprintf("Path has %d nodes", len(pathNodes)))
	for i, nodeName := range pathNodes {

		previousNode = currentNode
		if strings.Contains(nodeName, "[") {
			openB := strings.Index(nodeName,"[")
			closeB := strings.Index(nodeName,"]")
			nodePart := nodeName[:openB]
			index, err := strconv.Atoi(nodeName[openB+1:closeB])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Could not extract index. %s",err.Error()))
			}
			arrayNode, ok = currentNode[nodePart].([]any)
			if ok  {
				if index < len(arrayNode) {
					currentNode, ok = arrayNode[index].(map[string]any)
					if ok && currentNode != nil {
						logger.Debug(fmt.Sprintf("Found array item %d, in %s", index, nodePart))
					} else {
						return nil, errors.New(fmt.Sprintf("Found index %d but it is wrong type %T",index, arrayNode))
					}
				} else {
					return nil, errors.New(fmt.Sprintf("Node Index out of bounds %d >= %s",index, len(arrayNode)))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("Node not array %s (%T)",nodeName, currentNode[nodePart]))
			}
		} else {
			if i == len(pathNodes) -1 {
				logger.Debug(fmt.Sprintf("Found %s", nodeName))
				return currentNode[nodeName], nil
			}
			previousNode = currentNode
			currentNode, ok = currentNode[nodeName].(map[string]any)
			if !ok  {
				return nil, errors.New(fmt.Sprintf("Child node %s not found in %s. (%T)", nodeName, d.GetKeys(previousNode), previousNode))
			}
		}
		if currentNode != nil {
			logger.Debug(fmt.Sprintf("Found %s", nodeName))
		} else {
			logger.Error(fmt.Sprintf("Parent node: %s", previousNode))
			return nil, errors.New(fmt.Sprintf("Node not found %s in %s", nodeName, d.GetKeys(previousNode)))
		}
	}
	logger.Debug(fmt.Sprintf("GetNode response: %s",currentNode))
	return currentNode, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var pluginData map[string]interface{}
	json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &pluginData)
	var message = fmt.Sprintf("Data source is working. BaseUrl %s",pluginData["baseUrl"])

	logger.Debug(fmt.Sprintf("%s", message))
	// if rand.Int()%2 == 0 {
	// 	status = backend.HealthStatusError
	// 	message = "randomized error"
	// }

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
