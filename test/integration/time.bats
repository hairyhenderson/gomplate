#!/usr/bin/env bats

load helper

@test "'time.ZoneName'" {
  gomplate -i '{{ time.ZoneName }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == `date +"%Z"` ]]
}

@test "'(time.Now).Format'" {
  gomplate -i '{{ (time.Now).Format "2006-01-02 15 -0700" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == `date +"%Y-%m-%d %H %z"` ]]
}

@test "'(time.Parse).Format'" {
  in=`date -u --date='@1234567890'`
  gomplate -i "{{ (time.Parse \"Mon Jan 02 15:04:05 MST 2006\" \"${in}\").Format \"2006-01-02 15 -0700\" }}"
  [ "$status" -eq 0 ]
  [[ "${output}" == "2009-02-13 23 +0000" ]]
}

@test "'(time.Unix).UTC.Format' int" {
  gomplate -i '{{ (time.Unix 1234567890).UTC.Format "2006-01-02 15 -0700" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "2009-02-13 23 +0000" ]]
}

@test "'(time.Unix).UTC.Format' string" {
  gomplate -i '{{ (time.Unix "1234567890").UTC.Format "2006-01-02 15 -0700" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "2009-02-13 23 +0000" ]]
}
