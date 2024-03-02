package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	// "strings"
	"regexp"
	"strconv"

	"io"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

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

type (
	columnType struct {
		Name      			string
		Path      			string
		Regex     			string
		RegexEnabled		bool
		Converter 			string
		ConverterEnabled	bool
	}

	filterType struct {
		Field     			string
		When      			string
		Operation 			string
		Value     			string
	}

	queryModel struct {
		Columns  			[]columnType
		LoopPath 			string
		Uri      			string
		Filters  			[]filterType
	}
	datasourceSettingsStruct struct{ BaseUrl string }
)

//------------------------------Converters-----------------------------------------------------------------

func sum(values []string) (string, error) {
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

func int2Ip(valueStr []string) (string, error) {
	if len(valueStr) != 1 {
		return "", errors.New(fmt.Sprintf("int2Ip received %d values, it should have received 1", len(valueStr)))
	}
	value, err := strconv.Atoi(valueStr[0])
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error converting %s to integer", valueStr))
	}
	b1 := (value >> 24) & 0xff
	b2 := (value >> 16) & 0xff
	b3 := (value >> 8) & 0xff
	b4 := value & 0xff
	ip := fmt.Sprintf("%d.%d.%d.%d", b1, b2, b3, b4)
	return ip, nil
}

func First(values []string) (string, error) {
	value := values[0]
	valueF, err := strconv.ParseFloat(value, 64)
	if err == nil {
		valueI := int(valueF)
		value = strconv.Itoa(valueI)
	}
	return value, nil
}

var converters = map[string]func([]string) (string, error){"int2ip": int2Ip, "sum": sum}

// _____________________________________________________________________________________________________

// Filter array elements

func filterArray[T any](values []T, test func(T) bool) (ret []T) {
	for _, value := range values {
		if test(value) {
			ret = append(ret, value)
		}
	}
	return ret
}

// ------------------------------Filters------------------------------------------------------------------
func Equals(values []string, filterValues filterType) (bool, error) {
	logger.Info(fmt.Sprintf("Running Equals filter on %s", values))
	for _, value := range values {
		if value != filterValues.Value {
			return false, nil
		}
	}
	return true, nil
}

func NotEquals(values []string, filterValues filterType) (bool, error) {
	for _, value := range values {
		if value == filterValues.Value {
			return false, nil
		}
	}
	return true, nil
}

func GreaterThan(values []string, filterValues filterType) (bool, error) {
	for _, value := range values {
		if value < filterValues.Value {
			return false, nil
		}
	}
	return true, nil

}

func LessThan(values []string, filterValues filterType) (bool, error) {
	for _, value := range values {
		if value > filterValues.Value {
			return false, nil
		}
	}
	return true, nil

}

func RegexMatch(values []string, filterValues filterType) (bool, error) {
	for _, value := range values {
		accept, err := regexp.MatchString(filterValues.Value, value)
		if err != nil {
			return false, err
		}
		if !accept {
			return false, nil
		}
	}
	return true, nil
}

func NotRegexMatch(values []string, filterValues filterType) (bool, error) {
	for _, value := range values {
		accept, err := regexp.MatchString(filterValues.Value, value)
		if err != nil {
			return false, err
		}
		if accept {
			return false, nil
		}
	}
	return true, nil
}

var filterMapping = map[string]func([]string, filterType) (bool, error){
	"equals":      Equals,
	"!equals":       NotEquals,
	"gt":            GreaterThan,
	"lt":            LessThan,
	"regexMatch":    RegexMatch,
	"regexNotMatch": NotRegexMatch,
}

//_______________________________________________________________________________________________________

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

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	frame := data.NewFrame("response")

	var queryData queryModel

	logger.Debug(fmt.Sprintf("Raw Query: %s", query.JSON))
	err := json.Unmarshal(query.JSON, &queryData)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshalling query.JSON, %s", err.Error()))
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Error unmashalling query.JSON %v", err.Error()))
	}
	logger.Debug(fmt.Sprintf("Query: uri %s, loopPath %s", queryData.Uri, queryData.LoopPath))

	var odlResponse map[string]interface{}

	err1 := d.OdlGet(pCtx, queryData.Uri, &odlResponse)
	if err1 != nil {
		logger.Error(err1.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, err1.Error())
	}

	var err2 error
	values := d.GetData(queryData, odlResponse, &err2)
	if err2 != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Error generating response data %v", err.Error()))
	}

	logger.Info(fmt.Sprintf("query: values size %d", len(values)))

	var columnValues [][]string
	for i, row := range values {
		for j, value := range row {
			logger.Debug(fmt.Sprintf("Value %d row %d", j, i))
			if j >= len(columnValues) {
				columnValues = append(columnValues, []string{})
			}
			columnValues[j] = append(columnValues[j], value)
		}
	}

	if (len(queryData.Columns) == len(columnValues)) {
		for j, column := range queryData.Columns {
			frame.Fields = append(frame.Fields, data.NewField(column.Name, nil, columnValues[j]))
		}

		response.Frames = append(response.Frames, frame)
	} else {
		logger.Info(fmt.Sprintf("Data length (%d) != column length (%d)", len(columnValues), len(queryData.Columns)))
	}

	return response
}

func (d *Datasource) OdlGet(pCtx backend.PluginContext, uri string, response *map[string]any) error {

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
	return fmt.Sprintf("%s%s", baseUrl, uri)
}

func (d *Datasource) GetData(query queryModel, odlResponse map[string]any, err *error) [][]string {
	var data [][]string
	logger.Info(fmt.Sprintf("ODL Response size %d", len(odlResponse)))
	loopNodesAny, err1 := jsonpath.Get(query.LoopPath, odlResponse)
	if err1 != nil {
		*err = errors.New(fmt.Sprintf("Error fetching loop path nodes: %v", err1.Error()))
		return nil
	}
	logger.Info(fmt.Sprintf("Looping on %s", query.LoopPath))
	loopNodes, ok := loopNodesAny.([]any)
	if ok {
		var loopNode map[string]any
		var row []string
		var value any
		var accepted bool

		for i, loopNodeAny := range loopNodes {
			loopNode, ok = loopNodeAny.(map[string]any)
			if ok {
				row = []string{}
				accepted = false
				for _, column := range query.Columns {
					value, accepted, err1 = d.GetColumnValue(loopNode, column, query.Filters)
					if err1 != nil {
						*err = errors.New(fmt.Sprintf("Error fetching value %d for column %s: %v", i, column.Name, err1.Error()))
						return nil
					}
					if !accepted {
						logger.Debug(fmt.Sprintf("Not accepted"))
						break
					}
					logger.Debug(fmt.Sprintf("Accepted!"))
					valueStr, ok := value.(string)
					if ok {
						logger.Debug(fmt.Sprintf("Appending %s to %s", valueStr, row))
						row = append(row, valueStr)
					} else {
						*err = errors.New(fmt.Sprintf("Error converting value of type %T to string for column %s", value, column.Name))
						return nil
					}
				}
				if accepted {
					logger.Debug(fmt.Sprintf("Appending row %s", row))
					data = append(data, row)
				} else {
					logger.Debug(fmt.Sprintf("Row skipped"))
				}
			} else {
				*err = errors.New(fmt.Sprintf("Error casting loop path node %s of type %T to map[string]any", loopNode, loopNode))
				return nil
			}
		}
	} else {
		*err = errors.New(fmt.Sprintf("Error casting jsonpath response %s of type %T to []any. %v", loopNodesAny, loopNodesAny))
		return nil
	}

	return data
}

// For now, if ONE filter rejects the value, it is rejected.
func (d *Datasource) Filter(values []string, filters []filterType) (bool, error) {
	for _, filter := range filters {
		logger.Debug(fmt.Sprintf("Applying filter %s on %s", filter, values))
		accept, err := filterMapping[filter.Operation](values, filter)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error filtering value %s with filter filter %s. %v", values, filter, err.Error()))
		}
		if !accept {
			logger.Debug(fmt.Sprintf("Filter %s excluded %s", filter, values))
			return false, nil
		}
	}
	return true, nil
}

func (d *Datasource) GetColumnValue(loopNode map[string]any, column columnType, columnFilters []filterType) (any, bool, error) {
	columnFilters = filterArray(columnFilters, func(filter filterType) bool { return filter.Field == column.Name })

	var val string
	var values []string
	var ok bool
	valueAny, err := jsonpath.Get(column.Path, loopNode)
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("Error extracting %s from JSON response: %v", column.Path, err.Error()))
	}
	val, ok = valueAny.(string)
	if ok {
		values = []string{val}
	} else {
		values, ok = valueAny.([]string)
		if !ok {
			return nil, false, errors.New(fmt.Sprintf("%s (%T) is not a string or []string", valueAny, valueAny))
		}
	}

	usedFilters := filterArray(columnFilters, func(filter filterType) bool { return filter.When == "raw" })
	logger.Debug(fmt.Sprintf("UsedFilters 'raw' %s", usedFilters))

	accepted, err := d.Filter(values, usedFilters)
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("Error applying raw filters. %v", err.Error()))
	}
	if !accepted {
		return "", false, nil
	}
	if (column.RegexEnabled && column.Regex != "") {
		values, err = d.applyRegex(values, column.Regex)
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Error applying regex. %v", err.Error()))
		}
		usedFilters = filterArray(columnFilters, func(filter filterType) bool { return filter.When == "regex" })
		logger.Debug(fmt.Sprintf("UsedFilters 'regex' %s", usedFilters))
		accepted, err = d.Filter(values, usedFilters)
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Error applying raw regex. %v", err.Error()))
		}
		if !accepted {
			return "", false, nil
		}
	}
	if (column.ConverterEnabled && column.Converter != "none") {
		logger.Debug(fmt.Sprintf("Applying converter %s", column.Converter))
		values, err = d.applyConverter(values, column.Converter)
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Error applying converter. %v", err.Error()))
		}
		usedFilters = filterArray(columnFilters, func(filter filterType) bool { return filter.When == "conversion" })
		logger.Debug(fmt.Sprintf("UsedFilters 'conversion' %s", usedFilters))
		accepted, err = d.Filter(values, usedFilters)
		if err != nil {
			return nil, false, errors.New(fmt.Sprintf("Error applying raw regex. %v", err.Error()))
		}
		if !accepted {
			return "", false, nil
		}
	}
	if (len(values) == 1) {
		return values[0], true, nil
	} else {
		return nil, true, nil
	}
}

func (d *Datasource) applyRegex(valueArr []string, columnRegex string) ([]string, error) {
	var response []string
	var compiledRegex = regexp.MustCompile(columnRegex)
	for _, value := range valueArr {
		match := compiledRegex.FindStringSubmatch(value)
		logger.Debug(fmt.Sprintf("Applying %s to %s generated %d matches", columnRegex, value, len(match)))
		if len(match) > 0 {
			logger.Debug(fmt.Sprintf("New value after regex %s", match[1]))
			response = append(response, match[1])
		} else {
			response = append(response, "")
			logger.Error(fmt.Sprintf("Regex %s did not match %s", columnRegex, value))
		}
	}
	return response, nil
}

func (d *Datasource) applyConverter(values []string, converterName string) ([]string, error) {
	newValue, err := converters[converterName](values)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error applying converter: %s", err.Error()))
	}
	returnVal := []string{newValue}
	return returnVal, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var pluginData map[string]interface{}
	json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &pluginData)
	var message = fmt.Sprintf("Data source is working. BaseUrl %s", pluginData["baseUrl"])

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
