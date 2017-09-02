#!/usr/bin/env bats

load helper

function setup () {
  unset VAULT_TOKEN
  cat <<EOF | vault policy-write writepol - >& /dev/null
path "*" {
  policy = "write"
}
EOF
  cat <<EOF | vault policy-write readpol - >& /dev/null
path "*" {
  policy = "read"
}
EOF
  tmpdir=$(mktemp -d)
  cp ~/.vault-token ~/.vault-token.bak
  start_meta_svc
  start_aws_svc
}

function teardown () {
  mv ~/.vault-token.bak ~/.vault-token
  stop_meta_svc
  stop_aws_svc
  rm -rf $tmpdir
  unset VAULT_TOKEN
  vault delete secret/foo
  vault auth-disable userpass
  vault auth-disable userpass2
  vault auth-disable approle
  vault auth-disable approle2
  vault auth-disable app-id
  vault auth-disable app-id2
  vault auth-disable aws
  vault policy-delete writepol
  vault policy-delete readpol
  vault unmount ssh
}

@test "Testing token vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  VAULT_TOKEN=$(vault token-create -format=json -policy=readpol -use-limit=1 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing failure with non-existant secret" {
  VAULT_TOKEN=$(vault token-create -format=json -policy=readpol -use-limit=1 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:///secret -i '{{(datasource "vault" "bar").value}}'
  [ "$status" -eq 1 ]
  [[ "${output}" == *"No value found for [bar] from datasource 'vault'" ]]
}

@test "Testing token vault auth using file" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault token-create -format=json -policy=readpol -use-limit=1 -ttl=1m | jq -j .auth.client_token > $tmpdir/token
  unset VAULT_TOKEN
  VAULT_TOKEN_FILE=$tmpdir/token gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable userpass
  vault write auth/userpass/users/dave password=foo ttl=30s policies=readpol
  VAULT_AUTH_USERNAME=dave VAULT_AUTH_PASSWORD=foo gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth using files" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable userpass
  vault write auth/userpass/users/dave password=foo ttl=30s policies=readpol
  echo -n "dave" > $tmpdir/username
  echo -n "foo" > $tmpdir/password
  VAULT_AUTH_USERNAME_FILE=$tmpdir/username VAULT_AUTH_PASSWORD_FILE=$tmpdir/password gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=userpass2 userpass
  vault write auth/userpass2/users/dave password=foo ttl=30s policies=readpol
  VAULT_AUTH_USERPASS_MOUNT=userpass2 VAULT_AUTH_USERNAME=dave VAULT_AUTH_PASSWORD=foo gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing approle vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable approle
  vault write auth/approle/role/testrole secret_id_ttl=30s token_ttl=35s token_max_ttl=3m secret_id_num_uses=1 policies=readpol
  VAULT_ROLE_ID=$(vault read -field role_id auth/approle/role/testrole/role-id)
  VAULT_SECRET_ID=$(vault write -f -field=secret_id auth/approle/role/testrole/secret-id)
  VAULT_ROLE_ID=$VAULT_ROLE_ID VAULT_SECRET_ID=$VAULT_SECRET_ID gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing approle vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=approle2 approle
  vault write auth/approle2/role/testrole secret_id_ttl=30s token_ttl=35s token_max_ttl=3m secret_id_num_uses=1 policies=readpol
  VAULT_ROLE_ID=$(vault read -field role_id auth/approle2/role/testrole/role-id)
  VAULT_SECRET_ID=$(vault write -f -field=secret_id auth/approle2/role/testrole/secret-id)
  VAULT_AUTH_APPROLE_MOUNT=approle2 VAULT_ROLE_ID=$VAULT_ROLE_ID VAULT_SECRET_ID=$VAULT_SECRET_ID gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing app-id vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable app-id
  vault write auth/app-id/map/app-id/testappid value=readpol display_name=test_app_id
  vault write auth/app-id/map/user-id/testuserid value=testappid
  VAULT_APP_ID=testappid VAULT_USER_ID=testuserid gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing app-id vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=app-id2 app-id

  vault write auth/app-id2/map/app-id/testappid value=readpol display_name=test_app_id
  vault write auth/app-id2/map/user-id/testuserid value=testappid

  VAULT_APP_ID=testappid VAULT_USER_ID=testuserid VAULT_AUTH_APP_ID_MOUNT=app-id2 gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing ec2 vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable aws
  vault write auth/aws/config/client secret_key=secret access_key=access endpoint=http://127.0.0.1:8082/ec2 iam_endpoint=http://127.0.0.1:8082/iam sts_endpoint=http://127.0.0.1:8082/sts
  curl -o $tmpdir/certificate -s -f http://127.0.0.1:8081/certificate
  vault write auth/aws/config/certificate/testcert type=pkcs7 aws_public_cert=@$tmpdir/certificate
  vault write auth/aws/role/ami-00000000 auth_type=ec2 bound_ami_id=ami-00000000 policies=readpol
  unset VAULT_TOKEN
  rm ~/.vault-token
  AWS_META_ENDPOINT=http://127.0.0.1:8081 gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing vault auth with dynamic secret" {
  vault mount ssh
  vault write ssh/roles/test key_type=otp default_user=user cidr_list=10.0.0.0/8
  VAULT_TOKEN=$(vault token-create -format=json -policy=writepol -use-limit=2 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:/// -i '{{(datasource "vault" "ssh/creds/test?ip=10.1.2.3&username=user").ip}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "10.1.2.3" ]]
}

@test "Testing vault auth with dynamic secret using prefix" {
  vault mount ssh
  vault write ssh/roles/test key_type=otp default_user=user cidr_list=10.0.0.0/8
  VAULT_TOKEN=$(vault token-create -format=json -policy=writepol -use-limit=2 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:///ssh/creds/test -i '{{(datasource "vault" "?ip=10.1.2.3&username=user").ip}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "10.1.2.3" ]]
}

@test "Testing vault auth with dynamic secret using prefix and options in URL" {
  vault mount ssh
  vault write ssh/roles/test key_type=otp default_user=user cidr_list=10.0.0.0/8
  VAULT_TOKEN=$(vault token-create -format=json -policy=writepol -use-limit=2 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:///ssh/creds/test?ip=10.1.2.3\&username=user -i '{{(datasource "vault").ip}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "10.1.2.3" ]]
}

@test "Testing vault auth with dynamic secret using options in URL and path in template" {
  vault mount ssh
  vault write ssh/roles/test key_type=otp default_user=user cidr_list=10.0.0.0/8
  VAULT_TOKEN=$(vault token-create -format=json -policy=writepol -use-limit=2 -ttl=1m | jq -j .auth.client_token)
  VAULT_TOKEN=$VAULT_TOKEN gomplate -d vault=vault:///?ip=10.1.2.3\&username=user -i '{{(datasource "vault" "ssh/creds/test").ip}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "10.1.2.3" ]]
}

# TODO: test the github auth backend at some point... this needs a github token though, so...
# vault write auth/github/config organization=DockerOttawaMeetup
# vault write auth/github/map/teams/organizers value=pol
