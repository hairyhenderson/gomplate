#!/usr/bin/env bats

load helper

@test "'math.Add'" {
  gomplate -i '{{ math.Add 1 2 3 4 }} {{ add -5 5 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "10 0" ]]
}

@test "'math.Sub'" {
  gomplate -i '{{ math.Sub 10 5 }} {{ sub -5 5 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "5 -10" ]]
}

@test "'math.Mul'" {
  gomplate -i '{{ math.Mul 1 2 3 4 }} {{ mul -5 5 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "24 -25" ]]
}

@test "'math.Div'" {
  gomplate -i '{{ math.Div 5 3 }} {{ div -5 5 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "1 -1" ]]
}

@test "'math.Rem'" {
  gomplate -i '{{ math.Rem 5 3 }} {{ rem 2 2 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "2 0" ]]
}

@test "'math.Pow'" {
  gomplate -i '{{ math.Pow 8 4 }} {{ pow 2 2 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "4096 4" ]]
}