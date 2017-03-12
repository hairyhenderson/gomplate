#!/usr/bin/env bats

load helper

function setup () {
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
  orig_vault_token=$VAULT_TOKEN
}

function teardown () {
  rm -rf $tmpdir
  VAULT_TOKEN=$orig_vault_token
  vault delete secret/foo
  vault auth-disable userpass
  vault auth-disable userpass2
  vault auth-disable approle
  vault auth-disable approle2
  vault auth-disable app-id
  vault auth-disable app-id2
}

@test "Testing token vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  VAULT_TOKEN=$(vault token-create -format=json -policy=readpol -use-limit=1 -ttl=1m | jq -r .auth.client_token)
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing token vault auth using file" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault token-create -format=json -policy=readpol -use-limit=1 -ttl=1m | jq -r .auth.client_token > $tmpdir/token
  VAULT_TOKEN_FILE=$tmpdir/token
  unset VAULT_TOKEN
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable userpass
  vault write auth/userpass/users/dave password=foo ttl=30s policies=readpol
  VAULT_AUTH_USERNAME=dave
  VAULT_AUTH_PASSWORD=foo
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth using files" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable userpass
  vault write auth/userpass/users/dave password=foo ttl=30s policies=readpol
  echo -n "dave" > $tmpdir/username
  echo -n "foo" > $tmpdir/password
  VAULT_AUTH_USERNAME_FILE=$tmpdir/username
  VAULT_AUTH_PASSWORD_FILE=$tmpdir/password
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing userpass vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=userpass2 userpass
  vault write auth/userpass2/users/dave password=foo ttl=30s policies=readpol
  VAULT_AUTH_USERPASS_MOUNT=userpass2
  VAULT_AUTH_USERNAME=dave
  VAULT_AUTH_PASSWORD=foo
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing approle vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable approle
  vault write auth/approle/role/testrole secret_id_ttl=30s token_ttl=35s token_max_ttl=3m secret_id_num_uses=1 policies=readpol
  VAULT_ROLE_ID=$(vault read -field role_id auth/approle/role/testrole/role-id)
  VAULT_SECRET_ID=$(vault write -f -field=secret_id auth/approle/role/testrole/secret-id)
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing approle vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=approle2 approle
  vault write auth/approle2/role/testrole secret_id_ttl=30s token_ttl=35s token_max_ttl=3m secret_id_num_uses=1 policies=readpol
  VAULT_ROLE_ID=$(vault read -field role_id auth/approle2/role/testrole/role-id)
  VAULT_SECRET_ID=$(vault write -f -field=secret_id auth/approle2/role/testrole/secret-id)
  VAULT_AUTH_APPROLE_MOUNT=approle2
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing app-id vault auth" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable app-id
  vault write auth/app-id/map/app-id/testappid value=pol display_name=test_app_id
  vault write auth/app-id/map/user-id/testuserid value=testappid
  VAULT_APP_ID=testappid
  VAULT_USER_ID=testuserid
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Testing app-id vault auth with custom mount" {
  vault write secret/foo value="$BATS_TEST_DESCRIPTION"
  vault auth-enable -path=app-id2 app-id

  vault write auth/app-id2/map/app-id/testappid value=pol display_name=test_app_id
  vault write auth/app-id2/map/user-id/testuserid value=testappid

  VAULT_APP_ID=testappid
  VAULT_USER_ID=testuserid
  VAULT_AUTH_APPID_MOUNT=approle2
  gomplate -d vault=vault:///secret -i '{{(datasource "vault" "foo").value}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

# TODO: test the github auth backend at some point... this needs a github token though, so...
# vault write auth/github/config organization=DockerOttawaMeetup
# vault write auth/github/map/teams/organizers value=pol
