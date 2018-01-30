package data

import (
	"errors"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	gaws "github.com/hairyhenderson/gomplate/aws"
)

var modeOneLevel = "mode:one-level"
var modeRecursive = "mode:recursive"

func parseAWSSMPArgs(origPath string, args ...string) (paramPath, mode string, err error) {
	paramPath = origPath
	if len(args) >= 1 {
		paramPath = path.Join(paramPath, args[0])
	}

	if len(args) >= 2 {
		if modeRecursive == args[1] {
			mode = modeRecursive
		} else if modeOneLevel == args[1] {
			mode = modeOneLevel
		} else {
			err = errors.New("Optional third argument to aws+smp datasource must be 'mode:recursive' or 'mode:one-level'")
		}
	}

	if len(args) >= 3 {
		err = errors.New("Maximum three arguments to aws+smp datasource: dataSourceName extraPath, mode:(one-level/recusive)")		
	}
	return
}

func readAWSSMP(source *Source, args ...string) (output []byte, err error) {
	if source.ASMPG == nil {
		source.ASMPG = ssm.New(gaws.SDKSession())
	}

	paramPath, mode, err := parseAWSSMPArgs(source.URL.Path, args...)

	if mode != "" {
		source.Type = json_array_mimetype
		output, err = readAWSSMPParamsByPath(source, paramPath, mode == modeRecursive)
	} else {
		source.Type = json_mimetype
		output, err = readAWSSMPParam(source, paramPath)
	}
	return
}

func readAWSSMPParam(source *Source, paramPath string) ([]byte, error) {
	input := &ssm.GetParameterInput{
		Name: aws.String(paramPath),
		WithDecryption: aws.Bool(true),
	}

	response, err := source.ASMPG.GetParameter(input)

	if err != nil {
		logFatalf("Error reading aws+smp from AWS using GetParameter with input %v:\n%v",
			input, err)
		return nil, err
	}
	
	result := *response.Parameter

	output := ToJSON(result)
	return []byte(output), nil
}

func readAWSSMPParamsByPath(source *Source, paramPath string, recursive bool) ([]byte, error) {
	input := &ssm.GetParametersByPathInput{
		Path: aws.String(paramPath),
		Recursive: aws.Bool(recursive),
		WithDecryption: aws.Bool(true),
	}

	results := make([]*ssm.Parameter, 0, 50)
	err := source.ASMPG.GetParametersByPathPages(input,
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
					results = append(results, page.Parameters...)
					return true
	})

	if err != nil {
		logFatalf("Error reading aws+smp from AWS using GetParametersByPath with input %v:\n%v",
			input, err)
		return nil, err
	}

	output := ToJSON(results)
	return []byte(output), nil
}
