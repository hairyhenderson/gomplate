package data

import (
	"errors"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	gaws "github.com/hairyhenderson/gomplate/aws"
)

func parseAWSSMPArgs(origPath string, args ...string) (paramPath string, err error) {
	paramPath = origPath
	if len(args) >= 1 {
		paramPath = path.Join(paramPath, args[0])
	}

	if len(args) >= 2 {
		err = errors.New("Maximum two arguments to aws+smp datasource: alias, extraPath")		
	}
	return
}

func readAWSSMP(source *Source, args ...string) (output []byte, err error) {
	if source.ASMPG == nil {
		source.ASMPG = ssm.New(gaws.SDKSession())
	}

	paramPath, err := parseAWSSMPArgs(source.URL.Path, args...)

	source.Type = json_mimetype
	output, err = readAWSSMPParam(source, paramPath)
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
