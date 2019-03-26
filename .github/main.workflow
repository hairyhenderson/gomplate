workflow "Build" {
  on = "push"
  resolves = ["gomplate-ci-build"]
}

action "gomplate-ci-build" {
  uses = "docker://hairyhenderson/gomplate-ci-build"
  args = "make lint"
}
