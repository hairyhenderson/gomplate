package awsauth

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBackend_pathLogin_getCallerIdentityResponse(t *testing.T) {
	responseFromUser := `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::123456789012:user/MyUserName</Arn>
    <UserId>ASOMETHINGSOMETHINGSOMETHING</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>7f4fc40c-853a-11e6-8848-8d035d01eb87</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`
	expectedUserArn := "arn:aws:iam::123456789012:user/MyUserName"

	responseFromAssumedRole := `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
  <Arn>arn:aws:sts::123456789012:assumed-role/RoleName/RoleSessionName</Arn>
  <UserId>ASOMETHINGSOMETHINGELSE:RoleSessionName</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>7f4fc40c-853a-11e6-8848-8d035d01eb87</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`
	expectedRoleArn := "arn:aws:sts::123456789012:assumed-role/RoleName/RoleSessionName"

	parsedUserResponse, err := parseGetCallerIdentityResponse(responseFromUser)
	if parsed_arn := parsedUserResponse.GetCallerIdentityResult[0].Arn; parsed_arn != expectedUserArn {
		t.Errorf("expected to parse arn %#v, got %#v", expectedUserArn, parsed_arn)
	}

	parsedRoleResponse, err := parseGetCallerIdentityResponse(responseFromAssumedRole)
	if parsed_arn := parsedRoleResponse.GetCallerIdentityResult[0].Arn; parsed_arn != expectedRoleArn {
		t.Errorf("expected to parn arn %#v; got %#v", expectedRoleArn, parsed_arn)
	}

	_, err = parseGetCallerIdentityResponse("SomeRandomGibberish")
	if err == nil {
		t.Errorf("expected to NOT parse random giberish, but didn't get an error")
	}
}

func TestBackend_pathLogin_parseIamArn(t *testing.T) {
	testParser := func(inputArn, expectedCanonicalArn string, expectedEntity iamEntity) {
		entity, err := parseIamArn(inputArn)
		if err != nil {
			t.Fatal(err)
		}
		if expectedCanonicalArn != "" && entity.canonicalArn() != expectedCanonicalArn {
			t.Fatalf("expected to canonicalize ARN %q into %q but got %q instead", inputArn, expectedCanonicalArn, entity.canonicalArn())
		}
		if *entity != expectedEntity {
			t.Fatalf("expected to get iamEntity %#v from input ARN %q but instead got %#v", expectedEntity, inputArn, *entity)
		}
	}

	testParser("arn:aws:iam::123456789012:user/UserPath/MyUserName",
		"arn:aws:iam::123456789012:user/MyUserName",
		iamEntity{Partition: "aws", AccountNumber: "123456789012", Type: "user", Path: "UserPath", FriendlyName: "MyUserName"},
	)
	canonicalRoleArn := "arn:aws:iam::123456789012:role/RoleName"
	testParser("arn:aws:sts::123456789012:assumed-role/RoleName/RoleSessionName",
		canonicalRoleArn,
		iamEntity{Partition: "aws", AccountNumber: "123456789012", Type: "assumed-role", FriendlyName: "RoleName", SessionInfo: "RoleSessionName"},
	)
	testParser("arn:aws:iam::123456789012:role/RolePath/RoleName",
		canonicalRoleArn,
		iamEntity{Partition: "aws", AccountNumber: "123456789012", Type: "role", Path: "RolePath", FriendlyName: "RoleName"},
	)
	testParser("arn:aws:iam::123456789012:instance-profile/profilePath/InstanceProfileName",
		"",
		iamEntity{Partition: "aws", AccountNumber: "123456789012", Type: "instance-profile", Path: "profilePath", FriendlyName: "InstanceProfileName"},
	)
}

func TestBackend_validateVaultHeaderValue(t *testing.T) {
	const canaryHeaderValue = "Vault-Server"
	requestUrl, err := url.Parse("https://sts.amazonaws.com/")
	if err != nil {
		t.Fatalf("error parsing test URL: %v", err)
	}
	postHeadersMissing := http.Header{
		"Host":          []string{"Foo"},
		"Authorization": []string{"AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date;x-vault-aws-iam-server-id, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
	}
	postHeadersInvalid := http.Header{
		"Host":            []string{"Foo"},
		iamServerIdHeader: []string{"InvalidValue"},
		"Authorization":   []string{"AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date;x-vault-aws-iam-server-id, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
	}
	postHeadersUnsigned := http.Header{
		"Host":            []string{"Foo"},
		iamServerIdHeader: []string{canaryHeaderValue},
		"Authorization":   []string{"AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
	}
	postHeadersValid := http.Header{
		"Host":            []string{"Foo"},
		iamServerIdHeader: []string{canaryHeaderValue},
		"Authorization":   []string{"AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date;x-vault-aws-iam-server-id, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
	}

	postHeadersSplit := http.Header{
		"Host":            []string{"Foo"},
		iamServerIdHeader: []string{canaryHeaderValue},
		"Authorization":   []string{"AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/iam/aws4_request", "SignedHeaders=content-type;host;x-amz-date;x-vault-aws-iam-server-id, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
	}

	err = validateVaultHeaderValue(postHeadersMissing, requestUrl, canaryHeaderValue)
	if err == nil {
		t.Error("validated POST request with missing Vault header")
	}

	err = validateVaultHeaderValue(postHeadersInvalid, requestUrl, canaryHeaderValue)
	if err == nil {
		t.Error("validated POST request with invalid Vault header value")
	}

	err = validateVaultHeaderValue(postHeadersUnsigned, requestUrl, canaryHeaderValue)
	if err == nil {
		t.Error("validated POST request with unsigned Vault header")
	}

	err = validateVaultHeaderValue(postHeadersValid, requestUrl, canaryHeaderValue)
	if err != nil {
		t.Errorf("did NOT validate valid POST request: %v", err)
	}

	err = validateVaultHeaderValue(postHeadersSplit, requestUrl, canaryHeaderValue)
	if err != nil {
		t.Errorf("did NOT validate valid POST request with split Authorization header: %v", err)
	}
}
